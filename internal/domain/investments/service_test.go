package investments

import (
	"context"
	"testing"
	"time"

	"github.com/LucasSiedschlag/HausHaltsMeister/internal/domain/journal"
	"github.com/stretchr/testify/require"
)

type fakeRepo struct {
	role       string
	last       journal.CreateTransactionParams
	categories map[string]string
	sums       map[string]int64
}

func (f *fakeRepo) GetLedgerRole(ctx context.Context, ledgerID, userID string) (string, error) {
	return f.role, nil
}

func (f *fakeRepo) LedgerExists(ctx context.Context, ledgerID string) (bool, error) {
	return true, nil
}

func (f *fakeRepo) FindAccountByType(ctx context.Context, ledgerID, accountType string) (string, error) {
	if accountType == "wallet" {
		return "wallet-1", nil
	}
	if accountType == "current" {
		return "current-1", nil
	}
	return "inv-1", nil
}

func (f *fakeRepo) FindCategoryByName(ctx context.Context, ledgerID, name string) (string, error) {
	if id, ok := f.categories[name]; ok {
		return id, nil
	}
	return "cat-1", nil
}

func (f *fakeRepo) CreateTransaction(ctx context.Context, params journal.CreateTransactionParams) (journal.Transaction, error) {
	f.last = params
	return journal.Transaction{ID: "tx-1"}, nil
}

func (f *fakeRepo) SumByInvestmentAction(ctx context.Context, ledgerID, action string, from, to time.Time) (int64, error) {
	return f.sums[action], nil
}

func TestContributionRequiresEditor(t *testing.T) {
	repo := &fakeRepo{role: "viewer"}
	service := NewService(repo)

	_, err := service.Contribution(context.Background(), "user-1", "ledger-1", 1000, time.Now().UTC(), nil)
	require.Error(t, err)
	require.Equal(t, ErrAccessDenied, err)
}

func TestEarningsCreatesAdjustEntry(t *testing.T) {
	repo := &fakeRepo{role: "editor"}
	service := NewService(repo)

	occurredAt := time.Date(2026, 2, 28, 0, 0, 0, 0, time.UTC)
	_, err := service.Earnings(context.Background(), "user-1", "ledger-1", 500, occurredAt, nil)
	require.NoError(t, err)
	require.Equal(t, "Rendimento investimentos", repo.last.Description)
	require.NotNil(t, repo.last.InvestmentAction)
	require.Equal(t, "earnings", *repo.last.InvestmentAction)
	require.Equal(t, 1, len(repo.last.Entries))
	require.Equal(t, "adjust", repo.last.Entries[0].Kind)
}

func TestRedemptionCreatesTransferEntries(t *testing.T) {
	repo := &fakeRepo{role: "editor"}
	service := NewService(repo)

	occurredAt := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
	_, err := service.Redemption(context.Background(), "user-1", "ledger-1", 1000, occurredAt, nil)
	require.NoError(t, err)
	require.Equal(t, "Resgate investimentos", repo.last.Description)
	require.NotNil(t, repo.last.InvestmentAction)
	require.Equal(t, "redemption", *repo.last.InvestmentAction)
	require.Len(t, repo.last.Entries, 2)
	require.Equal(t, "transfer", repo.last.Entries[0].Kind)
	require.Equal(t, "transfer", repo.last.Entries[1].Kind)
}

func TestSummaryReturnsTotals(t *testing.T) {
	repo := &fakeRepo{
		role: "viewer",
		categories: map[string]string{
			"Investimentos (Entrada)": "cat-invest-in",
			"Investimentos (Saída)":   "cat-invest-out",
		},
		sums: map[string]int64{
			"contribution": 1000,
			"redemption":   2000,
			"earnings":     300,
			"loss":         50,
		},
	}
	service := NewService(repo)

	from := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 1, 31, 0, 0, 0, 0, time.UTC)
	summary, err := service.Summary(context.Background(), "user-1", "ledger-1", from, to)
	require.NoError(t, err)
	require.Equal(t, int64(1000), summary.TotalContributions)
	require.Equal(t, int64(2000), summary.TotalRedemptions)
	require.Equal(t, int64(300), summary.TotalEarnings)
	require.Equal(t, int64(50), summary.TotalLosses)
}
