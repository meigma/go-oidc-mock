package app_test

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/meigma/go-oidc-mock/internal/app"
	"github.com/meigma/go-oidc-mock/internal/config"
	"github.com/meigma/go-oidc-mock/internal/observability"
	"github.com/meigma/go-oidc-mock/internal/oidc"
)

func TestAppWiring(t *testing.T) {
	t.Parallel()

	vp := viper.New()
	vp.Set("issuer-url", "https://issuer.example.test")
	cfg := config.Load(vp)
	logger := observability.NewLogger(io.Discard, slog.LevelError, "json")

	application, err := app.New(context.Background(), cfg, logger, "test")
	require.NoError(t, err)

	handler := application.Handler()
	require.NotNil(t, handler)

	healthReq := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	healthRec := httptest.NewRecorder()
	handler.ServeHTTP(healthRec, healthReq)
	assert.Equal(t, http.StatusOK, healthRec.Code)

	discoveryReq := httptest.NewRequest(http.MethodGet, oidc.DiscoveryPath, nil)
	discoveryRec := httptest.NewRecorder()
	handler.ServeHTTP(discoveryRec, discoveryReq)
	require.Equal(t, http.StatusOK, discoveryRec.Code)

	var metadata oidc.ProviderMetadata
	require.NoError(t, json.Unmarshal(discoveryRec.Body.Bytes(), &metadata))
	assert.Equal(t, "https://issuer.example.test", metadata.Issuer)
	assert.Equal(t, "https://issuer.example.test/oauth2/token", metadata.TokenEndpoint)
}

func TestAppWiringRateLimits(t *testing.T) {
	t.Parallel()

	vp := viper.New()
	vp.Set("rate-limit-rps", 1)
	vp.Set("rate-limit-burst", 1)
	cfg := config.Load(vp)
	logger := observability.NewLogger(io.Discard, slog.LevelError, "json")

	application, err := app.New(context.Background(), cfg, logger, "test")
	require.NoError(t, err)
	handler := application.Handler()

	for range 3 {
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/healthz", nil))
		assert.Equal(t, http.StatusOK, rec.Code)
	}

	getDiscovery := func() int {
		req := httptest.NewRequest(http.MethodGet, oidc.DiscoveryPath, nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		return rec.Code
	}

	assert.Equal(t, http.StatusOK, getDiscovery())
	assert.Equal(t, http.StatusTooManyRequests, getDiscovery())
}
