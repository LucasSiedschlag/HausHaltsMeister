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
	"github.com/seuuser/cashflow/internal/domain/cashflow"
	"github.com/seuuser/cashflow/internal/domain/category"
	"github.com/seuuser/cashflow/internal/domain/installment"
	"github.com/seuuser/cashflow/internal/domain/payment"
	"github.com/seuuser/cashflow/internal/test/harness"
)

func TestUC08_InstallmentsAndInvoice(t *testing.T) {
	// Setup DB
	db := harness.SetupTestDB(t)
	defer db.Pool.Close()

	// Clean Stack
	// Repos
	catRepo := postgres.NewCategoryRepository(db.Pool)
	cfRepo := postgres.NewCashFlowRepository(db.Pool)
	payRepo := postgres.NewPaymentRepository(db.Pool)
	instRepo := postgres.NewInstallmentRepository(db.Pool)

	// Services
	cfService := cashflow.NewService(cfRepo, catRepo)
	payService := payment.NewService(payRepo) // Invoice uses Repo
	instService := installment.NewService(instRepo, cfService, payRepo)

	// Handlers
	cfHandler := http.NewCashFlowHandler(cfService)
	payHandler := http.NewPaymentHandler(payService)
	instHandler := http.NewInstallmentHandler(instService)

	// Echo
	e := echo.New()
	http.RegisterCashFlowRoutes(e, cfHandler)
	http.RegisterPaymentRoutes(e, payHandler)
	http.RegisterInstallmentRoutes(e, instHandler)
	client := harness.NewHTTPClient(e)

	// Seed Category
	ctx := context.Background()
	elecCat, _ := catRepo.Create(ctx, &category.Category{Name: "Electronics", Direction: "OUT", IsActive: true})

	// 1. Create Credit Card
	// UC-08 Requirement: Card exists.
	// We'll create via API or Repo. API is better test of payment domain.
	closingDay := 1
	dueDay := 10
	cardPayload := map[string]interface{}{
		"name":        "Nubank",
		"kind":        "CREDIT_CARD",
		"bank_name":   "Nu Pagamentos",
		"closing_day": closingDay,
		"due_day":     dueDay,
	}
	cardRec := client.Request(t, "POST", "/payment-methods", cardPayload)
	require.Equal(t, std_http.StatusCreated, cardRec.Code)

	var cardRes map[string]interface{}
	err := json.Unmarshal(cardRec.Body.Bytes(), &cardRes)
	require.NoError(t, err)
	cardID := int(cardRes["ID"].(float64))

	// 2. Register Installment Purchase (UC-08)
	// Purchase Date: Jan 20, 2024.
	// Closing Day: 1st.
	// Invoice should be Open (closes Feb 1).
	// Due Date: Feb 10.
	instPayload := map[string]interface{}{
		"description":       "MacBook",
		"total_amount":      1000.0,
		"count":             10,
		"category_id":       elecCat.ID,
		"payment_method_id": cardID,
		"purchase_date":     "2024-01-20",
	}

	instRec := client.Request(t, "POST", "/installments", instPayload)
	require.Equal(t, std_http.StatusCreated, instRec.Code)

	var instRes map[string]interface{}
	err = json.Unmarshal(instRec.Body.Bytes(), &instRes)
	require.NoError(t, err)
	assert.Equal(t, 100.0, instRes["InstallmentAmount"])

	// 3. Verify CashFlows (Future)
	// First payment should be Feb 10.
	listRec := client.Request(t, "GET", "/cashflows?month=2024-02-01", nil)
	require.Equal(t, std_http.StatusOK, listRec.Code)

	var listRes []map[string]interface{}
	err = json.Unmarshal(listRec.Body.Bytes(), &listRes)
	require.NoError(t, err)

	found := false
	for _, cf := range listRes {
		if cf["Title"] == "MacBook (1/10)" {
			found = true
			assert.Equal(t, 100.0, cf["Amount"])
			// Check date Feb 10
			dateStr := cf["Date"].(string)
			parsed, _ := time.Parse(time.RFC3339, dateStr)
			assert.Equal(t, 10, parsed.Day())
		}
	}
	assert.True(t, found, "First installment not found in Feb")

	// 4. View Invoice (UC-09)
	// Get Invoice for Feb (which includes Jan 20 purchase).
	// Actually, if Due Date is Feb 10, it appears in Feb invoice view?
	// UC-09 says "Visualizar Fatura". Usually means "Items due in Month X".
	// Or "Items closing in Month X".
	// Let's assume endpoint returns items due in that month (matching CashFlows).
	// Endpoint: GET /payment-methods/:id/invoice?month=2024-02-01

	// Endpoint: GET /payment-methods/:id/invoice?month=2024-02-01

	path := fmt.Sprintf("/payment-methods/%d/invoice?month=2024-02-01", cardID)
	invRec := client.Request(t, "GET", path, nil)
	require.Equal(t, std_http.StatusOK, invRec.Code)

	var invRes map[string]interface{}
	err = json.Unmarshal(invRec.Body.Bytes(), &invRes)
	require.NoError(t, err)

	// Expect total amount 100.0
	assert.Equal(t, 100.0, invRes["Total"])

	// Expect list items
	items := invRes["Entries"].([]interface{})
	require.NotEmpty(t, items)
	item := items[0].(map[string]interface{})
	assert.Equal(t, "MacBook (1/10)", item["Title"])
}

func jsonNumber(i int) string {
	return string(json.Number(time.Now().Format(""))) // just helper? No, simpler
	// just sprintf
	return ""
}
