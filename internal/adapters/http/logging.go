package httpapi

import (
	"log"
	"time"

	"github.com/labstack/echo/v4"
)

func RequestLogger() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()
			err := next(c)
			if err != nil {
				c.Error(err)
			}

			status := c.Response().Status
			if status == 0 {
				status = 200
			}

			requestID := c.Response().Header().Get(echo.HeaderXRequestID)
			if requestID == "" {
				requestID = c.Request().Header.Get(echo.HeaderXRequestID)
			}

			userID := "-"
			if user, ok := GetUser(c); ok {
				userID = user.ID
			}

			ledgerID := c.Param("ledgerId")
			if ledgerID == "" {
				ledgerID = "-"
			}

			latency := time.Since(start)

			log.Printf(
				"request_id=%s method=%s path=%s status=%d latency_ms=%d user_id=%s ledger_id=%s",
				requestID,
				c.Request().Method,
				c.Path(),
				status,
				latency.Milliseconds(),
				userID,
				ledgerID,
			)

			return nil
		}
	}
}
