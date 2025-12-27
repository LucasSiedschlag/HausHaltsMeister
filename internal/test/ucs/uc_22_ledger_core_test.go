package ucs

import (
	"encoding/json"
	std_http "net/http"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"

	"github.com/LucasSiedschlag/HausHaltsMeister/internal/adapters/http"
	"github.com/LucasSiedschlag/HausHaltsMeister/internal/adapters/postgres"
	"github.com/LucasSiedschlag/HausHaltsMeister/internal/domain/ledger"
	"github.com/LucasSiedschlag/HausHaltsMeister/internal/test/harness"
)

func TestUC22_LedgerCore(t *testing.T) {
	db := harness.SetupTestDB(t)
	defer db.Pool.Close()

	ledgerRepo := postgres.NewLedgerRepository(db.Pool)
	ledgerService := ledger.NewService(ledgerRepo)
	ledgerHandler := http.NewLedgerHandler(ledgerService)

	e := echo.New()
	http.RegisterLedgerRoutes(e, ledgerHandler)
	client := harness.NewHTTPClient(e)

	createAccount := func(name, accountType string) int32 {
		payload := map[string]interface{}{
			"name": name,
			"type": accountType,
		}
		rec := client.Request(t, "POST", "/ledger/accounts", payload)
		require.Equal(t, std_http.StatusCreated, rec.Code)

		var resp map[string]interface{}
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
		return int32(resp["id"].(float64))
	}

	bankID := createAccount("Banco", "ASSET")
	incomeID := createAccount("Salario", "INCOME")

	t.Run("UC-22: Create balanced transaction", func(t *testing.T) {
		payload := map[string]interface{}{
			"occurred_at": "2026-01-10",
			"description": "Salario janeiro",
			"postings": []map[string]interface{}{
				{
					"account_id": bankID,
					"side":       "DEBIT",
					"amount":     5000.00,
				},
				{
					"account_id": incomeID,
					"side":       "CREDIT",
					"amount":     5000.00,
				},
			},
		}

		rec := client.Request(t, "POST", "/ledger/transactions", payload)
		require.Equal(t, std_http.StatusCreated, rec.Code)

		var resp map[string]interface{}
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
		require.Equal(t, "Salario janeiro", resp["description"])
	})

	t.Run("UC-22: List transactions by month", func(t *testing.T) {
		rec := client.Request(t, "GET", "/ledger/transactions?month=2026-01-01", nil)
		require.Equal(t, std_http.StatusOK, rec.Code)

		var resp []map[string]interface{}
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
		require.Len(t, resp, 1)
	})

	t.Run("Error: Unbalanced transaction", func(t *testing.T) {
		payload := map[string]interface{}{
			"occurred_at": "2026-01-11",
			"description": "Erro de balanceamento",
			"postings": []map[string]interface{}{
				{
					"account_id": bankID,
					"side":       "DEBIT",
					"amount":     100.00,
				},
				{
					"account_id": incomeID,
					"side":       "CREDIT",
					"amount":     80.00,
				},
			},
		}

		rec := client.Request(t, "POST", "/ledger/transactions", payload)
		require.Equal(t, std_http.StatusBadRequest, rec.Code)
	})
}
