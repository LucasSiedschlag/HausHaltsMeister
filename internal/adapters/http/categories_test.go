package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/LucasSiedschlag/HausHaltsMeister/internal/domain/auth"
	"github.com/LucasSiedschlag/HausHaltsMeister/internal/domain/categories"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
)

type fakeCategoriesService struct {
	updateErr error
}

func (f fakeCategoriesService) ListCategories(ctx context.Context, userID, ledgerID string, direction *string, active *bool) ([]categories.Category, error) {
	return nil, nil
}

func (f fakeCategoriesService) GetCategory(ctx context.Context, userID, ledgerID, categoryID string) (categories.Category, error) {
	return categories.Category{}, nil
}

func (f fakeCategoriesService) CreateCategory(ctx context.Context, userID, ledgerID string, input categories.CreateCategoryParams) (categories.Category, error) {
	return categories.Category{}, nil
}

func (f fakeCategoriesService) UpdateCategory(ctx context.Context, userID, ledgerID, categoryID string, input categories.UpdateCategoryParams) (categories.Category, error) {
	return categories.Category{}, f.updateErr
}

func (f fakeCategoriesService) DeleteCategory(ctx context.Context, userID, ledgerID, categoryID string) error {
	return nil
}

func TestUpdateCategoryAccessDenied(t *testing.T) {
	e := echo.New()
	payload := map[string]string{"name": "Mercado"}
	body, err := json.Marshal(payload)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPatch, "/ledgers/ledger-1/categories/cat-1", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("ledgerId", "categoryId")
	c.SetParamValues("ledger-1", "cat-1")
	c.Set("user", auth.User{ID: "user-1"})

	handler := CategoriesHandler{Service: fakeCategoriesService{updateErr: categories.ErrAccessDenied}}

	err = handler.Update(c)
	require.NoError(t, err)
	require.Equal(t, http.StatusForbidden, rec.Code)
}
