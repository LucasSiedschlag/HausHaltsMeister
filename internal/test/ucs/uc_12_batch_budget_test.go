package ucs

import (
	"context"
	"encoding/json"
	std_http "net/http"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/seuuser/cashflow/internal/adapters/http"
	"github.com/seuuser/cashflow/internal/adapters/postgres"
	"github.com/seuuser/cashflow/internal/domain/budget"
	"github.com/seuuser/cashflow/internal/domain/category"
	"github.com/seuuser/cashflow/internal/test/harness"
)

func TestUC12_BatchBudget(t *testing.T) {
	// Setup DB
	db := harness.SetupTestDB(t)
	defer db.Pool.Close()

	// Repos
	catRepo := postgres.NewCategoryRepository(db.Pool)
	cfRepo := postgres.NewCashFlowRepository(db.Pool)
	budRepo := postgres.NewBudgetRepository(db.Pool)

	// Services
	budService := budget.NewService(budRepo, catRepo, cfRepo)

	// Handlers
	budHandler := http.NewBudgetHandler(budService)

	// Echo
	e := echo.New()
	http.RegisterBudgetRoutes(e, budHandler)
	client := harness.NewHTTPClient(e)

	// Setup Data
	ctx := context.Background()
	foodCat, _ := catRepo.Create(ctx, &category.Category{Name: "Food", Direction: "OUT", IsActive: true})

	t.Run("Apply Budget in Batch (UC-12)", func(t *testing.T) {
		payload := map[string]interface{}{
			"start_month":    "2024-06-01",
			"end_month":      "2024-08-01", // June, July, August
			"category_id":    foodCat.ID,
			"planned_amount": 600.0,
		}
		rec := client.Request(t, "POST", "/budgets/batch", payload)
		require.Equal(t, std_http.StatusOK, rec.Code)

		// Verify June
		recJune := client.Request(t, "GET", "/budgets/2024-06-01/summary", nil)
		var respJune map[string]interface{}
		json.Unmarshal(recJune.Body.Bytes(), &respJune)
		itemsJune := respJune["items"].([]interface{})
		assert.Equal(t, 600.0, itemsJune[0].(map[string]interface{})["planned_amount"])

		// Verify August
		recAug := client.Request(t, "GET", "/budgets/2024-08-01/summary", nil)
		var respAug map[string]interface{}
		json.Unmarshal(recAug.Body.Bytes(), &respAug)
		itemsAug := respAug["items"].([]interface{})
		assert.Equal(t, 600.0, itemsAug[0].(map[string]interface{})["planned_amount"])
	})
}
