// Package oidc contains the protocol-facing domain behavior for the mock
// OpenID Connect provider.
package oidc

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
)

const (
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

// Service provides deterministic OIDC provider metadata for the mock server.
type Service struct {
	issuer string
}

// NewService constructs a Service for issuerURL.
func NewService(issuerURL string) (*Service, error) {
	issuer, err := normalizeIssuerURL(issuerURL)
	if err != nil {
		return nil, err
	}

	return &Service{issuer: issuer}, nil
}

// ProviderMetadata returns the mock provider's discovery document.
func (s *Service) ProviderMetadata() ProviderMetadata {
	return ProviderMetadata{
		Issuer:                           s.issuer,
		AuthorizationEndpoint:            s.endpoint(AuthorizationPath),
		TokenEndpoint:                    s.endpoint(TokenPath),
		JWKSURI:                          s.endpoint(JWKSPath),
		UserInfoEndpoint:                 s.endpoint(UserInfoPath),
		ScopesSupported:                  []string{"openid", "profile", "email", "offline_access"},
		ResponseTypesSupported:           []string{"code"},
		GrantTypesSupported:              []string{"authorization_code", "refresh_token"},
		SubjectTypesSupported:            []string{"public"},
		IDTokenSigningAlgValuesSupported: []string{"RS256"},
		TokenEndpointAuthMethodsSupported: []string{
			"client_secret_basic",
			"client_secret_post",
			"none",
		},
		ClaimsSupported: []string{"sub", "name", "email", "email_verified"},
		CodeChallengeMethodsSupported: []string{
			"S256",
		},
	}
}

// JWKS returns the currently configured JSON Web Key Set.
func (s *Service) JWKS() JWKS {
	return JWKS{Keys: []JWK{}}
}

func (s *Service) endpoint(path string) string {
	return s.issuer + path
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
