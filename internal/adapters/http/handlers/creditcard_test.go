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
	"github.com/LucasSiedschlag/HausHaltsMeister/internal/domain/creditcard"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
)

type fakeCreditCardService struct {
	payErr error
}

func (f fakeCreditCardService) ListCardNetworks(ctx context.Context) ([]creditcard.CardNetwork, error) {
	return nil, nil
}

func (f fakeCreditCardService) CreateCardNetwork(ctx context.Context, code, displayName string) (creditcard.CardNetwork, error) {
	return creditcard.CardNetwork{}, nil
}

func (f fakeCreditCardService) UpdateCardNetwork(ctx context.Context, code, displayName string) (creditcard.CardNetwork, error) {
	return creditcard.CardNetwork{}, nil
}

func (f fakeCreditCardService) DeleteCardNetwork(ctx context.Context, code string) error {
	return nil
}

func (f fakeCreditCardService) ListCreditCards(ctx context.Context, userID, ledgerID string) ([]creditcard.CreditCard, error) {
	return nil, nil
}

func (f fakeCreditCardService) GetCreditCard(ctx context.Context, userID, ledgerID, cardAccountID string) (creditcard.CreditCard, error) {
	return creditcard.CreditCard{}, nil
}

func (f fakeCreditCardService) CreateCreditCard(ctx context.Context, userID, ledgerID string, card creditcard.CreditCard) (creditcard.CreditCard, error) {
	return creditcard.CreditCard{}, nil
}

func (f fakeCreditCardService) UpdateCreditCard(ctx context.Context, userID, ledgerID, cardAccountID string, card creditcard.CreditCard) (creditcard.CreditCard, error) {
	return creditcard.CreditCard{}, nil
}

func (f fakeCreditCardService) DeleteCreditCard(ctx context.Context, userID, ledgerID, cardAccountID string) error {
	return nil
}

func (f fakeCreditCardService) CreatePlan(ctx context.Context, userID, ledgerID, cardAccountID string, input creditcard.InstallmentPlan, installmentsCount int, installmentAmount int64) (creditcard.InstallmentPlan, error) {
	return creditcard.InstallmentPlan{}, nil
}

func (f fakeCreditCardService) ListPlans(ctx context.Context, userID, ledgerID, cardAccountID string, status *string) ([]creditcard.InstallmentPlan, error) {
	return nil, nil
}

func (f fakeCreditCardService) GetPlan(ctx context.Context, userID, ledgerID, cardAccountID, planID string) (creditcard.InstallmentPlan, error) {
	return creditcard.InstallmentPlan{}, nil
}

func (f fakeCreditCardService) CancelPlan(ctx context.Context, userID, ledgerID, cardAccountID, planID string) error {
	return nil
}

func (f fakeCreditCardService) ListInstallments(ctx context.Context, userID, ledgerID, cardAccountID string, month *time.Time, status *string) ([]creditcard.Installment, error) {
	return nil, nil
}

func (f fakeCreditCardService) UpdateInstallment(ctx context.Context, userID, ledgerID, installmentID, status string) (creditcard.Installment, error) {
	return creditcard.Installment{}, nil
}

func (f fakeCreditCardService) PostMonth(ctx context.Context, userID, ledgerID, cardAccountID string, month time.Time) (creditcard.PostingResult, error) {
	return creditcard.PostingResult{}, nil
}

func (f fakeCreditCardService) ListStatements(ctx context.Context, userID, ledgerID, cardAccountID string, month *time.Time) ([]creditcard.Statement, error) {
	return nil, nil
}

func (f fakeCreditCardService) GetStatement(ctx context.Context, userID, ledgerID, cardAccountID, statementID string) (creditcard.Statement, error) {
	return creditcard.Statement{}, nil
}

func (f fakeCreditCardService) CloseStatement(ctx context.Context, userID, ledgerID, cardAccountID string, month time.Time) (creditcard.Statement, error) {
	return creditcard.Statement{}, nil
}

func (f fakeCreditCardService) PayStatement(ctx context.Context, userID, ledgerID, cardAccountID, statementID, cashAccountID string, payAmount int64, paymentDate time.Time) (creditcard.Statement, error) {
	return creditcard.Statement{}, f.payErr
}

func TestPayStatementAlreadyPaid(t *testing.T) {
	e := echo.New()
	payload := map[string]interface{}{
		"statement_id":     "st-1",
		"payment_date":     "2026-02-10",
		"pay_amount_cents": 1000,
		"cash_account_id":  "acc-1",
	}
	body, err := json.Marshal(payload)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/ledgers/ledger-1/credit-cards/card-1/statements/pay", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("ledgerId", "cardAccountId")
	c.SetParamValues("ledger-1", "card-1")
	c.Set("user", auth.User{ID: "user-1"})

	handler := CreditCardHandler{Service: fakeCreditCardService{payErr: creditcard.ErrStatementAlreadyPaid}}

	err = handler.PayStatement(c)
	require.NoError(t, err)
	require.Equal(t, http.StatusConflict, rec.Code)
}
