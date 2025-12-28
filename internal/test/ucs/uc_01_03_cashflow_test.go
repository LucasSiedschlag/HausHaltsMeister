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
	"github.com/LucasSiedschlag/HausHaltsMeister/internal/domain/cashflow"
	"github.com/LucasSiedschlag/HausHaltsMeister/internal/domain/category"
	"github.com/LucasSiedschlag/HausHaltsMeister/internal/domain/installment"
	"github.com/LucasSiedschlag/HausHaltsMeister/internal/domain/ledger"
	"github.com/LucasSiedschlag/HausHaltsMeister/internal/domain/payment"
	"github.com/LucasSiedschlag/HausHaltsMeister/internal/test/harness"
)

func TestUC01_03_CashFlow(t *testing.T) {
	// Setup DB
	db := harness.SetupTestDB(t)
	defer db.Pool.Close()

	// Setup Clean Architecture Stack
	catRepo := postgres.NewCategoryRepository(db.Pool) // For seeding
	cfRepo := postgres.NewCashFlowRepository(db.Pool)
	ledgerRepo := postgres.NewLedgerRepository(db.Pool)
	ledgerService := ledger.NewService(ledgerRepo)
	payRepo := postgres.NewPaymentRepository(db.Pool)
	payService := payment.NewService(payRepo)
	instRepo := postgres.NewInstallmentRepository(db.Pool)
	instService := installment.NewService(instRepo, catRepo, ledgerService, payRepo)
	cfService := cashflow.NewService(cfRepo, catRepo, ledgerService, payRepo, instService)
	cfHandler := http.NewCashFlowHandler(cfService)

	// Setup Echo
	e := echo.New()
	http.RegisterCashFlowRoutes(e, cfHandler)
	client := harness.NewHTTPClient(e)

	// Seed Categories directly via Repo
	ctx := context.Background()
	salaryCat, err := catRepo.Create(ctx, &category.Category{Name: "Salário", Direction: "IN", IsActive: true, IsBudgetRelevant: true})
	require.NoError(t, err)

	rentCat, err := catRepo.Create(ctx, &category.Category{Name: "Aluguel", Direction: "OUT", IsActive: true, IsBudgetRelevant: true})
	require.NoError(t, err)

	paymentMethod, err := payService.CreatePaymentMethod(ctx, "Carteira", payment.KindCash, "", nil, nil, nil)
	require.NoError(t, err)

	t.Run("UC-01: Create Income (IN)", func(t *testing.T) {
		payload := map[string]interface{}{
			"date":        time.Now().Format("2006-01-02"),
			"category_id": salaryCat.ID,
			"payment_method_id": paymentMethod.ID,
			"direction":   "IN",
			"title":       "Salário Mensal",
			"amount":      5000.00,
		}

		rec := client.Request(t, "POST", "/cashflows", payload)
		require.Equal(t, std_http.StatusCreated, rec.Code)

		var resp map[string]interface{}
		err := json.Unmarshal(rec.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, "Salário Mensal", resp["title"])
		assert.Equal(t, 5000.00, resp["amount"])
		assert.Equal(t, "IN", resp["direction"])
	})

	var expenseID int32

	t.Run("UC-03: Create Expense (OUT)", func(t *testing.T) {
		payload := map[string]interface{}{
			"date":        time.Now().Format("2006-01-02"),
			"category_id": rentCat.ID,
			"payment_method_id": paymentMethod.ID,
			"direction":   "OUT",
			"title":       "Aluguel Dezembro",
			"amount":      1200.00,
		}

		rec := client.Request(t, "POST", "/cashflows", payload)
		require.Equal(t, std_http.StatusCreated, rec.Code)

		var resp map[string]interface{}
		err := json.Unmarshal(rec.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, "Aluguel Dezembro", resp["title"])
		assert.Equal(t, 1200.00, resp["amount"])
		if idVal, ok := resp["id"].(float64); ok {
			expenseID = int32(idVal)
		}
	})

	t.Run("UC-03: Reverse Expense", func(t *testing.T) {
		require.NotZero(t, expenseID)
		rec := client.Request(t, "POST", fmt.Sprintf("/cashflows/%d/reverse", expenseID), nil)
		require.Equal(t, std_http.StatusCreated, rec.Code)

		var resp map[string]interface{}
		err := json.Unmarshal(rec.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, float64(expenseID), resp["reversal_of_entry_id"])
	})

	t.Run("Error: Direction Mismatch", func(t *testing.T) {
		// Trying to use "IN" category with "OUT" direction
		payload := map[string]interface{}{
			"date":        time.Now().Format("2006-01-02"),
			"category_id": salaryCat.ID,
			"payment_method_id": paymentMethod.ID,
			"direction":   "OUT", // Wrong
			"title":       "Should Fail",
			"amount":      100.00,
		}

		rec := client.Request(t, "POST", "/cashflows", payload)

		// If handler doesn't wrap errors correctly, this might be 500. We expect 400.
		require.Equal(t, std_http.StatusBadRequest, rec.Code)
	})
}
