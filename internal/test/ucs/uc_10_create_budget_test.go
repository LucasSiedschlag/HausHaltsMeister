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

	"github.com/seuuser/cashflow/internal/adapters/http"
	"github.com/seuuser/cashflow/internal/adapters/postgres"
	"github.com/seuuser/cashflow/internal/domain/budget"
	"github.com/seuuser/cashflow/internal/domain/cashflow"
	"github.com/seuuser/cashflow/internal/domain/category"
	"github.com/seuuser/cashflow/internal/test/harness"
)

func TestUC10_BudgetManagement(t *testing.T) {
	// Setup DB
	db := harness.SetupTestDB(t)
	defer db.Pool.Close()

	// Repos
	catRepo := postgres.NewCategoryRepository(db.Pool)
	cfRepo := postgres.NewCashFlowRepository(db.Pool)
	budRepo := postgres.NewBudgetRepository(db.Pool)

	// Services
	cfService := cashflow.NewService(cfRepo, catRepo)
	budService := budget.NewService(budRepo, catRepo, cfRepo)

	// Handlers
	budHandler := http.NewBudgetHandler(budService) // Check signature? It expects budget.Service

	// Echo
	e := echo.New()
	http.RegisterBudgetRoutes(e, budHandler)
	client := harness.NewHTTPClient(e)

	// Setup Data
	ctx := context.Background()
	foodCat, _ := catRepo.Create(ctx, &category.Category{Name: "Food", Direction: "OUT", IsActive: true})

	// 1. Create CashFlow for Mar 2024 to verify "Used" amount logic
	mar1 := time.Date(2024, 3, 5, 0, 0, 0, 0, time.UTC)
	_, err := cfService.CreateCashFlow(ctx, mar1, foodCat.ID, "OUT", "Groceries", 150.0, false)
	require.NoError(t, err)

	monthParam := "2024-03-01"

	// 2. Set Budget Limit (UC-10)
	// Endpoint: POST /budgets/:month/items
	itemPayload := map[string]interface{}{
		"category_id":    foodCat.ID,
		"planned_amount": 500.0,
	}
	path := fmt.Sprintf("/budgets/%s/items", monthParam)
	rec := client.Request(t, "POST", path, itemPayload)
	require.Equal(t, std_http.StatusOK, rec.Code)

	var itemRes map[string]interface{}
	err = json.Unmarshal(rec.Body.Bytes(), &itemRes)
	require.NoError(t, err)
	assert.Equal(t, 500.0, itemRes["planned_amount"])

	// 3. Verify Budget Summary (UC-10 verify)
	// Endpoint: GET /budgets/:month/summary
	pathSummary := fmt.Sprintf("/budgets/%s/summary", monthParam)
	recSum := client.Request(t, "GET", pathSummary, nil)
	require.Equal(t, std_http.StatusOK, recSum.Code)

	var sumRes map[string]interface{}
	err = json.Unmarshal(recSum.Body.Bytes(), &sumRes)
	require.NoError(t, err)

	// Check Items list
	items := sumRes["items"].([]interface{})
	require.Len(t, items, 1)

	item := items[0].(map[string]interface{})
	assert.Equal(t, "Food", item["category_name"])
	assert.Equal(t, 500.0, item["planned_amount"])
	assert.Equal(t, 150.0, item["actual_amount"])
}
