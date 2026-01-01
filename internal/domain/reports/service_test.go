package reports

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type fakeRepo struct {
	role string
	balances []AccountBalance
	categories []CategorySummary
	cashflow []CashflowEntry
}

func (f *fakeRepo) GetLedgerRole(ctx context.Context, ledgerID, userID string) (string, error) {
	return f.role, nil
}

func (f *fakeRepo) LedgerExists(ctx context.Context, ledgerID string) (bool, error) {
	return true, nil
}

func (f *fakeRepo) ListAccountBalances(ctx context.Context, ledgerID string, cutoff time.Time) ([]AccountBalance, error) {
	return f.balances, nil
}

func (f *fakeRepo) ListCategorySummary(ctx context.Context, ledgerID string, from, to time.Time) ([]CategorySummary, error) {
	return f.categories, nil
}

func (f *fakeRepo) ListCashflow(ctx context.Context, ledgerID string, from, to time.Time) ([]CashflowEntry, error) {
	return f.cashflow, nil
}

func TestBalancesRejectsInvalidMonth(t *testing.T) {
	repo := &fakeRepo{role: "viewer"}
	service := NewService(repo)

	month := time.Date(2026, 2, 2, 0, 0, 0, 0, time.UTC)
	_, err := service.Balances(context.Background(), "user-1", "ledger-1", month)

	var appErr *Error
	require.ErrorAs(t, err, &appErr)
	require.Equal(t, "VALIDATION_ERROR", appErr.Code())
}

func TestCategorySummaryRejectsRange(t *testing.T) {
	repo := &fakeRepo{role: "viewer"}
	service := NewService(repo)

	from := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	_, err := service.CategorySummary(context.Background(), "user-1", "ledger-1", from, to)

	var appErr *Error
	require.ErrorAs(t, err, &appErr)
	require.Equal(t, "VALIDATION_ERROR", appErr.Code())
}

func TestCategorySummarySuccess(t *testing.T) {
	repo := &fakeRepo{
		role: "viewer",
		categories: []CategorySummary{{CategoryID: "cat-1", Name: "Mercado", Direction: "out", TotalCents: 1000}},
	}
	service := NewService(repo)

	from := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 2, 28, 0, 0, 0, 0, time.UTC)
	report, err := service.CategorySummary(context.Background(), "user-1", "ledger-1", from, to)
	require.NoError(t, err)
	require.Len(t, report.Items, 1)
}

func TestCashflowSuccess(t *testing.T) {
	repo := &fakeRepo{
		role: "viewer",
		cashflow: []CashflowEntry{{Month: time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC), TotalInCents: 1000, TotalOutCents: 500, NetCents: 500}},
	}
	service := NewService(repo)

	from := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 2, 28, 0, 0, 0, 0, time.UTC)
	report, err := service.Cashflow(context.Background(), "user-1", "ledger-1", from, to)
	require.NoError(t, err)
	require.Len(t, report.Items, 1)
}
