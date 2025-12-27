package ucs

import (
	"encoding/json"
	"fmt"
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

func TestUC20_PicuinhaCrud(t *testing.T) {
	db := harness.SetupTestDB(t)
	defer db.Pool.Close()

	picRepo := postgres.NewPicuinhaRepository(db.Pool)
	picService := picuinha.NewService(picRepo)
	picHandler := http.NewPicuinhaHandler(picService)

	e := echo.New()
	http.RegisterPicuinhaRoutes(e, picHandler)
	client := harness.NewHTTPClient(e)

	t.Run("Update person", func(t *testing.T) {
		payload := map[string]interface{}{
			"name":  "Ana",
			"notes": "Vizinha",
		}
		createRec := client.Request(t, "POST", "/picuinhas/persons", payload)
		require.Equal(t, std_http.StatusCreated, createRec.Code)

		var created map[string]interface{}
		require.NoError(t, json.Unmarshal(createRec.Body.Bytes(), &created))

		updatePayload := map[string]interface{}{
			"name":  "Ana Maria",
			"notes": "Colega do trabalho",
		}
		updateRec := client.Request(t, "PUT", "/picuinhas/persons/"+toID(created["id"]), updatePayload)
		require.Equal(t, std_http.StatusOK, updateRec.Code)

		var updated map[string]interface{}
		require.NoError(t, json.Unmarshal(updateRec.Body.Bytes(), &updated))
		assert.Equal(t, "Ana Maria", updated["name"])
	})

	t.Run("Cases list/update/delete", func(t *testing.T) {
		personPayload := map[string]interface{}{
			"name":  "Carlos",
			"notes": "",
		}
		personRec := client.Request(t, "POST", "/picuinhas/persons", personPayload)
		require.Equal(t, std_http.StatusCreated, personRec.Code)

		var person map[string]interface{}
		require.NoError(t, json.Unmarshal(personRec.Body.Bytes(), &person))

		casePayload := map[string]interface{}{
			"person_id":    person["id"],
			"title":        "Compra à vista",
			"case_type":    "ONE_OFF",
			"total_amount": 120.0,
			"start_date":   "2025-02-01",
		}
		caseRec := client.Request(t, "POST", "/picuinhas/cases", casePayload)
		require.Equal(t, std_http.StatusCreated, caseRec.Code)

		listRec := client.Request(t, "GET", "/picuinhas/cases?person_id="+toID(person["id"]), nil)
		require.Equal(t, std_http.StatusOK, listRec.Code)

		var cases []map[string]interface{}
		require.NoError(t, json.Unmarshal(listRec.Body.Bytes(), &cases))
		require.NotEmpty(t, cases)

		caseID := toID(cases[0]["id"])
		updatePayload := map[string]interface{}{
			"person_id":    person["id"],
			"title":        "Compra ajustada",
			"case_type":    "ONE_OFF",
			"total_amount": 50.0,
			"start_date":   "2025-02-01",
		}
		updateRec := client.Request(t, "PUT", "/picuinhas/cases/"+caseID, updatePayload)
		require.Equal(t, std_http.StatusOK, updateRec.Code)

		deleteRec := client.Request(t, "DELETE", "/picuinhas/cases/"+caseID, nil)
		require.Equal(t, std_http.StatusOK, deleteRec.Code)
	})

	t.Run("Delete person blocked when cases exist", func(t *testing.T) {
		personPayload := map[string]interface{}{
			"name":  "Marina",
			"notes": "",
		}
		personRec := client.Request(t, "POST", "/picuinhas/persons", personPayload)
		require.Equal(t, std_http.StatusCreated, personRec.Code)

		var person map[string]interface{}
		require.NoError(t, json.Unmarshal(personRec.Body.Bytes(), &person))

		casePayload := map[string]interface{}{
			"person_id":    person["id"],
			"title":        "Pix de apoio",
			"case_type":    "ONE_OFF",
			"total_amount": 30.0,
			"start_date":   "2025-02-01",
		}
		caseRec := client.Request(t, "POST", "/picuinhas/cases", casePayload)
		require.Equal(t, std_http.StatusCreated, caseRec.Code)

		deleteRec := client.Request(t, "DELETE", "/picuinhas/persons/"+toID(person["id"]), nil)
		require.Equal(t, std_http.StatusConflict, deleteRec.Code)
	})
}

func toID(value interface{}) string {
	switch v := value.(type) {
	case float64:
		return fmt.Sprintf("%.0f", v)
	case int:
		return fmt.Sprintf("%d", v)
	case int32:
		return fmt.Sprintf("%d", v)
	case string:
		return v
	default:
		return ""
	}
}
