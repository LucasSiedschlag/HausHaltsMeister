package ucs

import (
	"encoding/json"
	std_http "net/http"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/LucasSiedschlag/HausHaltsMeister/internal/adapters/http"
	"github.com/LucasSiedschlag/HausHaltsMeister/internal/adapters/postgres"
	"github.com/LucasSiedschlag/HausHaltsMeister/internal/domain/category"
	"github.com/LucasSiedschlag/HausHaltsMeister/internal/test/harness"
)

func TestUC04_CreateCategory(t *testing.T) {
	// Setup DB
	db := harness.SetupTestDB(t)
	defer db.Pool.Close()

	// Setup Clean Architecture Stack
	repo := postgres.NewCategoryRepository(db.Pool)
	svc := category.NewService(repo)
	handler := http.NewCategoryHandler(svc)

	// Setup Echo
	e := echo.New()
	http.RegisterCategoryRoutes(e, handler)
	client := harness.NewHTTPClient(e)

	t.Run("Success - Create Category", func(t *testing.T) {
		payload := map[string]interface{}{
			"name":               "New Category Test",
			"direction":          "OUT",
			"is_budget_relevant": true,
		}

		rec := client.Request(t, "POST", "/categories", payload)

		require.Equal(t, std_http.StatusCreated, rec.Code)

		var resp map[string]interface{}
		err := json.Unmarshal(rec.Body.Bytes(), &resp)
		require.NoError(t, err)

		assert.Equal(t, "New Category Test", resp["name"])
		assert.Equal(t, "OUT", resp["direction"])
		assert.NotZero(t, resp["id"])
		assert.True(t, resp["is_active"].(bool))
		assert.True(t, resp["is_budget_relevant"].(bool))
	})

	t.Run("Error - Invalid Direction", func(t *testing.T) {
		payload := map[string]interface{}{
			"name":      "Bad Category",
			"direction": "INVALID",
		}

		rec := client.Request(t, "POST", "/categories", payload)

		require.Equal(t, std_http.StatusBadRequest, rec.Code)
		assert.Contains(t, rec.Body.String(), "invalid direction")
	})

	t.Run("Error - Empty Name", func(t *testing.T) {
		payload := map[string]interface{}{
			"name":      "",
			"direction": "IN",
		}

		rec := client.Request(t, "POST", "/categories", payload)

		require.Equal(t, std_http.StatusBadRequest, rec.Code)
		assert.Contains(t, rec.Body.String(), "category name cannot be empty")
	})
}
