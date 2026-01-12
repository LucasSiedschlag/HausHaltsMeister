package handlers

import (
	"context"
	"net/http"
	"strings"

	"github.com/LucasSiedschlag/HausHaltsMeister/internal/adapters/http/dto"
	"github.com/LucasSiedschlag/HausHaltsMeister/internal/adapters/http/httpx"
	"github.com/LucasSiedschlag/HausHaltsMeister/internal/domain/accounts"
	"github.com/labstack/echo/v4"
)

type AccountsHandler struct {
	Service AccountsService
}

type AccountsService interface {
	ListAccounts(ctx context.Context, userID, ledgerID string) ([]accounts.Account, error)
	GetAccount(ctx context.Context, userID, ledgerID, accountID string) (accounts.Account, error)
	CreateAccount(ctx context.Context, userID, ledgerID, name, accountType, nature string, isActive bool) (accounts.Account, error)
	UpdateAccount(ctx context.Context, userID, ledgerID, accountID, name string, isActive bool) (accounts.Account, error)
	DeleteAccount(ctx context.Context, userID, ledgerID, accountID string) error
}

type accountRequest = dto.AccountRequest
type accountResponse = dto.AccountResponse

func (h *AccountsHandler) Register(g *echo.Group) {
	g.GET("", h.List)
	g.POST("", h.Create)
	g.GET("/:accountId", h.Get)
	g.PATCH("/:accountId", h.Update)
	g.DELETE("/:accountId", h.Delete)
}

func (h *AccountsHandler) List(c echo.Context) error {
	user, ok := httpx.GetUser(c)
	if !ok {
		return httpx.WriteError(c, http.StatusUnauthorized, "AUTH_INVALID_CREDENTIALS", "Credenciais invalidas", nil)
	}
	ledgerID := c.Param("ledgerId")
	if ledgerID == "" {
		return httpx.WriteError(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Parametros invalidos", map[string]string{"ledger_id": "required"})
	}

	items, err := h.Service.ListAccounts(c.Request().Context(), user.ID, ledgerID)
	if err != nil {
		return httpx.WriteAppError(c, err)
	}

	response := make([]accountResponse, 0, len(items))
	for _, item := range items {
		response = append(response, toAccountResponse(item))
	}
	return c.JSON(http.StatusOK, response)
}

func (h *AccountsHandler) Get(c echo.Context) error {
	user, ok := httpx.GetUser(c)
	if !ok {
		return httpx.WriteError(c, http.StatusUnauthorized, "AUTH_INVALID_CREDENTIALS", "Credenciais invalidas", nil)
	}
	ledgerID := c.Param("ledgerId")
	accountID, err := httpx.RequireUUIDParam(c, "accountId")
	if err != nil {
		return err
	}

	item, err := h.Service.GetAccount(c.Request().Context(), user.ID, ledgerID, accountID)
	if err != nil {
		return httpx.WriteAppError(c, err)
	}

	return c.JSON(http.StatusOK, toAccountResponse(item))
}

func (h *AccountsHandler) Create(c echo.Context) error {
	user, ok := httpx.GetUser(c)
	if !ok {
		return httpx.WriteError(c, http.StatusUnauthorized, "AUTH_INVALID_CREDENTIALS", "Credenciais invalidas", nil)
	}
	ledgerID := c.Param("ledgerId")
	if ledgerID == "" {
		return httpx.WriteError(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Parametros invalidos", map[string]string{"ledger_id": "required"})
	}

	var req accountRequest
	if err := c.Bind(&req); err != nil {
		return httpx.WriteError(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Payload invalido", nil)
	}

	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	created, err := h.Service.CreateAccount(c.Request().Context(), user.ID, ledgerID, req.Name, req.Type, req.Nature, isActive)
	if err != nil {
		return httpx.WriteAppError(c, err)
	}
	return c.JSON(http.StatusCreated, toAccountResponse(created))
}

func (h *AccountsHandler) Update(c echo.Context) error {
	user, ok := httpx.GetUser(c)
	if !ok {
		return httpx.WriteError(c, http.StatusUnauthorized, "AUTH_INVALID_CREDENTIALS", "Credenciais invalidas", nil)
	}
	ledgerID := c.Param("ledgerId")
	accountID, err := httpx.RequireUUIDParam(c, "accountId")
	if err != nil {
		return err
	}

	var req accountRequest
	if err := c.Bind(&req); err != nil {
		return httpx.WriteError(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Payload invalido", nil)
	}

	if strings.TrimSpace(req.Name) == "" {
		return httpx.WriteError(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Parametros invalidos", map[string]string{"name": "required"})
	}

	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	updated, err := h.Service.UpdateAccount(c.Request().Context(), user.ID, ledgerID, accountID, req.Name, isActive)
	if err != nil {
		return httpx.WriteAppError(c, err)
	}
	return c.JSON(http.StatusOK, toAccountResponse(updated))
}

func (h *AccountsHandler) Delete(c echo.Context) error {
	user, ok := httpx.GetUser(c)
	if !ok {
		return httpx.WriteError(c, http.StatusUnauthorized, "AUTH_INVALID_CREDENTIALS", "Credenciais invalidas", nil)
	}
	ledgerID := c.Param("ledgerId")
	accountID, err := httpx.RequireUUIDParam(c, "accountId")
	if err != nil {
		return err
	}

	if err := h.Service.DeleteAccount(c.Request().Context(), user.ID, ledgerID, accountID); err != nil {
		return httpx.WriteAppError(c, err)
	}
	return c.NoContent(http.StatusNoContent)
}

func toAccountResponse(item accounts.Account) accountResponse {
	return accountResponse{
		ID:        item.ID,
		LedgerID:  item.LedgerID,
		Name:      item.Name,
		Type:      item.Type,
		Nature:    item.Nature,
		IsActive:  item.IsActive,
		CreatedAt: item.CreatedAt,
		UpdatedAt: item.UpdatedAt,
	}
}
