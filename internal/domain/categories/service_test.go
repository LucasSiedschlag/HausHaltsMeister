package categories

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type fakeRepo struct {
	role         string
	created      CreateCategoryParams
	createErr    error
	createErrors map[string]error
	category     Category
	used         bool
}

func (f *fakeRepo) ListCategories(ctx context.Context, ledgerID string, direction *string, active *bool) ([]Category, error) {
	return nil, nil
}

func (f *fakeRepo) GetCategory(ctx context.Context, ledgerID, categoryID string) (Category, error) {
	return f.category, nil
}

func (f *fakeRepo) CreateCategory(ctx context.Context, params CreateCategoryParams) (Category, error) {
	f.created = params
	if f.createErrors != nil {
		if err, ok := f.createErrors[params.Name]; ok {
			return Category{}, err
		}
	}
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
	return f.used, nil
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

func TestCreateCategoryRequiresEditor(t *testing.T) {
	repo := &fakeRepo{role: "viewer"}
	service := NewService(repo)

	_, err := service.CreateCategory(context.Background(), "user-1", "ledger-1", CreateCategoryParams{
		Name:             "Mercado",
		Direction:        "out",
		IsBudgetBase:     false,
		IsBudgetRelevant: true,
		IsActive:         true,
	})
	require.Equal(t, ErrAccessDenied, err)
}

func TestCreateCategorySuccess(t *testing.T) {
	repo := &fakeRepo{role: "editor"}
	service := NewService(repo)

	_, err := service.CreateCategory(context.Background(), "user-1", "ledger-1", CreateCategoryParams{
		Name:             "Mercado",
		Direction:        "out",
		IsBudgetBase:     false,
		IsBudgetRelevant: true,
		IsActive:         true,
	})
	require.NoError(t, err)
	require.Equal(t, "Mercado", repo.created.Name)
	require.Equal(t, "out", repo.created.Direction)
}

func TestSeedCategoriesDefaultPreset(t *testing.T) {
	repo := &fakeRepo{
		role:         "editor",
		createErrors: map[string]error{"Conforto": ErrDuplicateName},
	}
	service := NewService(repo)

	result, err := service.SeedCategories(context.Background(), "user-1", "ledger-1", SeedCategoriesParams{})
	require.NoError(t, err)
	require.Contains(t, result.Created, "Gastos fixos")
	require.Contains(t, result.Created, "Investimentos (Saída)")
	require.Contains(t, result.Created, "Investimentos (Entrada)")
	require.Contains(t, result.Created, "Salário")
	require.Contains(t, result.Skipped, "Conforto")
}

func TestSeedCategoriesCustomRelevance(t *testing.T) {
	repo := &fakeRepo{role: "editor"}
	service := NewService(repo)

	_, err := service.SeedCategories(context.Background(), "user-1", "ledger-1", SeedCategoriesParams{
		Items: []SeedCategoryItem{
			{Name: "Gastos fixos", Direction: "out", IsBudgetRelevant: boolPtr(false)},
		},
	})
	require.NoError(t, err)
	require.False(t, repo.created.IsBudgetRelevant)
}

func TestSeedCategoriesRejectsUnknownNames(t *testing.T) {
	repo := &fakeRepo{role: "editor"}
	service := NewService(repo)

	_, err := service.SeedCategories(context.Background(), "user-1", "ledger-1", SeedCategoriesParams{
		Items: []SeedCategoryItem{{Name: "Categoria invalida", Direction: "out"}},
	})
	require.Error(t, err)
	appErr, ok := err.(*Error)
	require.True(t, ok)
	require.Equal(t, "VALIDATION_ERROR", appErr.Code())
	require.Equal(t, "invalid", appErr.Details()["items"])
}

func TestDeleteCategoryBlockedWhenInUse(t *testing.T) {
	repo := &fakeRepo{role: "editor", used: true}
	service := NewService(repo)

	err := service.DeleteCategory(context.Background(), "user-1", "ledger-1", "cat-1")
	require.Error(t, err)
	require.Equal(t, "VALIDATION_ERROR", err.(*Error).Code())
}
