package httpapi_test

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	adapterhttp "github.com/meigma/go-oidc-mock/internal/adapter/http"
	"github.com/meigma/go-oidc-mock/internal/observability"
	"github.com/meigma/go-oidc-mock/internal/oidc"
	"github.com/meigma/go-oidc-mock/internal/oidc/httpapi"
)

func TestFlowPlaceholdersRemainNotImplemented(t *testing.T) {
	t.Parallel()

	service, err := oidc.NewService("https://issuer.example.test")
	require.NoError(t, err)

	handler := adapterhttp.NewRouter(adapterhttp.RouterDeps{
		Logger:         observability.NewLogger(io.Discard, slog.LevelError, "json"),
		Metrics:        observability.NewMetrics(),
		Version:        "test",
		RequestTimeout: 5 * time.Second,
		Register: func(api huma.API) {
			httpapi.Register(api, service)
		},
	})

	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)

	for _, path := range []string{oidc.AuthorizationPath, oidc.UserInfoPath} {
		resp := get(t, srv, path)
		assert.Equal(t, http.StatusNotImplemented, resp.StatusCode)
		require.NoError(t, resp.Body.Close())
	}

	resp, err := srv.Client().Post(srv.URL+oidc.TokenPath, "application/x-www-form-urlencoded", nil)
	require.NoError(t, err)
	assert.Equal(t, http.StatusNotImplemented, resp.StatusCode)
	require.NoError(t, resp.Body.Close())
}

func get(t *testing.T, srv *httptest.Server, path string) *http.Response {
	t.Helper()

	resp, err := srv.Client().Get(srv.URL + path)
	require.NoError(t, err)

	return resp
}
