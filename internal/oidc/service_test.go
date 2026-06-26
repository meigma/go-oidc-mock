package oidc_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/meigma/go-oidc-mock/internal/oidc"
)

func TestServiceNormalizesIssuer(t *testing.T) {
	t.Parallel()

	service, err := oidc.NewService("https://issuer.example.test/tenant/")
	require.NoError(t, err)

	assert.Equal(t, "https://issuer.example.test/tenant", service.Issuer())
}

func TestJWKSStartsWithSigningKey(t *testing.T) {
	t.Parallel()

	service, err := oidc.NewService("http://localhost:8080")
	require.NoError(t, err)

	jwks := service.JWKS()
	require.Len(t, jwks.Keys, 1)
	assert.Equal(t, "go-oidc-mock-signing-key", jwks.Keys[0].KeyID)
	assert.Equal(t, "RS256", jwks.Keys[0].Algorithm)
}

func TestProviderHandlerServesDiscoveryAndJWKS(t *testing.T) {
	t.Parallel()

	service, err := oidc.NewService("https://issuer.example.test/tenant/")
	require.NoError(t, err)

	srv := httptest.NewServer(service.Handler())
	t.Cleanup(srv.Close)

	metadata := getJSON[oidc.ProviderMetadata](t, srv, oidc.DiscoveryPath, http.StatusOK)
	assert.Equal(t, "https://issuer.example.test/tenant", metadata.Issuer)
	assert.Equal(t, "https://issuer.example.test/tenant/oauth2/authorize", metadata.AuthorizationEndpoint)
	assert.Equal(t, "https://issuer.example.test/tenant/oauth2/token", metadata.TokenEndpoint)
	assert.Equal(t, "https://issuer.example.test/tenant/jwks.json", metadata.JWKSURI)
	assert.Equal(t, "https://issuer.example.test/tenant/userinfo", metadata.UserInfoEndpoint)
	assert.Contains(t, metadata.ScopesSupported, "openid")

	jwks := getJSON[oidc.JWKS](t, srv, oidc.JWKSPath, http.StatusOK)
	assert.Len(t, jwks.Keys, 1)
}

func TestNewServiceRejectsInvalidIssuerURL(t *testing.T) {
	t.Parallel()

	tests := []string{
		"",
		"localhost:8080",
		"ftp://issuer.example.test",
		"https://issuer.example.test?x=1",
		"https://issuer.example.test#fragment",
	}

	for _, issuer := range tests {
		t.Run(issuer, func(t *testing.T) {
			t.Parallel()

			_, err := oidc.NewService(issuer)
			require.Error(t, err)
		})
	}
}

func getJSON[T any](t *testing.T, srv *httptest.Server, path string, wantStatus int) T {
	t.Helper()

	resp, err := srv.Client().Get(srv.URL + path)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, wantStatus, resp.StatusCode)

	var out T
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&out))

	return out
}
