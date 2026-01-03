package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/LucasSiedschlag/HausHaltsMeister/internal/adapters/http/httpx"
	"github.com/LucasSiedschlag/HausHaltsMeister/internal/domain/auth"
	"github.com/LucasSiedschlag/HausHaltsMeister/internal/domain/ledger"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
)

type fakeLedgerAccess struct {
	role string
	err  error
}

func (f fakeLedgerAccess) GetMembership(ctx context.Context, userID, ledgerID string) (string, error) {
	if f.err != nil {
		return "", f.err
	}
	return f.role, nil
}

func TestLedgerGuardAllowsViewer(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/ledgers/11111111-1111-1111-1111-111111111111/accounts", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("ledgerId")
	c.SetParamValues("11111111-1111-1111-1111-111111111111")
	httpx.SetUser(c, auth.User{ID: "user-1"})

	guard := RequireLedgerRole(fakeLedgerAccess{role: "viewer"}, "viewer")
	called := false
	err := guard(func(c echo.Context) error {
		called = true
		ledgerCtx, ok := httpx.GetLedger(c)
		require.True(t, ok)
		require.Equal(t, "11111111-1111-1111-1111-111111111111", ledgerCtx.ID)
		require.Equal(t, "viewer", ledgerCtx.Role)
		return c.NoContent(http.StatusNoContent)
	})(c)

	require.NoError(t, err)
	require.True(t, called)
	require.Equal(t, http.StatusNoContent, rec.Code)
}

func TestLedgerGuardDeniesInsufficientRole(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/ledgers/11111111-1111-1111-1111-111111111111/accounts", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("ledgerId")
	c.SetParamValues("11111111-1111-1111-1111-111111111111")
	httpx.SetUser(c, auth.User{ID: "user-1"})

	guard := RequireLedgerRole(fakeLedgerAccess{role: "viewer"}, "editor")
	err := guard(func(c echo.Context) error {
		return c.NoContent(http.StatusNoContent)
	})(c)

	require.NoError(t, err)
	require.Equal(t, http.StatusForbidden, rec.Code)
}

func TestLedgerGuardHandlesAccessErrors(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/ledgers/22222222-2222-2222-2222-222222222222/accounts", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("ledgerId")
	c.SetParamValues("22222222-2222-2222-2222-222222222222")
	httpx.SetUser(c, auth.User{ID: "user-1"})

	guard := RequireLedgerRole(fakeLedgerAccess{err: ledger.ErrAccessDenied}, "viewer")
	err := guard(func(c echo.Context) error {
		return c.NoContent(http.StatusNoContent)
	})(c)

	require.NoError(t, err)
	require.Equal(t, http.StatusForbidden, rec.Code)

	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	c.SetParamNames("ledgerId")
	c.SetParamValues("22222222-2222-2222-2222-222222222222")
	httpx.SetUser(c, auth.User{ID: "user-1"})

	guard = RequireLedgerRole(fakeLedgerAccess{err: ledger.ErrLedgerNotFound}, "viewer")
	err = guard(func(c echo.Context) error {
		return c.NoContent(http.StatusNoContent)
	})(c)

	require.NoError(t, err)
	require.Equal(t, http.StatusNotFound, rec.Code)
}
