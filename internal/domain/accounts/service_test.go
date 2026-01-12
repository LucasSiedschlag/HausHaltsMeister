package accounts

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type fakeRepo struct {
	role string
	last Account
}

func (f *fakeRepo) ListAccounts(ctx context.Context, ledgerID string) ([]Account, error) {
	return nil, nil
}

func (f *fakeRepo) GetAccount(ctx context.Context, ledgerID, accountID string) (Account, error) {
	return Account{}, nil
}

func (f *fakeRepo) CreateAccount(ctx context.Context, ledgerID, name, accountType, nature string, isActive bool) (Account, error) {
	f.last = Account{ID: "acc-1", LedgerID: ledgerID, Name: name, Type: accountType, Nature: nature, IsActive: isActive}
	return f.last, nil
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

	_, err := service.CreateAccount(context.Background(), "user-1", "ledger-1", "Conta", "current", "asset", true)
	require.Error(t, err)
	require.Equal(t, ErrAccessDenied, err)
}

func TestCreateAccountValidatesType(t *testing.T) {
	repo := &fakeRepo{role: "editor"}
	service := NewService(repo)

	_, err := service.CreateAccount(context.Background(), "user-1", "ledger-1", "Conta", "invalid", "asset", true)
	require.Error(t, err)
	require.Equal(t, "VALIDATION_ERROR", err.(*Error).Code())
}

func TestCreateAccountSuccess(t *testing.T) {
	repo := &fakeRepo{role: "editor"}
	service := NewService(repo)

	account, err := service.CreateAccount(context.Background(), "user-1", "ledger-1", "Conta", "current", "asset", true)
	require.NoError(t, err)
	require.Equal(t, "Conta", account.Name)
	require.Equal(t, "current", account.Type)
}

func TestDeleteAccountSuccess(t *testing.T) {
	repo := &fakeRepo{role: "editor"}
	service := NewService(repo)

	err := service.DeleteAccount(context.Background(), "user-1", "ledger-1", "acc-1")
	require.NoError(t, err)
}
