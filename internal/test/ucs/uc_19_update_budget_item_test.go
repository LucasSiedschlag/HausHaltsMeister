package ucs

import (
	"context"
	"encoding/json"
	"fmt"
	std_http "net/http"
	"strconv"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/LucasSiedschlag/HausHaltsMeister/internal/adapters/http"
	"github.com/LucasSiedschlag/HausHaltsMeister/internal/adapters/postgres"
	"github.com/LucasSiedschlag/HausHaltsMeister/internal/domain/budget"
	"github.com/LucasSiedschlag/HausHaltsMeister/internal/domain/category"
	"github.com/LucasSiedschlag/HausHaltsMeister/internal/test/harness"
)

func TestUC19_UpdateBudgetItem(t *testing.T) {
	db := harness.SetupTestDB(t)
	defer db.Pool.Close()

	catRepo := postgres.NewCategoryRepository(db.Pool)
	cfRepo := postgres.NewCashFlowRepository(db.Pool)
	budRepo := postgres.NewBudgetRepository(db.Pool)

	budService := budget.NewService(budRepo, catRepo, cfRepo)
	budHandler := http.NewBudgetHandler(budService)

	e := echo.New()
	http.RegisterBudgetRoutes(e, budHandler)
	client := harness.NewHTTPClient(e)

	ctx := context.Background()
	cat, _ := catRepo.Create(ctx, &category.Category{Name: "BudgetCat", Direction: "OUT", IsActive: true, IsBudgetRelevant: true})

	monthParam := "2024-05-01"
	createPath := fmt.Sprintf("/budgets/%s/items", monthParam)
	createPayload := map[string]interface{}{
		"category_id":    cat.ID,
		"mode":           budget.ModePercentOfIncome,
		"target_percent": 10.0,
	}
	createRec := client.Request(t, "POST", createPath, createPayload)
	require.Equal(t, std_http.StatusOK, createRec.Code)

	var created map[string]interface{}
	err := json.Unmarshal(createRec.Body.Bytes(), &created)
	require.NoError(t, err)
	itemID := int(created["id"].(float64))

	t.Run("Success - Update Budget Item", func(t *testing.T) {
		updatePayload := map[string]interface{}{
			"mode":           budget.ModePercentOfIncome,
			"target_percent": 25.0,
		}
		updatePath := "/budgets/items/" + strconv.Itoa(itemID)
		rec := client.Request(t, "PUT", updatePath, updatePayload)
		require.Equal(t, std_http.StatusOK, rec.Code)

		var resp map[string]interface{}
		err := json.Unmarshal(rec.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, 25.0, resp["target_percent"])
	})

	t.Run("Error - Invalid Amount", func(t *testing.T) {
		updatePayload := map[string]interface{}{
			"mode":           budget.ModePercentOfIncome,
			"target_percent": 150.0,
		}
		updatePath := "/budgets/items/" + strconv.Itoa(itemID)
		rec := client.Request(t, "PUT", updatePath, updatePayload)
		require.Equal(t, std_http.StatusBadRequest, rec.Code)
		assert.Contains(t, rec.Body.String(), "percent must be between 0 and 100")
	})

	t.Run("Error - Not Found", func(t *testing.T) {
		updatePayload := map[string]interface{}{
			"mode":           budget.ModePercentOfIncome,
			"target_percent": 10.0,
		}
		rec := client.Request(t, "PUT", "/budgets/items/999999", updatePayload)
		require.Equal(t, std_http.StatusNotFound, rec.Code)
		assert.Contains(t, rec.Body.String(), "budget item not found")
	})
}
