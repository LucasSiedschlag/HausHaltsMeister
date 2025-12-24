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
	"github.com/seuuser/cashflow/internal/domain/picuinha"
	"github.com/seuuser/cashflow/internal/test/harness"
)

func TestUC13_Picuinhas(t *testing.T) {
	// Setup DB
	db := harness.SetupTestDB(t)
	defer db.Pool.Close()

	// Repos
	catRepo := postgres.NewCategoryRepository(db.Pool)
	cfRepo := postgres.NewCashFlowRepository(db.Pool)
	picRepo := postgres.NewPicuinhaRepository(db.Pool)

	// Services
	cfService := cashflow.NewService(cfRepo, catRepo)
	picService := picuinha.NewService(picRepo, cfService, catRepo)

	// Handlers
	picHandler := http.NewPicuinhaHandler(picService)

	// Echo
	e := echo.New()
	http.RegisterPicuinhaRoutes(e, picHandler)
	client := harness.NewHTTPClient(e)

	// Setup Categories (Required for Auto Cash Flow)
	ctx := context.Background()
	_, err := catRepo.Create(ctx, &category.Category{Name: "Picuinhas", Direction: "OUT", IsActive: true})
	require.NoError(t, err)
	_, err = catRepo.Create(ctx, &category.Category{Name: "Picuinhas", Direction: "IN", IsActive: true})
	require.NoError(t, err)

	// 1. Create Person (UC-13)
	// Endpoint: POST /picuinhas/persons
	personPayload := map[string]interface{}{
		"name":  "Alice",
		"notes": "Friend",
	}
	resRec := client.Request(t, "POST", "/picuinhas/persons", personPayload)
	require.Equal(t, std_http.StatusCreated, resRec.Code)

	var personRes map[string]interface{}
	err = json.Unmarshal(resRec.Body.Bytes(), &personRes)
	require.NoError(t, err)
	personID := int(personRes["ID"].(float64))

	// 2. Lend 100.0 (UC-14)
	// Kind PLUS, AutoCreateFlow TRUE.
	// Expects CashFlow OUT (Money leaving my pocket)
	lendPayload := map[string]interface{}{
		"person_id":        personID,
		"kind":             "PLUS",
		"amount":           100.0,
		"auto_create_flow": true,
	}
	lendRec := client.Request(t, "POST", "/picuinhas/entries", lendPayload)
	require.Equal(t, std_http.StatusCreated, lendRec.Code)

	// 3. Receive 40.0 (UC-15)
	// Kind MINUS, AutoCreateFlow TRUE.
	// Expects CashFlow IN (Money returning)
	recPayload := map[string]interface{}{
		"person_id":        personID,
		"kind":             "MINUS",
		"amount":           40.0,
		"auto_create_flow": true,
	}
	recRec := client.Request(t, "POST", "/picuinhas/entries", recPayload)
	require.Equal(t, std_http.StatusCreated, recRec.Code)

	// 4. Verify Balance (UC-17)
	// Endpoint: GET /picuinhas/persons (List includes balance)
	listRec := client.Request(t, "GET", "/picuinhas/persons", nil)
	require.Equal(t, std_http.StatusOK, listRec.Code)

	var listRes []map[string]interface{}
	err = json.Unmarshal(listRec.Body.Bytes(), &listRes)
	require.NoError(t, err)

	require.Len(t, listRes, 1)
	alice := listRes[0]
	assert.Equal(t, "Alice", alice["Name"])
	// Balance logic: PLUS adds debt, MINUS reduces debt.
	// 100 lent => Balance 100.
	// 40 received => Balance 60.
	assert.Equal(t, 60.0, alice["Balance"])

	// 5. Verify CashFlows Created
	// We should check DB logic or assume trust if no error.
	// But let's retrieve cashflows via repo strictly to confirm "Picuinhas" category usage.
	flows, err := cfRepo.ListByMonth(ctx, time.Now())
	require.NoError(t, err)
	// Should be 2 flows.
	assert.Len(t, flows, 2)
	// One OUT 100, One IN 40.

	var outFlow, inFlow *cashflow.CashFlow
	for _, f := range flows {
		if f.Direction == "OUT" {
			outFlow = f
		} else {
			inFlow = f
		}
	}
	require.NotNil(t, outFlow, "Expected OUT flow")
	assert.Equal(t, 100.0, outFlow.Amount)
	assert.Contains(t, outFlow.Title, "Picuinha: Alice")

	require.NotNil(t, inFlow, "Expected IN flow")
	assert.Equal(t, 40.0, inFlow.Amount)
	assert.Contains(t, inFlow.Title, "Recebimento: Alice")
}
