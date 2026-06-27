// Package app is the composition root: it wires the OIDC mock service,
// observability, rate limiting, and HTTP server into a runnable App.
package app

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"golang.org/x/time/rate"

	adapterhttp "github.com/meigma/go-oidc-mock/internal/adapter/http"
	"github.com/meigma/go-oidc-mock/internal/config"
	"github.com/meigma/go-oidc-mock/internal/observability"
	"github.com/meigma/go-oidc-mock/internal/oidc"
	"github.com/meigma/go-oidc-mock/internal/ratelimit"
)

// rateLimiterIdleTTL is how long an idle per-client bucket is kept before the
// in-process limiter evicts it, bounding memory under churning client keys.
const rateLimiterIdleTTL = 10 * time.Minute

// serviceName is the OpenTelemetry service.name reported by traces. It is a
// default; OTEL_SERVICE_NAME or OTEL_RESOURCE_ATTRIBUTES override it.
const serviceName = "go-oidc-mock"

// App is a fully wired API server ready to Run.
type App struct {
	server        *http.Server
	metricsServer *http.Server
	logger        *slog.Logger
	grace         time.Duration
	// rateLimiter is the in-process rate limiter whose janitor goroutine is
	// stopped during graceful shutdown. It is nil when rate limiting is disabled.
	rateLimiter *ratelimit.InMemory
	// traceShutdown flushes and shuts down the OpenTelemetry tracer provider on
	// graceful shutdown. It is a no-op when tracing is disabled.
	traceShutdown func(context.Context) error
}

// New wires the application from cfg and logger. version is reported in the
// OpenAPI document served by the API.
func New(
	ctx context.Context,
	cfg config.Config,
	logger *slog.Logger,
	version string,
) (*App, error) {
	service, err := oidc.NewService(cfg.IssuerURL)
	if err != nil {
		return nil, fmt.Errorf("init oidc service: %w", err)
	}

	metrics := observability.NewMetrics()
	rateLimiter, installRateLimit, limitRawRoute := buildRateLimiter(cfg, logger)

	// Configure tracing before serving so the global provider is in place when
	// requests start producing spans.
	traceShutdown, err := observability.NewTracerProvider(ctx, observability.TracingConfig{
		Enabled:        cfg.TracingEnabled,
		ServiceName:    serviceName,
		ServiceVersion: version,
	})
	if err != nil {
		return nil, fmt.Errorf("init tracing: %w", err)
	}

	// An empty metrics-addr co-locates /metrics on the API listener; otherwise a
	// dedicated metrics server (below) serves it off the API surface.
	serveMetricsInline := cfg.MetricsAddr == ""
	handler := adapterhttp.NewRouter(adapterhttp.RouterDeps{
		Logger:               logger,
		Metrics:              metrics,
		ServeMetricsEndpoint: serveMetricsInline,
		Version:              version,
		RequestTimeout:       cfg.RequestTimeout,
		CORSAllowedOrigins:   cfg.CORSAllowedOrigins,
		TrustedProxyHeader:   cfg.TrustedProxyHeader,
		Readiness:            nil,
		Register:             nil,
		RawRoutes:            protocolRoutes(service.Handler(), limitRawRoute),
		Tracing:              cfg.TracingEnabled,
		InstallRateLimit:     installRateLimit,
	})

	server := &http.Server{
		Addr:              cfg.Addr,
		Handler:           handler,
		ReadTimeout:       cfg.ReadTimeout,
		ReadHeaderTimeout: cfg.ReadHeaderTimeout,
		WriteTimeout:      cfg.WriteTimeout,
		IdleTimeout:       cfg.IdleTimeout,
	}

	var metricsServer *http.Server
	if !serveMetricsInline {
		metricsServer = &http.Server{
			Addr:              cfg.MetricsAddr,
			Handler:           adapterhttp.NewMetricsHandler(metrics),
			ReadTimeout:       cfg.ReadTimeout,
			ReadHeaderTimeout: cfg.ReadHeaderTimeout,
			WriteTimeout:      cfg.WriteTimeout,
			IdleTimeout:       cfg.IdleTimeout,
		}
	}

	return &App{
		server:        server,
		metricsServer: metricsServer,
		logger:        logger,
		grace:         cfg.ShutdownGrace,
		rateLimiter:   rateLimiter,
		traceShutdown: traceShutdown,
	}, nil
}

// buildRateLimiter constructs the rate limiter and hooks that install it on
// both Huma operations and raw application routes. When rate limiting is
// disabled it returns nil hooks, so NewRouter leaves the API unthrottled.
func buildRateLimiter(
	cfg config.Config,
	logger *slog.Logger,
) (*ratelimit.InMemory, func(huma.API), func(http.Handler) http.Handler) {
	if !cfg.RateLimitEnabled {
		return nil, nil, nil
	}

	limiter := ratelimit.NewInMemory(rate.Limit(cfg.RateLimitRPS), cfg.RateLimitBurst, rateLimiterIdleTTL)
	install := func(api huma.API) {
		ratelimit.NewMiddleware(api, limiter, adapterhttp.ClientIPKeyFunc, logger, true).Install()
	}
	limitRawRoute := adapterhttp.RateLimitHandler(limiter, adapterhttp.ClientIPRequestKeyFunc, logger, true)

	return limiter, install, limitRawRoute
}

// Handler returns the assembled HTTP handler, primarily for functional tests.
func (a *App) Handler() http.Handler {
	return a.server.Handler
}

// OpenAPIYAML builds the API without binding a listener and returns the
// OpenAPI 3.0.3 specification as YAML.
func OpenAPIYAML(version string) ([]byte, error) {
	spec, err := adapterhttp.SpecYAML(version, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("build openapi spec: %w", err)
	}

	return spec, nil
}

func protocolRoutes(
	providerHandler http.Handler,
	limitRawRoute func(http.Handler) http.Handler,
) []adapterhttp.RawRoute {
	if limitRawRoute != nil {
		providerHandler = limitRawRoute(providerHandler)
	}

	return []adapterhttp.RawRoute{
		{Method: http.MethodGet, Path: oidc.DiscoveryPath, Handler: providerHandler},
		{Method: http.MethodGet, Path: oidc.JWKSPath, Handler: providerHandler},
		{Method: http.MethodGet, Path: oidc.AuthorizationPath, Handler: providerHandler},
		{Method: http.MethodPost, Path: oidc.AuthorizationPath, Handler: providerHandler},
		{Method: http.MethodGet, Path: oidc.AuthorizationPath + "/*", Handler: providerHandler},
		{Method: http.MethodPost, Path: oidc.AuthorizationPath + "/*", Handler: providerHandler},
		{Method: http.MethodPost, Path: oidc.TokenPath, Handler: providerHandler},
		{Method: http.MethodGet, Path: oidc.UserInfoPath, Handler: providerHandler},
		{Method: http.MethodPost, Path: oidc.UserInfoPath, Handler: providerHandler},
	}
}
