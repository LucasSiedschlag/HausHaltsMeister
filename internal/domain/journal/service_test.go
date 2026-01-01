package journal

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type fakeRepo struct {
	role       string
	accounts   map[string]bool
	categories map[string]string
	referenced bool
}

func (f *fakeRepo) CreateTransaction(ctx context.Context, params CreateTransactionParams) (Transaction, error) {
	return Transaction{}, nil
}

func (f *fakeRepo) ListTransactions(ctx context.Context, params ListTransactionsParams) (ListResult, error) {
	return ListResult{}, nil
}

func (f *fakeRepo) GetTransaction(ctx context.Context, ledgerID, transactionID string) (Transaction, error) {
	return Transaction{}, nil
}

func (f *fakeRepo) UpdateTransaction(ctx context.Context, params UpdateTransactionParams) (Transaction, error) {
	return Transaction{}, nil
}

func (f *fakeRepo) DeleteTransaction(ctx context.Context, ledgerID, transactionID string) error {
	return nil
}

func (f *fakeRepo) HasTransactionReferences(ctx context.Context, ledgerID, transactionID string) (bool, error) {
	return f.referenced, nil
}

func (f *fakeRepo) GetLedgerRole(ctx context.Context, ledgerID, userID string) (string, error) {
	return f.role, nil
}

func (f *fakeRepo) LedgerExists(ctx context.Context, ledgerID string) (bool, error) {
	return true, nil
}

func (f *fakeRepo) GetAccountsByIDs(ctx context.Context, ledgerID string, ids []string) (map[string]bool, error) {
	return f.accounts, nil
}

func (f *fakeRepo) GetCategoriesByIDs(ctx context.Context, ledgerID string, ids []string) (map[string]string, error) {
	return f.categories, nil
}

func TestTransferMustBalance(t *testing.T) {
	repo := &fakeRepo{
		role:       "editor",
		accounts:   map[string]bool{"acc-1": true, "acc-2": true},
		categories: map[string]string{"cat-out": "out", "cat-in": "in"},
	}
	service := NewService(repo)

	_, err := service.CreateTransaction(context.Background(), "user-1", "ledger-1", CreateTransactionParams{
		OccurredAt:  time.Now().UTC(),
		Description: "Transfer",
		Entries: []EntryInput{
			{AccountID: "acc-1", CategoryID: ptr("cat-out"), Kind: "transfer", AmountCents: 100},
			{AccountID: "acc-2", CategoryID: ptr("cat-in"), Kind: "transfer", AmountCents: 50},
		},
	})
	require.Error(t, err)
	require.Equal(t, ErrTransferNotBalanced, err)
}

func TestUpdateTransactionBlockedWhenReferenced(t *testing.T) {
	repo := &fakeRepo{
		role:       "editor",
		accounts:   map[string]bool{"acc-1": true},
		categories: map[string]string{"cat-out": "out"},
		referenced: true,
	}
	service := NewService(repo)

	desc := "Atualizado"
	_, err := service.UpdateTransaction(context.Background(), "user-1", "ledger-1", "tx-1", UpdateTransactionParams{
		Description: &desc,
	})
	require.Error(t, err)
	require.Equal(t, ErrTransactionReferenced, err)
}

func TestDeleteTransactionBlockedWhenReferenced(t *testing.T) {
	repo := &fakeRepo{
		role:       "editor",
		accounts:   map[string]bool{"acc-1": true},
		categories: map[string]string{"cat-out": "out"},
		referenced: true,
	}
	service := NewService(repo)

	err := service.DeleteTransaction(context.Background(), "user-1", "ledger-1", "tx-1")
	require.Error(t, err)
	require.Equal(t, ErrTransactionReferenced, err)
}

func ptr(value string) *string {
	return &value
}
