package ucs

import (
	"context"
	"encoding/json"
	std_http "net/http"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/LucasSiedschlag/HausHaltsMeister/internal/adapters/http"
	"github.com/LucasSiedschlag/HausHaltsMeister/internal/adapters/postgres"
	"github.com/LucasSiedschlag/HausHaltsMeister/internal/domain/cashflow"
	"github.com/LucasSiedschlag/HausHaltsMeister/internal/domain/category"
	"github.com/LucasSiedschlag/HausHaltsMeister/internal/domain/picuinha"
	"github.com/LucasSiedschlag/HausHaltsMeister/internal/test/harness"
)

func TestUC15_ReceiveMoney(t *testing.T) {
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
	_, _ = catRepo.Create(ctx, &category.Category{Name: "Picuinhas", Direction: "IN", IsActive: true})

	// Create Person
	person, _ := picService.CreatePerson(ctx, "Charlie", "")

	t.Run("Success - Receive Money with Auto Flow", func(t *testing.T) {
		payload := map[string]interface{}{
			"person_id":        person.ID,
			"kind":             "MINUS",
			"amount":           50.0,
			"auto_create_flow": true,
		}
		rec := client.Request(t, "POST", "/picuinhas/entries", payload)
		require.Equal(t, std_http.StatusCreated, rec.Code)

		var entryRes map[string]interface{}
		json.Unmarshal(rec.Body.Bytes(), &entryRes)
		assert.Equal(t, 50.0, entryRes["amount"])
		assert.Equal(t, "MINUS", entryRes["kind"])
		assert.NotNil(t, entryRes["cash_flow_id"])
	})
}
