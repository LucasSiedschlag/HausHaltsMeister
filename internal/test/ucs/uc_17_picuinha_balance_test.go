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
	picService := picuinha.NewService(picRepo, nil, nil) // No cf service needed for balance

	// Handlers
	picHandler := http.NewPicuinhaHandler(picService)

	// Echo
	e := echo.New()
	http.RegisterPicuinhaRoutes(e, picHandler)
	client := harness.NewHTTPClient(e)

	ctx := context.Background()
	person, _ := picService.CreatePerson(ctx, "Dave", "")

	// Manually add entries via Service to check balance calculation
	picService.AddDiff(ctx, person.ID, 100.0, "PLUS", nil, nil, picuinha.CardOwnerSelf, false)
	picService.AddDiff(ctx, person.ID, 30.0, "MINUS", nil, nil, picuinha.CardOwnerSelf, false)

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
		assert.Equal(t, 70.0, dave["balance"])
	})
}
