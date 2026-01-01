package investments

import (
	"context"
	"testing"
	"time"

	"github.com/LucasSiedschlag/HausHaltsMeister/internal/domain/journal"
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

func (f *fakeRepo) FindAccountByType(ctx context.Context, ledgerID, accountType string) (string, error) {
	if accountType == "cash" {
		return "cash-1", nil
	}
	return "inv-1", nil
}

func (f *fakeRepo) FindCategoryByName(ctx context.Context, ledgerID, name string) (string, error) {
	return "cat-1", nil
}

func (f *fakeRepo) GetCategoryDirection(ctx context.Context, ledgerID, categoryID string) (string, error) {
	return "out", nil
}

func (f *fakeRepo) CreateTransaction(ctx context.Context, params journal.CreateTransactionParams) (journal.Transaction, error) {
	return journal.Transaction{}, nil
}

func (f *fakeRepo) SumByCategory(ctx context.Context, ledgerID, categoryID string, from, to time.Time) (int64, error) {
	return 0, nil
}

func TestContributionRequiresEditor(t *testing.T) {
	repo := &fakeRepo{role: "viewer"}
	service := NewService(repo)

	_, err := service.Contribution(context.Background(), "user-1", "ledger-1", 1000, time.Now().UTC(), nil)
	require.Error(t, err)
	require.Equal(t, ErrAccessDenied, err)
}
