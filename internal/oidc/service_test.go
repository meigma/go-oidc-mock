package oidc_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/meigma/go-oidc-mock/internal/oidc"
)

func TestProviderMetadataDerivesEndpointURLsFromIssuer(t *testing.T) {
	t.Parallel()

	service, err := oidc.NewService("https://issuer.example.test/tenant/")
	require.NoError(t, err)

	metadata := service.ProviderMetadata()

	assert.Equal(t, "https://issuer.example.test/tenant", metadata.Issuer)
	assert.Equal(t, "https://issuer.example.test/tenant/oauth2/authorize", metadata.AuthorizationEndpoint)
	assert.Equal(t, "https://issuer.example.test/tenant/oauth2/token", metadata.TokenEndpoint)
	assert.Equal(t, "https://issuer.example.test/tenant/jwks.json", metadata.JWKSURI)
	assert.Equal(t, "https://issuer.example.test/tenant/userinfo", metadata.UserInfoEndpoint)
	assert.Contains(t, metadata.ScopesSupported, "openid")
	assert.Equal(t, []string{"code"}, metadata.ResponseTypesSupported)
}

func TestJWKSStartsEmpty(t *testing.T) {
	t.Parallel()

	service, err := oidc.NewService("http://localhost:8080")
	require.NoError(t, err)

	assert.Empty(t, service.JWKS().Keys)
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
