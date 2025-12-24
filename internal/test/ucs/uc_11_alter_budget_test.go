package ucs

import (
	"context"
	"encoding/json"
	"fmt"
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

func TestUC11_AlterBudget(t *testing.T) {
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
	otherCat, _ := catRepo.Create(ctx, &category.Category{Name: "Other", Direction: "OUT", IsActive: true})

	monthParam := "2024-04-01"
	path := fmt.Sprintf("/budgets/%s/items", monthParam)

	// 1. Create Initial Item
	payload := map[string]interface{}{
		"category_id":    otherCat.ID,
		"planned_amount": 200.0,
	}
	client.Request(t, "POST", path, payload)

	t.Run("Alter Budget Item (UC-11)", func(t *testing.T) {
		// Change planned amount to 350.0
		updatePayload := map[string]interface{}{
			"category_id":    otherCat.ID,
			"planned_amount": 350.0,
		}
		rec := client.Request(t, "POST", path, updatePayload)
		require.Equal(t, std_http.StatusOK, rec.Code)

		var resp map[string]interface{}
		json.Unmarshal(rec.Body.Bytes(), &resp)
		assert.Equal(t, 350.0, resp["planned_amount"])

		// Verify via Summary
		summaryPath := fmt.Sprintf("/budgets/%s/summary", monthParam)
		sumRec := client.Request(t, "GET", summaryPath, nil)
		var sumResp map[string]interface{}
		json.Unmarshal(sumRec.Body.Bytes(), &sumResp)
		items := sumResp["items"].([]interface{})
		assert.Equal(t, 350.0, items[0].(map[string]interface{})["planned_amount"])
	})
}
