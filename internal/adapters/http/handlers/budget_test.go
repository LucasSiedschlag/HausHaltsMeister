package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/LucasSiedschlag/HausHaltsMeister/internal/domain/auth"
	"github.com/LucasSiedschlag/HausHaltsMeister/internal/domain/budget"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
)

type fakeBudgetService struct {
	monthlyErr error
	periodErr  error
}

func (f fakeBudgetService) GetPlan(ctx context.Context, userID, ledgerID string) (budget.Plan, error) {
	return budget.Plan{}, nil
}

func (f fakeBudgetService) CreatePlan(ctx context.Context, userID, ledgerID, name string) (budget.Plan, error) {
	return budget.Plan{}, nil
}

func (f fakeBudgetService) UpdatePlan(ctx context.Context, userID, ledgerID, name string) (budget.Plan, error) {
	return budget.Plan{}, nil
}

func (f fakeBudgetService) ListVersions(ctx context.Context, userID, ledgerID string, from, to *time.Time) ([]budget.Version, error) {
	return nil, nil
}

func (f fakeBudgetService) GetVersion(ctx context.Context, userID, ledgerID, versionID string) (budget.Version, error) {
	return budget.Version{}, nil
}

func (f fakeBudgetService) CreateVersion(ctx context.Context, userID, ledgerID string, effectiveFrom time.Time, lines []budget.LineInput) (budget.Version, error) {
	return budget.Version{}, nil
}

func (f fakeBudgetService) AddLine(ctx context.Context, userID, ledgerID, versionID string, line budget.LineInput) (budget.Line, error) {
	return budget.Line{}, nil
}

func (f fakeBudgetService) UpdateLine(ctx context.Context, userID, ledgerID, lineID string, percent float64, includeChildren bool) (budget.Line, error) {
	return budget.Line{}, nil
}

func (f fakeBudgetService) DeleteLine(ctx context.Context, userID, ledgerID, lineID string) error {
	return nil
}

func (f fakeBudgetService) MonthlySummary(ctx context.Context, userID, ledgerID string, month time.Time) (budget.MonthlySummary, error) {
	return budget.MonthlySummary{}, f.monthlyErr
}

func (f fakeBudgetService) PeriodSummary(ctx context.Context, userID, ledgerID string, from, to time.Time) (budget.PeriodSummary, error) {
	return budget.PeriodSummary{From: from, To: to}, f.periodErr
}

func TestMonthlyRequiresMonth(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/ledgers/11111111-1111-1111-1111-111111111111/budget/monthly", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("ledgerId")
	c.SetParamValues("11111111-1111-1111-1111-111111111111")
	c.Set("user", auth.User{ID: "user-1"})

	handler := BudgetHandler{Service: fakeBudgetService{}}

	err := handler.Monthly(c)
	require.NoError(t, err)
	require.Equal(t, http.StatusUnprocessableEntity, rec.Code)
}

func TestPeriodRequiresRange(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/ledgers/11111111-1111-1111-1111-111111111111/budget/period", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("ledgerId")
	c.SetParamValues("11111111-1111-1111-1111-111111111111")
	c.Set("user", auth.User{ID: "user-1"})

	handler := BudgetHandler{Service: fakeBudgetService{}}

	err := handler.PeriodSummary(c)
	require.NoError(t, err)
	require.Equal(t, http.StatusUnprocessableEntity, rec.Code)
}
