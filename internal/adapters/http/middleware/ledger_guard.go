package middleware

import (
	"context"
	"net/http"

	"github.com/LucasSiedschlag/HausHaltsMeister/internal/adapters/http/httpx"
	"github.com/labstack/echo/v4"
)

type LedgerAccess interface {
	GetMembership(ctx context.Context, userID, ledgerID string) (string, error)
}

func RequireLedgerRole(access LedgerAccess, minRole string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			user, ok := httpx.GetUser(c)
			if !ok {
				return httpx.WriteError(c, http.StatusUnauthorized, "AUTH_INVALID_CREDENTIALS", "Credenciais invalidas", nil)
			}

			ledgerID, err := httpx.RequireUUIDParam(c, "ledgerId")
			if err != nil {
				return err
			}

			role, err := access.GetMembership(c.Request().Context(), user.ID, ledgerID)
			if err != nil {
				return httpx.WriteAppError(c, err)
			}
			if roleRank(role) < roleRank(minRole) {
				return httpx.WriteError(c, http.StatusForbidden, "LEDGER_ACCESS_DENIED", "Acesso negado ao ledger", nil)
			}

			httpx.SetLedger(c, ledgerID, role)
			return next(c)
		}
	}
}

func roleRank(role string) int {
	switch role {
	case "owner":
		return 3
	case "editor":
		return 2
	case "viewer":
		return 1
	default:
		return 0
	}
}
