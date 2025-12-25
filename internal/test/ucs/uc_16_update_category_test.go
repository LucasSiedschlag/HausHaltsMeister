package ucs

import (
	"context"
	"encoding/json"
	std_http "net/http"
	"strconv"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/LucasSiedschlag/HausHaltsMeister/internal/adapters/http"
	"github.com/LucasSiedschlag/HausHaltsMeister/internal/adapters/postgres"
	"github.com/LucasSiedschlag/HausHaltsMeister/internal/domain/category"
	"github.com/LucasSiedschlag/HausHaltsMeister/internal/test/harness"
)

func TestUC16_UpdateCategory(t *testing.T) {
	db := harness.SetupTestDB(t)
	defer db.Pool.Close()

	repo := postgres.NewCategoryRepository(db.Pool)
	svc := category.NewService(repo)
	handler := http.NewCategoryHandler(svc)

	e := echo.New()
	http.RegisterCategoryRoutes(e, handler)
	client := harness.NewHTTPClient(e)

	created, err := repo.Create(context.Background(), &category.Category{
		Name:             "Initial",
		Direction:        "OUT",
		IsBudgetRelevant: true,
		IsActive:         true,
	})
	require.NoError(t, err)

	t.Run("Success - Update Category", func(t *testing.T) {
		payload := map[string]interface{}{
			"name":               "Updated",
			"direction":          "IN",
			"is_budget_relevant": false,
			"is_active":          true,
		}

		rec := client.Request(t, "PUT", "/categories/"+strconv.Itoa(int(created.ID)), payload)

		require.Equal(t, std_http.StatusOK, rec.Code)

		var resp map[string]interface{}
		err := json.Unmarshal(rec.Body.Bytes(), &resp)
		require.NoError(t, err)

		assert.Equal(t, "Updated", resp["name"])
		assert.Equal(t, "IN", resp["direction"])
		assert.Equal(t, false, resp["is_budget_relevant"])
		assert.Equal(t, true, resp["is_active"])
	})

	t.Run("Error - Invalid Direction", func(t *testing.T) {
		payload := map[string]interface{}{
			"name":               "Bad",
			"direction":          "INVALID",
			"is_budget_relevant": true,
			"is_active":          true,
		}

		rec := client.Request(t, "PUT", "/categories/"+strconv.Itoa(int(created.ID)), payload)

		require.Equal(t, std_http.StatusBadRequest, rec.Code)
		assert.Contains(t, rec.Body.String(), "invalid direction")
	})

	t.Run("Error - Not Found", func(t *testing.T) {
		payload := map[string]interface{}{
			"name":               "Missing",
			"direction":          "OUT",
			"is_budget_relevant": true,
			"is_active":          true,
		}

		rec := client.Request(t, "PUT", "/categories/999999", payload)

		require.Equal(t, std_http.StatusNotFound, rec.Code)
		assert.Contains(t, rec.Body.String(), "category not found")
	})
}
