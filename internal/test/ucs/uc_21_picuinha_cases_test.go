package ucs

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"

	httpAdapter "github.com/LucasSiedschlag/HausHaltsMeister/internal/adapters/http"
	"github.com/LucasSiedschlag/HausHaltsMeister/internal/adapters/postgres"
	"github.com/LucasSiedschlag/HausHaltsMeister/internal/domain/picuinha"
	"github.com/LucasSiedschlag/HausHaltsMeister/internal/test/harness"
)

func TestUC21_PicuinhaCases(t *testing.T) {
	db := harness.SetupTestDB(t)
	defer db.Pool.Close()

	picRepo := postgres.NewPicuinhaRepository(db.Pool)
	picService := picuinha.NewService(picRepo)
	picHandler := httpAdapter.NewPicuinhaHandler(picService)

	e := echo.New()
	httpAdapter.RegisterPicuinhaRoutes(e, picHandler)
	client := harness.NewHTTPClient(e)

	personPayload := map[string]interface{}{
		"name":  "Joaozinho",
		"notes": "",
	}
	personRec := client.Request(t, "POST", "/picuinhas/persons", personPayload)
	require.Equal(t, http.StatusCreated, personRec.Code)

	var person map[string]interface{}
	require.NoError(t, json.Unmarshal(personRec.Body.Bytes(), &person))

	casePayload := map[string]interface{}{
		"person_id":    person["id"],
		"title":        "Pix emergencial",
		"case_type":    "ONE_OFF",
		"total_amount": 200.0,
		"start_date":   "2025-01-01",
	}
	caseRec := client.Request(t, "POST", "/picuinhas/cases", casePayload)
	require.Equal(t, http.StatusCreated, caseRec.Code)

	listRec := client.Request(t, "GET", "/picuinhas/cases?person_id="+toID(person["id"]), nil)
	require.Equal(t, http.StatusOK, listRec.Code)

	var cases []map[string]interface{}
	require.NoError(t, json.Unmarshal(listRec.Body.Bytes(), &cases))
	require.Len(t, cases, 1)
	require.Equal(t, "OPEN", cases[0]["status"])

	caseID := toID(cases[0]["id"])
	instRec := client.Request(t, "GET", "/picuinhas/cases/"+caseID+"/installments", nil)
	require.Equal(t, http.StatusOK, instRec.Code)

	var installments []map[string]interface{}
	require.NoError(t, json.Unmarshal(instRec.Body.Bytes(), &installments))
	require.Len(t, installments, 1)

	installmentID := toID(installments[0]["id"])
	updatePayload := map[string]interface{}{
		"is_paid":      true,
		"extra_amount": 5.0,
	}
	updateRec := client.Request(t, "PUT", "/picuinhas/installments/"+installmentID, updatePayload)
	require.Equal(t, http.StatusOK, updateRec.Code)

	listRec = client.Request(t, "GET", "/picuinhas/cases?person_id="+toID(person["id"]), nil)
	require.Equal(t, http.StatusOK, listRec.Code)
	require.NoError(t, json.Unmarshal(listRec.Body.Bytes(), &cases))
	require.Len(t, cases, 1)
	require.Equal(t, "PAID", cases[0]["status"])
}
