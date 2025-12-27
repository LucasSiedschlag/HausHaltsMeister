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
	"github.com/LucasSiedschlag/HausHaltsMeister/internal/domain/picuinha"
	"github.com/LucasSiedschlag/HausHaltsMeister/internal/test/harness"
)

func TestUC17_PicuinhaBalance(t *testing.T) {
	// Setup DB
	db := harness.SetupTestDB(t)
	defer db.Pool.Close()

	// Repos
	picRepo := postgres.NewPicuinhaRepository(db.Pool)

	// Services
	picService := picuinha.NewService(picRepo)

	// Handlers
	picHandler := http.NewPicuinhaHandler(picService)

	// Echo
	e := echo.New()
	http.RegisterPicuinhaRoutes(e, picHandler)
	client := harness.NewHTTPClient(e)

	ctx := context.Background()
	person, _ := picService.CreatePerson(ctx, "Dave", "")

	_, err := picService.CreateCase(ctx, picuinha.CreateCaseRequest{
		PersonID:    person.ID,
		Title:       "Compra avulsa",
		CaseType:    picuinha.CaseTypeOneOff,
		TotalAmount: 100.0,
		StartDate:   time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
	})
	require.NoError(t, err)

	caseTwo, err := picService.CreateCase(ctx, picuinha.CreateCaseRequest{
		PersonID:    person.ID,
		Title:       "Compra quitada",
		CaseType:    picuinha.CaseTypeOneOff,
		TotalAmount: 30.0,
		StartDate:   time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
	})
	require.NoError(t, err)

	installments, err := picService.ListInstallmentsByCase(ctx, caseTwo.ID)
	require.NoError(t, err)
	require.Len(t, installments, 1)
	_, err = picService.UpdateInstallment(ctx, installments[0].ID, picuinha.UpdateInstallmentRequest{
		IsPaid:      true,
		ExtraAmount: 0,
	})
	require.NoError(t, err)

	t.Run("Verify Balance in List", func(t *testing.T) {
		rec := client.Request(t, "GET", "/picuinhas/persons", nil)
		require.Equal(t, std_http.StatusOK, rec.Code)

		var listRes []map[string]interface{}
		json.Unmarshal(rec.Body.Bytes(), &listRes)

		var dave map[string]interface{}
		for _, p := range listRes {
			if int(p["id"].(float64)) == int(person.ID) {
				dave = p
				break
			}
		}

		require.NotNil(t, dave)
		assert.Equal(t, 100.0, dave["balance"])
	})
}
