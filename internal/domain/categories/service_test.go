package categories

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type fakeRepo struct {
	role      string
	created   CreateCategoryParams
	createErr error
	category  Category
}

func (f *fakeRepo) ListCategories(ctx context.Context, ledgerID string, direction *string, active *bool) ([]Category, error) {
	return nil, nil
}

func (f *fakeRepo) GetCategory(ctx context.Context, ledgerID, categoryID string) (Category, error) {
	return f.category, nil
}

func (f *fakeRepo) CreateCategory(ctx context.Context, params CreateCategoryParams) (Category, error) {
	f.created = params
	return Category{}, f.createErr
}

func (f *fakeRepo) UpdateCategory(ctx context.Context, params UpdateCategoryParams) (Category, error) {
	return Category{}, nil
}

func (f *fakeRepo) DeactivateCategory(ctx context.Context, ledgerID, categoryID string, updatedAt time.Time) error {
	return nil
}

func (f *fakeRepo) LedgerExists(ctx context.Context, ledgerID string) (bool, error) {
	return true, nil
}

func (f *fakeRepo) GetLedgerRole(ctx context.Context, ledgerID, userID string) (string, error) {
	return f.role, nil
}

func (f *fakeRepo) IsCategoryUsed(ctx context.Context, ledgerID, categoryID string) (bool, error) {
	return false, nil
}

func TestCreateCategoryForcesFlags(t *testing.T) {
	repo := &fakeRepo{role: "editor"}
	service := NewService(repo)

	_, err := service.CreateCategory(context.Background(), "user-1", "ledger-1", CreateCategoryParams{
		Name:             "Salario",
		Direction:        "in",
		IsBudgetBase:     true,
		IsBudgetRelevant: true,
		IsActive:         true,
	})
	require.NoError(t, err)
	require.True(t, repo.created.IsBudgetBase)
	require.False(t, repo.created.IsBudgetRelevant)
}
