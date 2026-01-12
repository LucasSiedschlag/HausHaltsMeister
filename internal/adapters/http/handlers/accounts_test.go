package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/LucasSiedschlag/HausHaltsMeister/internal/domain/accounts"
	"github.com/LucasSiedschlag/HausHaltsMeister/internal/domain/auth"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
)

type fakeAccountsService struct {
	createErr error
}

func (f fakeAccountsService) ListAccounts(ctx context.Context, userID, ledgerID string) ([]accounts.Account, error) {
	return nil, nil
}

func (f fakeAccountsService) GetAccount(ctx context.Context, userID, ledgerID, accountID string) (accounts.Account, error) {
	return accounts.Account{}, nil
}

func (f fakeAccountsService) CreateAccount(ctx context.Context, userID, ledgerID, name, accountType, nature string, isActive bool) (accounts.Account, error) {
	return accounts.Account{}, f.createErr
}

func (f fakeAccountsService) UpdateAccount(ctx context.Context, userID, ledgerID, accountID, name string, isActive bool) (accounts.Account, error) {
	return accounts.Account{}, nil
}

func (f fakeAccountsService) DeleteAccount(ctx context.Context, userID, ledgerID, accountID string) error {
	return nil
}

func TestCreateAccountAccessDenied(t *testing.T) {
	e := echo.New()
	payload := map[string]interface{}{"name": "Conta", "type": "current"}
	body, err := json.Marshal(payload)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/ledgers/11111111-1111-1111-1111-111111111111/accounts", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("ledgerId")
	c.SetParamValues("11111111-1111-1111-1111-111111111111")
	c.Set("user", auth.User{ID: "user-1"})

	handler := AccountsHandler{Service: fakeAccountsService{createErr: accounts.ErrAccessDenied}}

	err = handler.Create(c)
	require.NoError(t, err)
	require.Equal(t, http.StatusForbidden, rec.Code)
}
