package ucs

import (
	"context"
	"encoding/json"
	std_http "net/http"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/seuuser/cashflow/internal/adapters/http"
	"github.com/seuuser/cashflow/internal/adapters/postgres"
	"github.com/seuuser/cashflow/internal/domain/cashflow"
	"github.com/seuuser/cashflow/internal/domain/category"
	"github.com/seuuser/cashflow/internal/test/harness"
)

func TestUC06_RegisterFixedExpense(t *testing.T) {
	// Setup DB
	db := harness.SetupTestDB(t)
	defer db.Pool.Close()

	// Setup Clean Architecture Stack
	catRepo := postgres.NewCategoryRepository(db.Pool)
	cfRepo := postgres.NewCashFlowRepository(db.Pool)
	cfService := cashflow.NewService(cfRepo, catRepo)
	cfHandler := http.NewCashFlowHandler(cfService)

	// Setup Echo
	e := echo.New()
	http.RegisterCashFlowRoutes(e, cfHandler)
	client := harness.NewHTTPClient(e)

	// Seed Category
	ctx := context.Background()
	fixedCat, err := catRepo.Create(ctx, &category.Category{Name: "Internet", Direction: "OUT", IsActive: true, IsBudgetRelevant: true})
	require.NoError(t, err)

	t.Run("Create Fixed Expense", func(t *testing.T) {
		payload := map[string]interface{}{
			"date":        time.Now().Format("2006-01-02"),
			"category_id": fixedCat.ID,
			"direction":   "OUT",
			"title":       "Fibra Óptica",
			"amount":      150.00,
			"is_fixed":    true, // UC-06 requirement
		}

		rec := client.Request(t, "POST", "/cashflows", payload)
		require.Equal(t, std_http.StatusCreated, rec.Code)

		var resp map[string]interface{}
		err := json.Unmarshal(rec.Body.Bytes(), &resp)
		require.NoError(t, err)

		assert.Equal(t, "Fibra Óptica", resp["title"])
		assert.Equal(t, 150.00, resp["amount"])
		assert.True(t, resp["is_fixed"].(bool), "Response should indicate is_fixed=true")
	})
}
