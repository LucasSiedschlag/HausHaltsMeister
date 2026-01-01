package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/LucasSiedschlag/HausHaltsMeister/internal/domain/auth"
	"github.com/LucasSiedschlag/HausHaltsMeister/internal/domain/ledger"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
)

type fakeLedgerService struct {
	updateErr error
}

func (f fakeLedgerService) ListLedgers(ctx context.Context, userID string) ([]ledger.LedgerWithRole, error) {
	return nil, nil
}

func (f fakeLedgerService) CreateLedger(ctx context.Context, userID, name, currencyCode string) (ledger.Ledger, error) {
	return ledger.Ledger{}, nil
}

func (f fakeLedgerService) GetLedger(ctx context.Context, userID, ledgerID string) (ledger.LedgerWithRole, error) {
	return ledger.LedgerWithRole{}, nil
}

func (f fakeLedgerService) UpdateLedger(ctx context.Context, userID, ledgerID, name string) (ledger.Ledger, error) {
	return ledger.Ledger{}, f.updateErr
}

func (f fakeLedgerService) DeleteLedger(ctx context.Context, userID, ledgerID string) error {
	return nil
}

func (f fakeLedgerService) ListMembers(ctx context.Context, userID, ledgerID string) ([]ledger.Member, error) {
	return nil, nil
}

func (f fakeLedgerService) AddMember(ctx context.Context, userID, ledgerID, memberUserID, role string) (ledger.Member, error) {
	return ledger.Member{}, nil
}

func (f fakeLedgerService) UpdateMember(ctx context.Context, userID, ledgerID, memberUserID, role string) (ledger.Member, error) {
	return ledger.Member{}, nil
}

func (f fakeLedgerService) RemoveMember(ctx context.Context, userID, ledgerID, memberUserID string) error {
	return nil
}

func TestUpdateLedgerAccessDenied(t *testing.T) {
	e := echo.New()
	payload := map[string]string{"name": "Novo"}
	body, err := json.Marshal(payload)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPatch, "/ledgers/ledger-1", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("ledgerId")
	c.SetParamValues("ledger-1")
	c.Set("user", auth.User{ID: "user-1"})

	handler := LedgerHandler{Service: fakeLedgerService{updateErr: ledger.ErrAccessDenied}}

	err = handler.UpdateLedger(c)
	require.NoError(t, err)
	require.Equal(t, http.StatusForbidden, rec.Code)
}
