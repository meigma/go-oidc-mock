package config_test

import (
	"testing"
	"time"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/meigma/go-oidc-mock/internal/config"
)

func TestLoadDefaults(t *testing.T) {
	t.Parallel()

	cfg := config.Load(viper.New())

	assert.Equal(t, ":8080", cfg.Addr)
	assert.Equal(t, ":9090", cfg.MetricsAddr)
	assert.Equal(t, config.DefaultIssuerURL, cfg.IssuerURL)
	assert.Equal(t, 5*time.Second, cfg.ReadTimeout)
	assert.Equal(t, 5*time.Second, cfg.ReadHeaderTimeout)
	assert.Equal(t, 10*time.Second, cfg.WriteTimeout)
	assert.Equal(t, 120*time.Second, cfg.IdleTimeout)
	assert.Equal(t, 15*time.Second, cfg.RequestTimeout)
	assert.Equal(t, 15*time.Second, cfg.ShutdownGrace)
	assert.Equal(t, "info", cfg.LogLevel)
	assert.Equal(t, "json", cfg.LogFormat)
	assert.Empty(t, cfg.CORSAllowedOrigins)
	assert.Empty(t, cfg.TrustedProxyHeader)
	assert.True(t, cfg.RateLimitEnabled)
	assert.InEpsilon(t, 10.0, cfg.RateLimitRPS, 0.001)
	assert.Equal(t, 20, cfg.RateLimitBurst)
	assert.False(t, cfg.TracingEnabled)
}

func TestLoadFromFlags(t *testing.T) {
	t.Parallel()

	flags := pflag.NewFlagSet("test", pflag.ContinueOnError)
	config.RegisterFlags(flags)

	require.NoError(t, flags.Set("addr", ":9091"))
	require.NoError(t, flags.Set("metrics-addr", ""))
	require.NoError(t, flags.Set("issuer-url", "https://issuer.example.test"))
	require.NoError(t, flags.Set("rate-limit-enabled", "false"))
	require.NoError(t, flags.Set("tracing-enabled", "true"))

	vp := viper.New()
	require.NoError(t, vp.BindPFlags(flags))

	cfg := config.Load(vp)

	assert.Equal(t, ":9091", cfg.Addr)
	assert.Empty(t, cfg.MetricsAddr)
	assert.Equal(t, "https://issuer.example.test", cfg.IssuerURL)
	assert.False(t, cfg.RateLimitEnabled)
	assert.True(t, cfg.TracingEnabled)
}

func TestValidate(t *testing.T) {
	t.Parallel()

	base := config.Load(viper.New())
	require.NoError(t, base.Validate())

	tests := []struct {
		name   string
		mutate func(*config.Config)
	}{
		{
			name: "empty addr",
			mutate: func(cfg *config.Config) {
				cfg.Addr = ""
			},
		},
		{
			name: "matching metrics addr",
			mutate: func(cfg *config.Config) {
				cfg.MetricsAddr = cfg.Addr
			},
		},
		{
			name: "issuer without scheme",
			mutate: func(cfg *config.Config) {
				cfg.IssuerURL = "localhost:8080"
			},
		},
		{
			name: "issuer with query",
			mutate: func(cfg *config.Config) {
				cfg.IssuerURL = "https://issuer.example.test?tenant=a"
			},
		},
		{
			name: "bad log format",
			mutate: func(cfg *config.Config) {
				cfg.LogFormat = "pretty"
			},
		},
		{
			name: "bad rate limit rps",
			mutate: func(cfg *config.Config) {
				cfg.RateLimitRPS = 0
			},
		},
		{
			name: "bad rate limit burst",
			mutate: func(cfg *config.Config) {
				cfg.RateLimitBurst = 0
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cfg := base
			tt.mutate(&cfg)
			require.Error(t, cfg.Validate())
		})
	}
}
