package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/LucasSiedschlag/HausHaltsMeister/internal/domain/auth"
	"github.com/LucasSiedschlag/HausHaltsMeister/internal/domain/journal"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
)

type fakeJournalService struct {
	createErr error
}

func (f fakeJournalService) CreateTransaction(ctx context.Context, userID, ledgerID string, params journal.CreateTransactionParams) (journal.Transaction, error) {
	return journal.Transaction{}, f.createErr
}

func (f fakeJournalService) ListTransactions(ctx context.Context, userID, ledgerID string, params journal.ListTransactionsParams) (journal.ListResult, error) {
	return journal.ListResult{}, nil
}

func (f fakeJournalService) GetTransaction(ctx context.Context, userID, ledgerID, transactionID string) (journal.Transaction, error) {
	return journal.Transaction{}, nil
}

func (f fakeJournalService) UpdateTransaction(ctx context.Context, userID, ledgerID, transactionID string, params journal.UpdateTransactionParams) (journal.Transaction, error) {
	return journal.Transaction{}, nil
}

func (f fakeJournalService) DeleteTransaction(ctx context.Context, userID, ledgerID, transactionID string) error {
	return nil
}

func TestCreateTransactionTransferNotBalanced(t *testing.T) {
	e := echo.New()
	payload := map[string]interface{}{
		"occurred_at": "2026-01-10",
		"description": "Transfer",
		"entries": []map[string]interface{}{
			{"account_id": "acc-1", "category_id": "cat-out", "kind": "transfer", "amount_cents": 100},
			{"account_id": "acc-2", "category_id": "cat-in", "kind": "transfer", "amount_cents": 50},
		},
	}
	body, err := json.Marshal(payload)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/ledgers/ledger-1/transactions", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("ledgerId")
	c.SetParamValues("ledger-1")
	c.Set("user", auth.User{ID: "user-1"})

	handler := JournalHandler{Service: fakeJournalService{createErr: journal.ErrTransferNotBalanced}}

	err = handler.Create(c)
	require.NoError(t, err)
	require.Equal(t, http.StatusUnprocessableEntity, rec.Code)
}

func TestListTransactionsInvalidCursor(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/ledgers/ledger-1/transactions?cursor_occurred_at=invalid&cursor_id=1", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("ledgerId")
	c.SetParamValues("ledger-1")
	c.Set("user", auth.User{ID: "user-1"})

	handler := JournalHandler{Service: fakeJournalService{}}

	err := handler.List(c)
	require.NoError(t, err)
	require.Equal(t, http.StatusUnprocessableEntity, rec.Code)
}

func TestParseDateTime(t *testing.T) {
	parsed, err := parseDateTime("2026-01-10")
	require.NoError(t, err)
	require.Equal(t, 2026, parsed.Year())

	parsed, err = parseDateTime(time.Now().UTC().Format(time.RFC3339))
	require.NoError(t, err)
}
