// Package oidc contains the protocol-facing domain behavior for the mock
// OpenID Connect provider.
package oidc

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"errors"
	"fmt"
	"net/http"
	"net/url"
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
)

// Service owns the embedded OpenID Provider used by the mock server.
type Service struct {
	issuer   string
	provider *provider.Provider
	jwks     goidc.JSONWebKeySet
}

// NewService constructs a Service for issuerURL.
func NewService(issuerURL string) (*Service, error) {
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
		provider.WithSecretBasicAuthn(),
		provider.WithSecretPostAuthn(),
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
