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
	"github.com/LucasSiedschlag/HausHaltsMeister/internal/domain/category"
	"github.com/LucasSiedschlag/HausHaltsMeister/internal/domain/ledger"
	"github.com/LucasSiedschlag/HausHaltsMeister/internal/test/harness"
)

func TestUC10_BudgetManagement(t *testing.T) {
	// Setup DB
	db := harness.SetupTestDB(t)
	defer db.Pool.Close()

	// Repos
	catRepo := postgres.NewCategoryRepository(db.Pool)
	budRepo := postgres.NewBudgetRepository(db.Pool)
	ledgerRepo := postgres.NewLedgerRepository(db.Pool)

	// Services
	ledgerService := ledger.NewService(ledgerRepo)
	budService := budget.NewService(budRepo, catRepo, ledgerRepo)

	// Handlers
	budHandler := http.NewBudgetHandler(budService) // Check signature? It expects budget.Service

	// Echo
	e := echo.New()
	http.RegisterBudgetRoutes(e, budHandler)
	client := harness.NewHTTPClient(e)

	// Setup Data
	ctx := context.Background()
	foodCat, _ := catRepo.Create(ctx, &category.Category{Name: "Food", Direction: "OUT", IsActive: true, IsBudgetRelevant: true})
	incomeCat, _ := catRepo.Create(ctx, &category.Category{Name: "Ganho", Direction: "IN", IsActive: true, IsBudgetRelevant: true})

	// 1. Create Ledger Transactions for Mar 2024 to verify "Used" amount logic
	mar1 := time.Date(2024, 3, 5, 0, 0, 0, 0, time.UTC)
	bankAccount, err := ledgerService.CreateAccount(ctx, "Banco", ledger.AccountTypeAsset, "BRL", true)
	require.NoError(t, err)
	incomeAccount, err := ledgerService.CreateAccount(ctx, "Salario", ledger.AccountTypeIncome, "BRL", true)
	require.NoError(t, err)
	expenseAccount, err := ledgerService.CreateAccount(ctx, "Despesa", ledger.AccountTypeExpense, "BRL", true)
	require.NoError(t, err)

	_, err = ledgerService.CreateTransaction(ctx, mar1, "Groceries", "", "", []*ledger.Posting{
		{
			AccountID:  expenseAccount.ID,
			CategoryID: &foodCat.ID,
			Side:       ledger.PostingSideDebit,
			Amount:     150.0,
		},
		{
			AccountID: bankAccount.ID,
			Side:      ledger.PostingSideCredit,
			Amount:    150.0,
		},
	})
	require.NoError(t, err)

	_, err = ledgerService.CreateTransaction(ctx, mar1, "Salario", "", "", []*ledger.Posting{
		{
			AccountID: bankAccount.ID,
			Side:      ledger.PostingSideDebit,
			Amount:    1000.0,
		},
		{
			AccountID:  incomeAccount.ID,
			CategoryID: &incomeCat.ID,
			Side:       ledger.PostingSideCredit,
			Amount:     1000.0,
		},
	})
	require.NoError(t, err)

	monthParam := "2024-03-01"

	// 2. Set Budget Limit (UC-10)
	// Endpoint: POST /budgets/:month/items
	itemPayload := map[string]interface{}{
		"category_id":    foodCat.ID,
		"mode":           budget.ModePercentOfIncome,
		"target_percent": 50.0,
	}
	path := fmt.Sprintf("/budgets/%s/items", monthParam)
	rec := client.Request(t, "POST", path, itemPayload)
	require.Equal(t, std_http.StatusOK, rec.Code)

	var itemRes map[string]interface{}
	err = json.Unmarshal(rec.Body.Bytes(), &itemRes)
	require.NoError(t, err)
	assert.Equal(t, 50.0, itemRes["target_percent"])

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
