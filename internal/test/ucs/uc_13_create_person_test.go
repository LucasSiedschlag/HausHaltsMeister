package ucs

import (
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

func TestUC13_CreatePerson(t *testing.T) {
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

	t.Run("Success - Create Person", func(t *testing.T) {
		payload := map[string]interface{}{
			"name":  "Alice",
			"notes": "Friend",
		}
		resRec := client.Request(t, "POST", "/picuinhas/persons", payload)
		require.Equal(t, std_http.StatusCreated, resRec.Code)

		var personRes map[string]interface{}
		err := json.Unmarshal(resRec.Body.Bytes(), &personRes)
		require.NoError(t, err)
		assert.Equal(t, "Alice", personRes["name"])
		assert.NotZero(t, personRes["id"])
	})

	t.Run("Error - Missing Name", func(t *testing.T) {
		payload := map[string]interface{}{
			"name": "",
		}
		resRec := client.Request(t, "POST", "/picuinhas/persons", payload)
		require.Equal(t, std_http.StatusBadRequest, resRec.Code)
		assert.Contains(t, resRec.Body.String(), "name is required")
	})
}
