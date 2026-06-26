package http

import (
	"log/slog"
	"math"
	"net/http"
	"strconv"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	chimiddleware "github.com/go-chi/chi/v5/middleware"

	"github.com/meigma/go-oidc-mock/internal/adapter/http/problem"
	"github.com/meigma/go-oidc-mock/internal/ratelimit"
)

// ClientIPKeyFunc keys the rate limiter by the resolved client IP. It reads the
// IP the ClientIP middleware stored on the request context — spoof-safe, since
// that middleware honors --trusted-proxy-header and otherwise trusts only the
// TCP peer — so the limiter and the access log agree on who the client is.
//
// It has the signature of ratelimit.KeyFunc and is the default key for the
// rate-limit middleware. To limit authenticated callers instead, swap in a key
// keying lives here, in the transport, so the limiter core stays
// router-agnostic. It never errors: an unresolved IP yields the empty key, which
// simply shares one bucket rather than failing the request.
func ClientIPKeyFunc(ctx huma.Context) (string, error) {
	r, _ := humachi.Unwrap(ctx)

	return chimiddleware.GetClientIP(r.Context()), nil
}

// ClientIPRequestKeyFunc keys a raw HTTP route by the resolved client IP.
func ClientIPRequestKeyFunc(r *http.Request) (string, error) {
	return chimiddleware.GetClientIP(r.Context()), nil
}

// RateLimitHandler wraps a raw HTTP handler with the same per-client limiter
// used by Huma operations.
func RateLimitHandler(
	limiter ratelimit.Limiter,
	key func(*http.Request) (string, error),
	logger *slog.Logger,
	enabled bool,
) func(http.Handler) http.Handler {
	if logger == nil {
		logger = slog.Default()
	}
	if key == nil {
		key = ClientIPRequestKeyFunc
	}

	return func(next http.Handler) http.Handler {
		if !enabled || limiter == nil {
			return next
		}

		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			limitKey, err := key(r)
			if err != nil {
				logger.WarnContext(r.Context(), "rate-limit key extraction failed; allowing request",
					slog.Any("error", err))
				next.ServeHTTP(w, r)

				return
			}

			decision, err := limiter.Allow(r.Context(), limitKey)
			if err != nil {
				logger.ErrorContext(r.Context(), "rate limiter unavailable; allowing request",
					slog.Any("error", err))
				next.ServeHTTP(w, r)

				return
			}

			if decision.Allowed {
				next.ServeHTTP(w, r)

				return
			}

			if decision.RetryAfter > 0 {
				seconds := int(math.Ceil(decision.RetryAfter.Seconds()))
				w.Header().Set("Retry-After", strconv.Itoa(seconds))
			}

			logger.InfoContext(r.Context(), "rate limit exceeded", slog.String("key", limitKey))
			problem.Write(w, http.StatusTooManyRequests, "rate limit exceeded; retry later")
		})
	}
}
