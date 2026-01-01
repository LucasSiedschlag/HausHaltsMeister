package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/LucasSiedschlag/HausHaltsMeister/internal/domain/auth"
	"github.com/LucasSiedschlag/HausHaltsMeister/internal/domain/reports"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
)

type fakeReportsService struct{}

func (f fakeReportsService) Balances(ctx context.Context, userID, ledgerID string, month time.Time) (reports.BalanceReport, error) {
	return reports.BalanceReport{}, nil
}

func (f fakeReportsService) CategorySummary(ctx context.Context, userID, ledgerID string, from, to time.Time) (reports.CategoryReport, error) {
	return reports.CategoryReport{}, nil
}

func (f fakeReportsService) Cashflow(ctx context.Context, userID, ledgerID string, from, to time.Time) (reports.CashflowReport, error) {
	return reports.CashflowReport{}, nil
}

func TestBalancesRequiresMonth(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/ledgers/ledger-1/reports/balances", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("ledgerId")
	c.SetParamValues("ledger-1")
	c.Set("user", auth.User{ID: "user-1"})

	handler := ReportsHandler{Service: fakeReportsService{}}

	err := handler.Balances(c)
	require.NoError(t, err)
	require.Equal(t, http.StatusUnprocessableEntity, rec.Code)
}

func TestCashflowRequiresRange(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/ledgers/ledger-1/reports/cashflow", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("ledgerId")
	c.SetParamValues("ledger-1")
	c.Set("user", auth.User{ID: "user-1"})

	handler := ReportsHandler{Service: fakeReportsService{}}

	err := handler.Cashflow(c)
	require.NoError(t, err)
	require.Equal(t, http.StatusUnprocessableEntity, rec.Code)
}
