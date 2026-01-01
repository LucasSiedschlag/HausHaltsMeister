package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/LucasSiedschlag/HausHaltsMeister/internal/domain/auth"
	"github.com/LucasSiedschlag/HausHaltsMeister/internal/domain/investments"
	"github.com/LucasSiedschlag/HausHaltsMeister/internal/domain/journal"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
)

type fakeInvestmentsService struct {
	contributionErr error
}

func (f fakeInvestmentsService) Contribution(ctx context.Context, userID, ledgerID string, amount int64, occurredAt time.Time, memo *string) (journal.Transaction, error) {
	return journal.Transaction{}, f.contributionErr
}

func (f fakeInvestmentsService) Redemption(ctx context.Context, userID, ledgerID string, amount int64, occurredAt time.Time, memo *string) (journal.Transaction, error) {
	return journal.Transaction{}, nil
}

func (f fakeInvestmentsService) Earnings(ctx context.Context, userID, ledgerID string, amount int64, occurredAt time.Time, memo *string) (journal.Transaction, error) {
	return journal.Transaction{}, nil
}

func (f fakeInvestmentsService) Summary(ctx context.Context, userID, ledgerID string, from, to time.Time) (investments.Summary, error) {
	return investments.Summary{}, nil
}

func TestContributionValidationError(t *testing.T) {
	e := echo.New()
	payload := map[string]interface{}{"amount_cents": -10, "occurred_at": "2026-01-10"}
	body, err := json.Marshal(payload)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/ledgers/ledger-1/investments/contributions", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("ledgerId")
	c.SetParamValues("ledger-1")
	c.Set("user", auth.User{ID: "user-1"})

	handler := InvestmentsHandler{Service: fakeInvestmentsService{contributionErr: investments.ErrValidation}}

	err = handler.Contribution(c)
	require.NoError(t, err)
	require.Equal(t, http.StatusUnprocessableEntity, rec.Code)
}

func TestSummaryMissingRange(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/ledgers/ledger-1/investments/summary", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("ledgerId")
	c.SetParamValues("ledger-1")
	c.Set("user", auth.User{ID: "user-1"})

	handler := InvestmentsHandler{Service: fakeInvestmentsService{}}

	err := handler.Summary(c)
	require.NoError(t, err)
	require.Equal(t, http.StatusUnprocessableEntity, rec.Code)
}
