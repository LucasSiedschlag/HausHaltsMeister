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

func (f fakeCreditCardService) ListCreditCards(ctx context.Context, userID, accountID string) ([]creditcard.CreditCard, error) {
	return nil, nil
}

func (f fakeCreditCardService) GetCreditCard(ctx context.Context, userID, cardID string) (creditcard.CreditCard, error) {
	return creditcard.CreditCard{}, nil
}

func (f fakeCreditCardService) CreateCreditCard(ctx context.Context, userID, accountID string, card creditcard.CreditCard) (creditcard.CreditCard, error) {
	return creditcard.CreditCard{}, nil
}

func (f fakeCreditCardService) UpdateCreditCard(ctx context.Context, userID, cardID string, card creditcard.CreditCard) (creditcard.CreditCard, error) {
	return creditcard.CreditCard{}, nil
}

func (f fakeCreditCardService) DeleteCreditCard(ctx context.Context, userID, cardID string) error {
	return nil
}

func (f fakeCreditCardService) CreatePlan(ctx context.Context, userID, cardID string, input creditcard.InstallmentPlan, installmentsCount int, installmentAmount int64) (creditcard.InstallmentPlan, error) {
	return creditcard.InstallmentPlan{}, nil
}

func (f fakeCreditCardService) ListPlans(ctx context.Context, userID, cardID string, status *string) ([]creditcard.InstallmentPlan, error) {
	return nil, nil
}

func (f fakeCreditCardService) GetPlan(ctx context.Context, userID, cardID, planID string) (creditcard.InstallmentPlan, error) {
	return creditcard.InstallmentPlan{}, nil
}

func (f fakeCreditCardService) CancelPlan(ctx context.Context, userID, cardID, planID string) error {
	return nil
}

func (f fakeCreditCardService) ListInstallments(ctx context.Context, userID, cardID string, month *time.Time, status *string) ([]creditcard.Installment, error) {
	return nil, nil
}

func (f fakeCreditCardService) UpdateInstallment(ctx context.Context, userID, cardID, installmentID, status string) (creditcard.Installment, error) {
	return creditcard.Installment{}, nil
}

func (f fakeCreditCardService) PostMonth(ctx context.Context, userID, cardID string, month time.Time) (creditcard.PostingResult, error) {
	return creditcard.PostingResult{}, nil
}

func (f fakeCreditCardService) ListStatements(ctx context.Context, userID, cardID string, month *time.Time) ([]creditcard.Statement, error) {
	return nil, nil
}

func (f fakeCreditCardService) GetStatement(ctx context.Context, userID, cardID, statementID string) (creditcard.Statement, error) {
	return creditcard.Statement{}, nil
}

func (f fakeCreditCardService) CloseStatement(ctx context.Context, userID, cardID string, month time.Time) (creditcard.Statement, error) {
	return creditcard.Statement{}, nil
}

func (f fakeCreditCardService) PayStatement(ctx context.Context, userID, cardID, statementID, payingAccountID string, payAmount int64, paymentDate time.Time) (creditcard.Statement, error) {
	return creditcard.Statement{}, f.payErr
}

func TestPayStatementAlreadyPaid(t *testing.T) {
	e := echo.New()
	payload := map[string]interface{}{
		"statement_id":      "st-1",
		"payment_date":      "2026-02-10",
		"pay_amount_cents":  1000,
		"paying_account_id": "acc-1",
	}
	body, err := json.Marshal(payload)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/credit-cards/33333333-3333-3333-3333-333333333333/statements/pay", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("cardId")
	c.SetParamValues("33333333-3333-3333-3333-333333333333")
	c.Set("user", auth.User{ID: "user-1"})

	handler := CreditCardHandler{Service: fakeCreditCardService{payErr: creditcard.ErrStatementAlreadyPaid}}

	err = handler.PayStatement(c)
	require.NoError(t, err)
	require.Equal(t, http.StatusConflict, rec.Code)
}
