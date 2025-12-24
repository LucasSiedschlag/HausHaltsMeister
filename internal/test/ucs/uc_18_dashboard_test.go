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

	"github.com/LucasSiedschlag/HausHaltsMeister/internal/adapters/http"
	"github.com/LucasSiedschlag/HausHaltsMeister/internal/adapters/postgres"
	"github.com/LucasSiedschlag/HausHaltsMeister/internal/domain/cashflow"
	"github.com/LucasSiedschlag/HausHaltsMeister/internal/domain/category"
	"github.com/LucasSiedschlag/HausHaltsMeister/internal/test/harness"
)

func TestUC18_Dashboard(t *testing.T) {
	// Setup DB
	db := harness.SetupTestDB(t)
	defer db.Pool.Close()

	// Clean Stack
	catRepo := postgres.NewCategoryRepository(db.Pool)
	cfRepo := postgres.NewCashFlowRepository(db.Pool)
	cfService := cashflow.NewService(cfRepo, catRepo)
	cfHandler := http.NewCashFlowHandler(cfService)

	e := echo.New()
	http.RegisterCashFlowRoutes(e, cfHandler)
	client := harness.NewHTTPClient(e)

	// Seed
	ctx := context.Background()
	inCat, _ := catRepo.Create(ctx, &category.Category{Name: "Salary", Direction: "IN", IsActive: true})
	outCat, _ := catRepo.Create(ctx, &category.Category{Name: "Food", Direction: "OUT", IsActive: true})

	// Create Data for Jan
	jan1 := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
	_, err := cfService.CreateCashFlow(ctx, jan1, inCat.ID, "IN", "Jan Salary", 5000.0, false)
	require.NoError(t, err)
	_, err = cfService.CreateCashFlow(ctx, jan1, outCat.ID, "OUT", "Jan Food", 1200.0, false)
	require.NoError(t, err)
	_, err = cfService.CreateCashFlow(ctx, jan1, outCat.ID, "OUT", "Jan Snacks", 300.0, false)
	require.NoError(t, err)

	t.Run("UC-18: Monthly Summary", func(t *testing.T) {
		rec := client.Request(t, "GET", "/cashflows/summary?month=2024-01-01", nil)
		require.Equal(t, std_http.StatusOK, rec.Code)

		var res map[string]interface{}
		err = json.Unmarshal(rec.Body.Bytes(), &res)
		require.NoError(t, err)

		assert.Equal(t, 5000.0, res["total_income"])
		assert.Equal(t, 1500.0, res["total_expense"])
		assert.Equal(t, 3500.0, res["balance"])
	})

	t.Run("UC-20: Category Summary", func(t *testing.T) {
		rec := client.Request(t, "GET", "/cashflows/category-summary?month=2024-01-01", nil)
		require.Equal(t, std_http.StatusOK, rec.Code)

		var res []map[string]interface{}
		err = json.Unmarshal(rec.Body.Bytes(), &res)
		require.NoError(t, err)

		// Expect 2 items: Salary (5000), Food (1500)
		require.Len(t, res, 2)

		// Order is by amount DESC
		first := res[0]
		assert.Equal(t, "Salary", first["category_name"])
		assert.Equal(t, 5000.0, first["total_amount"])

		second := res[1]
		assert.Equal(t, "Food", second["category_name"])
		assert.Equal(t, 1500.0, second["total_amount"])
	})
}
