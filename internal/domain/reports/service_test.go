package reports

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type fakeRepo struct {
	role string
}

func (f *fakeRepo) GetLedgerRole(ctx context.Context, ledgerID, userID string) (string, error) {
	return f.role, nil
}

func (f *fakeRepo) LedgerExists(ctx context.Context, ledgerID string) (bool, error) {
	return true, nil
}

func (f *fakeRepo) ListAccountBalances(ctx context.Context, ledgerID string, cutoff time.Time) ([]AccountBalance, error) {
	return nil, nil
}

func (f *fakeRepo) ListCategorySummary(ctx context.Context, ledgerID string, from, to time.Time) ([]CategorySummary, error) {
	return nil, nil
}

func (f *fakeRepo) ListCashflow(ctx context.Context, ledgerID string, from, to time.Time) ([]CashflowEntry, error) {
	return nil, nil
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
