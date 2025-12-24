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

func TestUC07_CopyFixedExpenses(t *testing.T) {
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
	fixedCat, _ := catRepo.Create(ctx, &category.Category{Name: "Fixa", Direction: "OUT", IsActive: true})
	varCat, _ := catRepo.Create(ctx, &category.Category{Name: "Var", Direction: "OUT", IsActive: true})

	// 1. Create Jan Expenses
	// Fixed
	_, err := cfService.CreateCashFlow(ctx, time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC), fixedCat.ID, "OUT", "Internet Jan", 100.0, true)
	require.NoError(t, err)
	// Variable
	_, err = cfService.CreateCashFlow(ctx, time.Date(2024, 1, 20, 0, 0, 0, 0, time.UTC), varCat.ID, "OUT", "Jantar Jan", 200.0, false)
	require.NoError(t, err)

	// 2. Copy to Feb
	payload := map[string]interface{}{
		"from_month": "2024-01-01",
		"to_month":   "2024-02-01",
	}
	rec := client.Request(t, "POST", "/cashflows/copy-fixed", payload)
	require.Equal(t, std_http.StatusOK, rec.Code)

	var res map[string]interface{}
	err = json.Unmarshal(rec.Body.Bytes(), &res)
	require.NoError(t, err)
	assert.Equal(t, 1.0, res["copied_count"], "Should copy exactly 1 fixed expense")

	// 3. Verify Feb
	febListRec := client.Request(t, "GET", "/cashflows?month=2024-02-01", nil)
	require.Equal(t, std_http.StatusOK, febListRec.Code)

	var febList []map[string]interface{}
	err = json.Unmarshal(febListRec.Body.Bytes(), &febList)
	require.NoError(t, err)

	require.Len(t, febList, 1)
	entry := febList[0]
	assert.Equal(t, "Internet Jan", entry["title"])
	assert.Equal(t, 100.0, entry["amount"])
	assert.True(t, entry["is_fixed"].(bool))

	// Verify Date (should be 15th Feb)
	// entry["Date"] comes as string RFC3339 if using standard JSON marshaling
	dateStr := entry["date"].(string)
	parsedDate, _ := time.Parse("2006-01-02", dateStr) // DTO uses Format("2006-01-02")
	assert.Equal(t, 15, parsedDate.Day())
	assert.Equal(t, time.Month(2), parsedDate.Month())
}
