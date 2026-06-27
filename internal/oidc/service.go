// Package oidc contains the protocol-facing domain behavior for the mock
// OpenID Connect provider.
package oidc

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"errors"
	"fmt"
	"maps"
	"net/http"
	"net/url"
	"slices"
	"strings"

	"github.com/luikyv/go-oidc/pkg/goidc"
	"github.com/luikyv/go-oidc/pkg/provider"
)

const (
	signingKeyBits = 2048

	// DiscoveryPath is the OpenID Provider Configuration endpoint path.
	DiscoveryPath = "/.well-known/openid-configuration"
	// JWKSPath is the JSON Web Key Set endpoint path.
	JWKSPath = "/jwks.json"
	// AuthorizationPath is the OAuth authorization endpoint path.
	AuthorizationPath = "/oauth2/authorize"
	// TokenPath is the OAuth token endpoint path.
	TokenPath = "/oauth2/token" //nolint:gosec // OAuth endpoint path, not a hardcoded credential.
	// UserInfoPath is the OpenID Connect UserInfo endpoint path.
	UserInfoPath = "/userinfo"

	// DefaultPublicClientID is the built-in public client identifier.
	DefaultPublicClientID = "go-oidc-mock-public"
	// DefaultConfidentialClientID is the built-in confidential client identifier.
	DefaultConfidentialClientID = "go-oidc-mock-confidential"
	// DefaultConfidentialClientSecret is the built-in confidential client secret.
	DefaultConfidentialClientSecret = "go-oidc-mock-secret" //nolint:gosec // Intentional mock client secret.
	// DefaultRedirectURI is the built-in localhost redirect URI.
	DefaultRedirectURI = "http://localhost:3000/callback"
	// DefaultLoopbackRedirectURI is the built-in loopback redirect URI.
	DefaultLoopbackRedirectURI = "http://127.0.0.1:3000/callback"

	// DefaultSubject is the fixed fallback subject used when no profiles are mounted.
	DefaultSubject = "go-oidc-mock-user"
	// DefaultName is the fixed fallback name claim used when no profiles are mounted.
	DefaultName = "Mock User"
	// DefaultEmail is the fixed fallback email claim used when no profiles are mounted.
	DefaultEmail = "user@example.test"
)

const (
	idClaimsStoreKey   = "go_oidc_mock_id_claims"
	infoClaimsStoreKey = "go_oidc_mock_info_claims"
)

// Service owns the embedded OpenID Provider used by the mock server.
type Service struct {
	issuer   string
	provider *provider.Provider
	jwks     goidc.JSONWebKeySet
}

type serviceConfig struct {
	clients  []*goidc.Client
	profiles []Profile
}

// ServiceOption customizes Service construction for internal tests and future wiring.
type ServiceOption func(*serviceConfig)

// WithClients replaces the built-in static clients used by the provider.
func WithClients(clients ...*goidc.Client) ServiceOption {
	return func(cfg *serviceConfig) {
		cfg.clients = cloneClients(clients)
	}
}

// WithProfiles replaces the startup profiles used by auto-approval.
func WithProfiles(profiles ...Profile) ServiceOption {
	return func(cfg *serviceConfig) {
		cfg.profiles = cloneProfiles(profiles)
	}
}

// NewService constructs a Service for issuerURL.
func NewService(issuerURL string) (*Service, error) {
	return NewServiceWithOptions(issuerURL)
}

// NewServiceWithOptions constructs a Service for issuerURL with internal options.
func NewServiceWithOptions(issuerURL string, opts ...ServiceOption) (*Service, error) {
	cfg := serviceConfig{clients: defaultClients()}
	for _, opt := range opts {
		opt(&cfg)
	}
	if len(cfg.clients) == 0 {
		return nil, errors.New("at least one static client is required")
	}
	profiles, err := NormalizeProfiles(cfg.profiles)
	if err != nil {
		return nil, fmt.Errorf("validate profiles: %w", err)
	}

	issuer, err := normalizeIssuerURL(issuerURL)
	if err != nil {
		return nil, err
	}

	jwks, err := newSigningJWKS()
	if err != nil {
		return nil, fmt.Errorf("generate signing jwks: %w", err)
	}

	op, err := provider.New(
		issuer,
		nil,
		func(context.Context) (goidc.JSONWebKeySet, error) {
			return jwks, nil
		},
		provider.WithJWKSEndpoint(JWKSPath),
		provider.WithAuthorizeEndpoint(AuthorizationPath),
		provider.WithTokenEndpoint(TokenPath),
		provider.WithUserInfoEndpoint(UserInfoPath),
		provider.WithScopes(goidc.ScopeProfile, goidc.ScopeEmail, goidc.ScopeOfflineAccess),
		provider.WithClaims(goidc.ClaimSubject, goidc.ClaimName, goidc.ClaimEmail, goidc.ClaimEmailVerified),
		provider.WithIDTokenSignatureAlgs(goidc.RS256),
		provider.WithAuthCodeGrant(nil, goidc.ResponseTypeCode),
		provider.WithPKCE(goidc.CodeChallengeMethodSHA256),
		provider.WithNoneAuthn(),
		provider.WithSecretBasicAuthn(),
		provider.WithSecretPostAuthn(),
		provider.WithStaticClients(cfg.clients[0], cfg.clients[1:]...),
		provider.WithPolicies(newAuthorizationPolicy(profiles)),
		provider.WithIDTokenClaims(idTokenClaimsFromStore(idClaimsStoreKey)),
		provider.WithUserInfoClaims(userInfoClaimsFromStore(infoClaimsStoreKey)),
		provider.WithJTIConsumer(func(context.Context, string) error {
			return nil
		}),
	)
	if err != nil {
		return nil, fmt.Errorf("init embedded oidc provider: %w", err)
	}

	return &Service{
		issuer:   issuer,
		provider: op,
		jwks:     jwks,
	}, nil
}

