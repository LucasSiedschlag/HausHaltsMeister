package ucs

import (
	"context"
	"encoding/json"
	"fmt"
	std_http "net/http"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/LucasSiedschlag/HausHaltsMeister/internal/adapters/http"
	"github.com/LucasSiedschlag/HausHaltsMeister/internal/adapters/postgres"
	"github.com/LucasSiedschlag/HausHaltsMeister/internal/domain/budget"
	"github.com/LucasSiedschlag/HausHaltsMeister/internal/domain/cashflow"
	"github.com/LucasSiedschlag/HausHaltsMeister/internal/domain/category"
	"github.com/LucasSiedschlag/HausHaltsMeister/internal/test/harness"
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
	otherCat, _ := catRepo.Create(ctx, &category.Category{Name: "Other", Direction: "OUT", IsActive: true, IsBudgetRelevant: true})
	incomeCat, _ := catRepo.Create(ctx, &category.Category{Name: "Ganho", Direction: "IN", IsActive: true, IsBudgetRelevant: true})

	cfService := cashflow.NewService(cfRepo, catRepo)
	_, err := cfService.CreateCashFlow(ctx, time.Date(2024, 4, 10, 0, 0, 0, 0, time.UTC), incomeCat.ID, "IN", "Salario", 2000.0, false)
	require.NoError(t, err)

	monthParam := "2024-04-01"
	path := fmt.Sprintf("/budgets/%s/items", monthParam)

	// 1. Create Initial Item
	payload := map[string]interface{}{
		"category_id":    otherCat.ID,
		"mode":           budget.ModePercentOfIncome,
		"target_percent": 20.0,
	}
	client.Request(t, "POST", path, payload)

	t.Run("Alter Budget Item (UC-11)", func(t *testing.T) {
		// Change percent to 35
		updatePayload := map[string]interface{}{
			"category_id":    otherCat.ID,
			"mode":           budget.ModePercentOfIncome,
			"target_percent": 35.0,
		}
		rec := client.Request(t, "POST", path, updatePayload)
		require.Equal(t, std_http.StatusOK, rec.Code)

		var resp map[string]interface{}
		json.Unmarshal(rec.Body.Bytes(), &resp)
		assert.Equal(t, 35.0, resp["target_percent"])

		// Verify via Summary
		summaryPath := fmt.Sprintf("/budgets/%s/summary", monthParam)
		sumRec := client.Request(t, "GET", summaryPath, nil)
		var sumResp map[string]interface{}
		json.Unmarshal(sumRec.Body.Bytes(), &sumResp)
		items := sumResp["items"].([]interface{})
		assert.Equal(t, 35.0, items[0].(map[string]interface{})["target_percent"])
	})
}
