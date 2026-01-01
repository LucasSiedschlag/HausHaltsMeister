package accounts

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type fakeRepo struct {
	role string
}

func (f *fakeRepo) ListAccounts(ctx context.Context, ledgerID string) ([]Account, error) {
	return nil, nil
}

func (f *fakeRepo) GetAccount(ctx context.Context, ledgerID, accountID string) (Account, error) {
	return Account{}, nil
}

func (f *fakeRepo) CreateAccount(ctx context.Context, ledgerID, name, accountType string, isActive bool) (Account, error) {
	return Account{}, nil
}

func (f *fakeRepo) UpdateAccount(ctx context.Context, ledgerID, accountID, name string, isActive bool, updatedAt time.Time) (Account, error) {
	return Account{}, nil
}

func (f *fakeRepo) DeactivateAccount(ctx context.Context, ledgerID, accountID string, updatedAt time.Time) error {
	return nil
}

func (f *fakeRepo) LedgerExists(ctx context.Context, ledgerID string) (bool, error) {
	return true, nil
}

func (f *fakeRepo) GetLedgerRole(ctx context.Context, ledgerID, userID string) (string, error) {
	return f.role, nil
}

func TestCreateAccountRequiresEditor(t *testing.T) {
	repo := &fakeRepo{role: "viewer"}
	service := NewService(repo)

	_, err := service.CreateAccount(context.Background(), "user-1", "ledger-1", "Conta", "cash", true)
	require.Error(t, err)
	require.Equal(t, ErrAccessDenied, err)
}

func TestCreateAccountValidatesType(t *testing.T) {
	repo := &fakeRepo{role: "editor"}
	service := NewService(repo)

	_, err := service.CreateAccount(context.Background(), "user-1", "ledger-1", "Conta", "invalid", true)
	require.Error(t, err)
	require.Equal(t, "VALIDATION_ERROR", err.(*Error).Code())
}
