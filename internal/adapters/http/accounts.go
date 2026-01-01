package httpapi

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/LucasSiedschlag/HausHaltsMeister/internal/domain/accounts"
	"github.com/labstack/echo/v4"
)

type AccountsHandler struct {
	Service AccountsService
}

type AccountsService interface {
	ListAccounts(ctx context.Context, userID, ledgerID string) ([]accounts.Account, error)
	GetAccount(ctx context.Context, userID, ledgerID, accountID string) (accounts.Account, error)
	CreateAccount(ctx context.Context, userID, ledgerID, name, accountType string, isActive bool) (accounts.Account, error)
	UpdateAccount(ctx context.Context, userID, ledgerID, accountID, name string, isActive bool) (accounts.Account, error)
	DeleteAccount(ctx context.Context, userID, ledgerID, accountID string) error
}

type accountRequest struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	IsActive *bool  `json:"is_active"`
}

type accountResponse struct {
	ID        string     `json:"id"`
	LedgerID  string     `json:"ledger_id"`
	Name      string     `json:"name"`
	Type      string     `json:"type"`
	IsActive  bool       `json:"is_active"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt *time.Time `json:"updated_at,omitempty"`
}

func (h *AccountsHandler) Register(g *echo.Group) {
	g.GET("", h.List)
	g.POST("", h.Create)
	g.GET("/:accountId", h.Get)
	g.PATCH("/:accountId", h.Update)
	g.DELETE("/:accountId", h.Delete)
}

func (h *AccountsHandler) List(c echo.Context) error {
	user, ok := GetUser(c)
	if !ok {
		return WriteError(c, http.StatusUnauthorized, "AUTH_INVALID_CREDENTIALS", "Credenciais invalidas", nil)
	}
	ledgerID := c.Param("ledgerId")
	if ledgerID == "" {
		return WriteError(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Parametros invalidos", map[string]string{"ledger_id": "required"})
	}

	items, err := h.Service.ListAccounts(c.Request().Context(), user.ID, ledgerID)
	if err != nil {
		return WriteAppError(c, err)
	}

	response := make([]accountResponse, 0, len(items))
	for _, item := range items {
		response = append(response, toAccountResponse(item))
	}
	return c.JSON(http.StatusOK, response)
}

func (h *AccountsHandler) Get(c echo.Context) error {
	user, ok := GetUser(c)
	if !ok {
		return WriteError(c, http.StatusUnauthorized, "AUTH_INVALID_CREDENTIALS", "Credenciais invalidas", nil)
	}
	ledgerID := c.Param("ledgerId")
	accountID := c.Param("accountId")
	if ledgerID == "" || accountID == "" {
		return WriteError(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Parametros invalidos", map[string]string{"account_id": "required"})
	}

	item, err := h.Service.GetAccount(c.Request().Context(), user.ID, ledgerID, accountID)
	if err != nil {
		return WriteAppError(c, err)
	}

	return c.JSON(http.StatusOK, toAccountResponse(item))
}

func (h *AccountsHandler) Create(c echo.Context) error {
	user, ok := GetUser(c)
	if !ok {
		return WriteError(c, http.StatusUnauthorized, "AUTH_INVALID_CREDENTIALS", "Credenciais invalidas", nil)
	}
	ledgerID := c.Param("ledgerId")
	if ledgerID == "" {
		return WriteError(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Parametros invalidos", map[string]string{"ledger_id": "required"})
	}

	var req accountRequest
	if err := c.Bind(&req); err != nil {
		return WriteError(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Payload invalido", nil)
	}

	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	created, err := h.Service.CreateAccount(c.Request().Context(), user.ID, ledgerID, req.Name, req.Type, isActive)
	if err != nil {
		return WriteAppError(c, err)
	}
	return c.JSON(http.StatusCreated, toAccountResponse(created))
}

func (h *AccountsHandler) Update(c echo.Context) error {
	user, ok := GetUser(c)
	if !ok {
		return WriteError(c, http.StatusUnauthorized, "AUTH_INVALID_CREDENTIALS", "Credenciais invalidas", nil)
	}
	ledgerID := c.Param("ledgerId")
	accountID := c.Param("accountId")
	if ledgerID == "" || accountID == "" {
		return WriteError(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Parametros invalidos", map[string]string{"account_id": "required"})
	}

	var req accountRequest
	if err := c.Bind(&req); err != nil {
		return WriteError(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Payload invalido", nil)
	}

	if strings.TrimSpace(req.Name) == "" {
		return WriteError(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Parametros invalidos", map[string]string{"name": "required"})
	}

	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	updated, err := h.Service.UpdateAccount(c.Request().Context(), user.ID, ledgerID, accountID, req.Name, isActive)
	if err != nil {
		return WriteAppError(c, err)
	}
	return c.JSON(http.StatusOK, toAccountResponse(updated))
}

func (h *AccountsHandler) Delete(c echo.Context) error {
	user, ok := GetUser(c)
	if !ok {
		return WriteError(c, http.StatusUnauthorized, "AUTH_INVALID_CREDENTIALS", "Credenciais invalidas", nil)
	}
	ledgerID := c.Param("ledgerId")
	accountID := c.Param("accountId")
	if ledgerID == "" || accountID == "" {
		return WriteError(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Parametros invalidos", map[string]string{"account_id": "required"})
	}

	if err := h.Service.DeleteAccount(c.Request().Context(), user.ID, ledgerID, accountID); err != nil {
		return WriteAppError(c, err)
	}
	return c.NoContent(http.StatusNoContent)
}

func toAccountResponse(item accounts.Account) accountResponse {
	return accountResponse{
		ID:        item.ID,
		LedgerID:  item.LedgerID,
		Name:      item.Name,
		Type:      item.Type,
		IsActive:  item.IsActive,
		CreatedAt: item.CreatedAt,
		UpdatedAt: item.UpdatedAt,
	}
}
