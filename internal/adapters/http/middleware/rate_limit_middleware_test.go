package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
)

func TestRequireRateLimitBlocksAfterLimit(t *testing.T) {
	e := echo.New()
	limiter := NewRateLimiter(1, time.Minute)
	calls := 0

	mw := RequireRateLimit(limiter, func(c echo.Context) string { return "reports:user-1:ledger-1" })
	handler := mw(func(c echo.Context) error {
		calls++
		return c.NoContent(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodGet, "/ledgers/ledger-1/reports/balances", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler(c)
	require.NoError(t, err)
	require.Equal(t, http.StatusNoContent, rec.Code)
	require.Equal(t, 1, calls)

	rec2 := httptest.NewRecorder()
	c2 := e.NewContext(req, rec2)

	err = handler(c2)
	require.NoError(t, err)
	require.Equal(t, http.StatusTooManyRequests, rec2.Code)
	require.Contains(t, rec2.Body.String(), "RATE_LIMITED")
	require.Equal(t, 1, calls)
}