// Issuer returns the normalized issuer URL for this mock provider.
func (s *Service) Issuer() string {
	return s.issuer
}

// Handler returns the embedded provider's HTTP handler.
func (s *Service) Handler() http.Handler {
	return s.provider.Handler()
}

// JWKS returns the public signing keys currently advertised by the provider.
func (s *Service) JWKS() goidc.JSONWebKeySet {
	return s.jwks.Public()
}

func defaultClients() []*goidc.Client {
	redirectURIs := []string{DefaultRedirectURI, DefaultLoopbackRedirectURI}
	scopeIDs := strings.Join([]string{
		goidc.ScopeOpenID.ID,
		goidc.ScopeProfile.ID,
		goidc.ScopeEmail.ID,
	}, " ")
	grantTypes := []goidc.GrantType{goidc.GrantAuthorizationCode}
	responseTypes := []goidc.ResponseType{goidc.ResponseTypeCode}

	return []*goidc.Client{
		{
			ID: DefaultPublicClientID,
			ClientMeta: goidc.ClientMeta{
				Name:             "go-oidc-mock public client",
				RedirectURIs:     slices.Clone(redirectURIs),
				GrantTypes:       slices.Clone(grantTypes),
				ResponseTypes:    slices.Clone(responseTypes),
				ScopeIDs:         scopeIDs,
				TokenAuthnMethod: goidc.AuthnMethodNone,
			},
		},
		{
			ID:     DefaultConfidentialClientID,
			Secret: DefaultConfidentialClientSecret,
			ClientMeta: goidc.ClientMeta{
				Name:             "go-oidc-mock confidential client",
				RedirectURIs:     slices.Clone(redirectURIs),
				GrantTypes:       slices.Clone(grantTypes),
				ResponseTypes:    slices.Clone(responseTypes),
				ScopeIDs:         scopeIDs,
				TokenAuthnMethod: goidc.AuthnMethodSecretBasic,
			},
		},
	}
}

func cloneClients(clients []*goidc.Client) []*goidc.Client {
	cloned := make([]*goidc.Client, 0, len(clients))
	for _, client := range clients {
		if client == nil {
			continue
		}
		copyClient := *client
		copyClient.RedirectURIs = slices.Clone(client.RedirectURIs)
		copyClient.RequestURIs = slices.Clone(client.RequestURIs)
		copyClient.GrantTypes = slices.Clone(client.GrantTypes)
		copyClient.ResponseTypes = slices.Clone(client.ResponseTypes)
		copyClient.Contacts = slices.Clone(client.Contacts)
		copyClient.AuthDetailTypes = slices.Clone(client.AuthDetailTypes)
		copyClient.PostLogoutRedirectURIs = slices.Clone(client.PostLogoutRedirectURIs)
		cloned = append(cloned, &copyClient)
	}

	return cloned
}

func cloneProfiles(profiles []Profile) []Profile {
	cloned := make([]Profile, 0, len(profiles))
	for _, profile := range profiles {
		cloned = append(cloned, profile.clone())
	}

	return cloned
}

func idTokenClaimsFromStore(key string) goidc.IDTokenClaimsFunc {
	return func(_ context.Context, grant *goidc.Grant) map[string]any {
		return claimsFromStore(grant, key)
	}
}

func userInfoClaimsFromStore(key string) goidc.UserInfoClaimsFunc {
	return func(_ context.Context, grant *goidc.Grant) map[string]any {
		return claimsFromStore(grant, key)
	}
}

func claimsFromStore(grant *goidc.Grant, key string) map[string]any {
	if grant == nil || grant.Store == nil {
		return nil
	}

	claims, ok := grant.Store[key].(map[string]any)
	if !ok {
		return nil
	}

	out := make(map[string]any, len(claims))
	maps.Copy(out, claims)

	return out
}

func newSigningJWKS() (goidc.JSONWebKeySet, error) {
	key, err := rsa.GenerateKey(rand.Reader, signingKeyBits)
	if err != nil {
		return goidc.JSONWebKeySet{}, err
	}

	return goidc.JSONWebKeySet{
		Keys: []goidc.JSONWebKey{
			{
				KeyID:     "go-oidc-mock-signing-key",
				Key:       key,
				Use:       "sig",
				Algorithm: string(goidc.RS256),
			},
		},
	}, nil
}

func normalizeIssuerURL(raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", errors.New("issuer URL must not be empty")
	}

	parsed, err := url.Parse(trimmed)
	if err != nil {
		return "", fmt.Errorf("parse issuer URL: %w", err)
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return "", errors.New("issuer URL must use http or https")
	}
	if parsed.Host == "" {
		return "", errors.New("issuer URL must include a host")
	}
	if parsed.RawQuery != "" || parsed.Fragment != "" {
		return "", errors.New("issuer URL must not include a query or fragment")
	}

	parsed.Path = strings.TrimRight(parsed.Path, "/")
	parsed.RawQuery = ""
	parsed.Fragment = ""

	return parsed.String(), nil
}
