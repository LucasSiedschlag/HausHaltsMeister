package httpx

import (
	"github.com/LucasSiedschlag/HausHaltsMeister/internal/domain/auth"
	"github.com/labstack/echo/v4"
)

const (
	userContextKey   = "user"
	ledgerContextKey = "ledger"
)

type LedgerContext struct {
	ID   string
	Role string
}

func SetUser(c echo.Context, user auth.User) {
	c.Set(userContextKey, user)
}

func GetUser(c echo.Context) (auth.User, bool) {
	value := c.Get(userContextKey)
	if value == nil {
		return auth.User{}, false
	}
	user, ok := value.(auth.User)
	return user, ok
}

func SetLedger(c echo.Context, ledgerID, role string) {
	c.Set(ledgerContextKey, LedgerContext{ID: ledgerID, Role: role})
}

func GetLedger(c echo.Context) (LedgerContext, bool) {
	value := c.Get(ledgerContextKey)
	if value == nil {
		return LedgerContext{}, false
	}
	ledgerCtx, ok := value.(LedgerContext)
	return ledgerCtx, ok
}
