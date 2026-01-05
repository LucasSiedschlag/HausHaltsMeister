package middleware

import (
	"net/http"
	"strconv"

	"github.com/LucasSiedschlag/HausHaltsMeister/internal/adapters/http/httpx"
	"github.com/labstack/echo/v4"
)

func RequireRateLimit(limiter *RateLimiter, keyFn func(echo.Context) string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if limiter == nil || keyFn == nil {
				return next(c)
			}
			key := keyFn(c)
			if key == "" {
				return next(c)
			}

			allowed, retryAfter := limiter.Allow(key)
			if !allowed {
				return httpx.WriteError(c, http.StatusTooManyRequests, "RATE_LIMITED", "Muitas requisicoes", map[string]string{
					"retry_after": strconv.Itoa(int(retryAfter.Seconds())),
				})
			}

			return next(c)
		}
	}
}
