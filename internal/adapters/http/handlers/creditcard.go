package handlers

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/LucasSiedschlag/HausHaltsMeister/internal/adapters/http/dto"
	"github.com/LucasSiedschlag/HausHaltsMeister/internal/adapters/http/httpx"
	"github.com/LucasSiedschlag/HausHaltsMeister/internal/domain/creditcard"
	"github.com/labstack/echo/v4"
)

type CreditCardHandler struct {
	Service CreditCardService
}

type CreditCardService interface {
	ListCardNetworks(ctx context.Context) ([]creditcard.CardNetwork, error)
	CreateCardNetwork(ctx context.Context, code, displayName string) (creditcard.CardNetwork, error)
	UpdateCardNetwork(ctx context.Context, code, displayName string) (creditcard.CardNetwork, error)
	DeleteCardNetwork(ctx context.Context, code string) error

	ListCreditCards(ctx context.Context, userID, ledgerID string) ([]creditcard.CreditCard, error)
	GetCreditCard(ctx context.Context, userID, ledgerID, cardAccountID string) (creditcard.CreditCard, error)
	CreateCreditCard(ctx context.Context, userID, ledgerID string, card creditcard.CreditCard) (creditcard.CreditCard, error)
	UpdateCreditCard(ctx context.Context, userID, ledgerID, cardAccountID string, card creditcard.CreditCard) (creditcard.CreditCard, error)
	DeleteCreditCard(ctx context.Context, userID, ledgerID, cardAccountID string) error

	CreatePlan(ctx context.Context, userID, ledgerID, cardAccountID string, input creditcard.InstallmentPlan, installmentsCount int, installmentAmount int64) (creditcard.InstallmentPlan, error)
	ListPlans(ctx context.Context, userID, ledgerID, cardAccountID string, status *string) ([]creditcard.InstallmentPlan, error)
	GetPlan(ctx context.Context, userID, ledgerID, cardAccountID, planID string) (creditcard.InstallmentPlan, error)
	CancelPlan(ctx context.Context, userID, ledgerID, cardAccountID, planID string) error

	ListInstallments(ctx context.Context, userID, ledgerID, cardAccountID string, month *time.Time, status *string) ([]creditcard.Installment, error)
	UpdateInstallment(ctx context.Context, userID, ledgerID, installmentID, status string) (creditcard.Installment, error)

	PostMonth(ctx context.Context, userID, ledgerID, cardAccountID string, month time.Time) (creditcard.PostingResult, error)

	ListStatements(ctx context.Context, userID, ledgerID, cardAccountID string, month *time.Time) ([]creditcard.Statement, error)
	GetStatement(ctx context.Context, userID, ledgerID, cardAccountID, statementID string) (creditcard.Statement, error)
	CloseStatement(ctx context.Context, userID, ledgerID, cardAccountID string, month time.Time) (creditcard.Statement, error)
	PayStatement(ctx context.Context, userID, ledgerID, cardAccountID, statementID, cashAccountID string, payAmount int64, paymentDate time.Time) (creditcard.Statement, error)
}

type cardNetworkRequest = dto.CardNetworkRequest
type cardNetworkResponse = dto.CardNetworkResponse
type creditCardRequest = dto.CreditCardRequest
type creditCardResponse = dto.CreditCardResponse
type installmentPlanRequest = dto.InstallmentPlanRequest
type installmentPlanPatchRequest = dto.InstallmentPlanPatchRequest
type installmentPlanResponse = dto.InstallmentPlanResponse
type installmentResponse = dto.InstallmentResponse
type installmentPatchRequest = dto.InstallmentPatchRequest
type postingResponse = dto.PostingResponse
type statementResponse = dto.StatementResponse
type statementPayRequest = dto.StatementPayRequest

func (h *CreditCardHandler) Register(base *echo.Group) {
	plans := base.Group("/:cardAccountId/plans")
	plans.POST("", h.CreatePlan)
	plans.GET("", h.ListPlans)
	plans.GET("/:planId", h.GetPlan)
	plans.PATCH("/:planId", h.UpdatePlan)
	plans.DELETE("/:planId", h.DeletePlan)

	installments := base.Group("/:cardAccountId/installments")
	installments.GET("", h.ListInstallments)
	installments.PATCH("/:installmentId", h.UpdateInstallment)

	base.POST("/:cardAccountId/post", h.PostMonth)

	statements := base.Group("/:cardAccountId/statements")
	statements.GET("", h.ListStatements)
	statements.GET("/:statementId", h.GetStatement)
	statements.POST("/close", h.CloseStatement)
	statements.POST("/pay", h.PayStatement)
	statements.PATCH("/:statementId", h.UpdateStatement)

	base.GET("", h.ListCards)
	base.POST("", h.CreateCard)
	base.GET("/:cardAccountId", h.GetCard)
	base.PATCH("/:cardAccountId", h.UpdateCard)
	base.DELETE("/:cardAccountId", h.DeleteCard)
}

func (h *CreditCardHandler) RegisterNetworks(base *echo.Group) {
	base.GET("", h.ListNetworks)
	base.POST("", h.CreateNetwork)
	base.PATCH("/:code", h.UpdateNetwork)
	base.DELETE("/:code", h.DeleteNetwork)
}

func (h *CreditCardHandler) ListNetworks(c echo.Context) error {
	if _, ok := httpx.GetUser(c); !ok {
		return httpx.WriteError(c, http.StatusUnauthorized, "AUTH_INVALID_CREDENTIALS", "Credenciais invalidas", nil)
	}

	items, err := h.Service.ListCardNetworks(c.Request().Context())
	if err != nil {
		return httpx.WriteAppError(c, err)
	}

	response := make([]cardNetworkResponse, 0, len(items))
	for _, item := range items {
		response = append(response, toCardNetworkResponse(item))
	}
	return c.JSON(http.StatusOK, response)
}

func (h *CreditCardHandler) CreateNetwork(c echo.Context) error {
	if _, ok := httpx.GetUser(c); !ok {
		return httpx.WriteError(c, http.StatusUnauthorized, "AUTH_INVALID_CREDENTIALS", "Credenciais invalidas", nil)
	}

	var req cardNetworkRequest
	if err := c.Bind(&req); err != nil {
		return httpx.WriteError(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Payload invalido", nil)
	}

	created, err := h.Service.CreateCardNetwork(c.Request().Context(), req.Code, req.DisplayName)
	if err != nil {
		return httpx.WriteAppError(c, err)
	}

	return c.JSON(http.StatusCreated, toCardNetworkResponse(created))
}

func (h *CreditCardHandler) UpdateNetwork(c echo.Context) error {
	if _, ok := httpx.GetUser(c); !ok {
		return httpx.WriteError(c, http.StatusUnauthorized, "AUTH_INVALID_CREDENTIALS", "Credenciais invalidas", nil)
	}

	code := strings.TrimSpace(c.Param("code"))
	if code == "" {
		return httpx.WriteError(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Parametros invalidos", map[string]string{"code": "required"})
	}

	var req cardNetworkRequest
	if err := c.Bind(&req); err != nil {
		return httpx.WriteError(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Payload invalido", nil)
	}

	updated, err := h.Service.UpdateCardNetwork(c.Request().Context(), code, req.DisplayName)
	if err != nil {
		return httpx.WriteAppError(c, err)
	}
	return c.JSON(http.StatusOK, toCardNetworkResponse(updated))
}

func (h *CreditCardHandler) DeleteNetwork(c echo.Context) error {
	if _, ok := httpx.GetUser(c); !ok {
		return httpx.WriteError(c, http.StatusUnauthorized, "AUTH_INVALID_CREDENTIALS", "Credenciais invalidas", nil)
	}
	code := strings.TrimSpace(c.Param("code"))
	if code == "" {
		return httpx.WriteError(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Parametros invalidos", map[string]string{"code": "required"})
	}

	if err := h.Service.DeleteCardNetwork(c.Request().Context(), code); err != nil {
		return httpx.WriteAppError(c, err)
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *CreditCardHandler) ListCards(c echo.Context) error {
	user, ok := httpx.GetUser(c)
	if !ok {
		return httpx.WriteError(c, http.StatusUnauthorized, "AUTH_INVALID_CREDENTIALS", "Credenciais invalidas", nil)
	}
	ledgerID := c.Param("ledgerId")
	if ledgerID == "" {
		return httpx.WriteError(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Parametros invalidos", map[string]string{"ledger_id": "required"})
	}

	items, err := h.Service.ListCreditCards(c.Request().Context(), user.ID, ledgerID)
	if err != nil {
		return httpx.WriteAppError(c, err)
	}

	response := make([]creditCardResponse, 0, len(items))
	for _, item := range items {
		response = append(response, toCreditCardResponse(item))
	}
	return c.JSON(http.StatusOK, response)
}

func (h *CreditCardHandler) GetCard(c echo.Context) error {
	user, ok := httpx.GetUser(c)
	if !ok {
		return httpx.WriteError(c, http.StatusUnauthorized, "AUTH_INVALID_CREDENTIALS", "Credenciais invalidas", nil)
	}
	ledgerID := c.Param("ledgerId")
	cardAccountID, err := httpx.RequireUUIDParam(c, "cardAccountId")
	if err != nil {
		return err
	}

	card, err := h.Service.GetCreditCard(c.Request().Context(), user.ID, ledgerID, cardAccountID)
	if err != nil {
		return httpx.WriteAppError(c, err)
	}
	return c.JSON(http.StatusOK, toCreditCardResponse(card))
}

func (h *CreditCardHandler) CreateCard(c echo.Context) error {
	user, ok := httpx.GetUser(c)
	if !ok {
		return httpx.WriteError(c, http.StatusUnauthorized, "AUTH_INVALID_CREDENTIALS", "Credenciais invalidas", nil)
	}
	ledgerID := c.Param("ledgerId")
	if ledgerID == "" {
		return httpx.WriteError(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Parametros invalidos", map[string]string{"ledger_id": "required"})
	}

	var req creditCardRequest
	if err := c.Bind(&req); err != nil {
		return httpx.WriteError(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Payload invalido", nil)
	}

	created, err := h.Service.CreateCreditCard(c.Request().Context(), user.ID, ledgerID, creditcard.CreditCard{
		AccountID:        req.AccountID,
		IssuerName:       req.IssuerName,
		Network:          req.Network,
		Nickname:         req.Nickname,
		Last4:            req.Last4,
		CreditLimitCents: req.CreditLimitCents,
		ClosingDay:       req.ClosingDay,
		DueDay:           req.DueDay,
	})
	if err != nil {
		return httpx.WriteAppError(c, err)
	}

	return c.JSON(http.StatusCreated, toCreditCardResponse(created))
}

func (h *CreditCardHandler) UpdateCard(c echo.Context) error {
	user, ok := httpx.GetUser(c)
	if !ok {
		return httpx.WriteError(c, http.StatusUnauthorized, "AUTH_INVALID_CREDENTIALS", "Credenciais invalidas", nil)
	}
	ledgerID := c.Param("ledgerId")
	cardAccountID, err := httpx.RequireUUIDParam(c, "cardAccountId")
	if err != nil {
		return err
	}

	var req creditCardRequest
	if err := c.Bind(&req); err != nil {
		return httpx.WriteError(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Payload invalido", nil)
	}

	updated, err := h.Service.UpdateCreditCard(c.Request().Context(), user.ID, ledgerID, cardAccountID, creditcard.CreditCard{
		AccountID:        cardAccountID,
		IssuerName:       req.IssuerName,
		Network:          req.Network,
		Nickname:         req.Nickname,
		Last4:            req.Last4,
		CreditLimitCents: req.CreditLimitCents,
		ClosingDay:       req.ClosingDay,
		DueDay:           req.DueDay,
	})
	if err != nil {
		return httpx.WriteAppError(c, err)
	}

	return c.JSON(http.StatusOK, toCreditCardResponse(updated))
}

func (h *CreditCardHandler) DeleteCard(c echo.Context) error {
	user, ok := httpx.GetUser(c)
	if !ok {
		return httpx.WriteError(c, http.StatusUnauthorized, "AUTH_INVALID_CREDENTIALS", "Credenciais invalidas", nil)
	}
	ledgerID := c.Param("ledgerId")
	cardAccountID, err := httpx.RequireUUIDParam(c, "cardAccountId")
	if err != nil {
		return err
	}

	if err := h.Service.DeleteCreditCard(c.Request().Context(), user.ID, ledgerID, cardAccountID); err != nil {
		return httpx.WriteAppError(c, err)
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *CreditCardHandler) CreatePlan(c echo.Context) error {
	user, ok := httpx.GetUser(c)
	if !ok {
		return httpx.WriteError(c, http.StatusUnauthorized, "AUTH_INVALID_CREDENTIALS", "Credenciais invalidas", nil)
	}
	ledgerID := c.Param("ledgerId")
	cardAccountID, err := httpx.RequireUUIDParam(c, "cardAccountId")
	if err != nil {
		return err
	}

	var req installmentPlanRequest
	if err := c.Bind(&req); err != nil {
		return httpx.WriteError(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Payload invalido", nil)
	}
	purchaseAt, err := httpx.ParseDateTime(req.PurchaseOccurredAt)
	if err != nil {
		return httpx.WriteError(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Parametros invalidos", map[string]string{"purchase_occurred_at": "invalid"})
	}
	firstDue, err := httpx.ParseMonth(req.FirstDueMonth)
	if err != nil {
		return httpx.WriteError(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Parametros invalidos", map[string]string{"first_due_month": "invalid"})
	}

	created, err := h.Service.CreatePlan(c.Request().Context(), user.ID, ledgerID, cardAccountID, creditcard.InstallmentPlan{
		PurchaseOccurredAt: purchaseAt,
		Merchant:           req.Merchant,
		Description:        req.Description,
		CategoryID:         req.CategoryID,
		TotalAmountCents:   req.TotalAmountCents,
		FirstDueMonth:      firstDue,
	}, req.InstallmentsCount, req.InstallmentAmountCents)
	if err != nil {
		return httpx.WriteAppError(c, err)
	}
	return c.JSON(http.StatusCreated, toInstallmentPlanResponse(created))
}

func (h *CreditCardHandler) ListPlans(c echo.Context) error {
	user, ok := httpx.GetUser(c)
	if !ok {
		return httpx.WriteError(c, http.StatusUnauthorized, "AUTH_INVALID_CREDENTIALS", "Credenciais invalidas", nil)
	}
	ledgerID := c.Param("ledgerId")
	cardAccountID, err := httpx.RequireUUIDParam(c, "cardAccountId")
	if err != nil {
		return err
	}

	var status *string
	if value := strings.TrimSpace(c.QueryParam("status")); value != "" {
		status = &value
	}

	items, err := h.Service.ListPlans(c.Request().Context(), user.ID, ledgerID, cardAccountID, status)
	if err != nil {
		return httpx.WriteAppError(c, err)
	}

	response := make([]installmentPlanResponse, 0, len(items))
	for _, item := range items {
		response = append(response, toInstallmentPlanResponse(item))
	}
	return c.JSON(http.StatusOK, response)
}

func (h *CreditCardHandler) GetPlan(c echo.Context) error {
	user, ok := httpx.GetUser(c)
	if !ok {
		return httpx.WriteError(c, http.StatusUnauthorized, "AUTH_INVALID_CREDENTIALS", "Credenciais invalidas", nil)
	}
	ledgerID := c.Param("ledgerId")
	cardAccountID, err := httpx.RequireUUIDParam(c, "cardAccountId")
	if err != nil {
		return err
	}
	planID, err := httpx.RequireUUIDParam(c, "planId")
	if err != nil {
		return err
	}

	plan, err := h.Service.GetPlan(c.Request().Context(), user.ID, ledgerID, cardAccountID, planID)
	if err != nil {
		return httpx.WriteAppError(c, err)
	}
	return c.JSON(http.StatusOK, toInstallmentPlanResponse(plan))
}

func (h *CreditCardHandler) UpdatePlan(c echo.Context) error {
	user, ok := httpx.GetUser(c)
	if !ok {
		return httpx.WriteError(c, http.StatusUnauthorized, "AUTH_INVALID_CREDENTIALS", "Credenciais invalidas", nil)
	}
	ledgerID := c.Param("ledgerId")
	cardAccountID, err := httpx.RequireUUIDParam(c, "cardAccountId")
	if err != nil {
		return err
	}
	planID, err := httpx.RequireUUIDParam(c, "planId")
	if err != nil {
		return err
	}

	var req installmentPlanPatchRequest
	if err := c.Bind(&req); err != nil {
		return httpx.WriteError(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Payload invalido", nil)
	}
	if strings.ToLower(strings.TrimSpace(req.Status)) != "cancelled" {
		return httpx.WriteError(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Parametros invalidos", map[string]string{"status": "invalid"})
	}

	if err := h.Service.CancelPlan(c.Request().Context(), user.ID, ledgerID, cardAccountID, planID); err != nil {
		return httpx.WriteAppError(c, err)
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *CreditCardHandler) DeletePlan(c echo.Context) error {
	user, ok := httpx.GetUser(c)
	if !ok {
		return httpx.WriteError(c, http.StatusUnauthorized, "AUTH_INVALID_CREDENTIALS", "Credenciais invalidas", nil)
	}
	ledgerID := c.Param("ledgerId")
	cardAccountID, err := httpx.RequireUUIDParam(c, "cardAccountId")
	if err != nil {
		return err
	}
	planID, err := httpx.RequireUUIDParam(c, "planId")
	if err != nil {
		return err
	}

	if err := h.Service.CancelPlan(c.Request().Context(), user.ID, ledgerID, cardAccountID, planID); err != nil {
		return httpx.WriteAppError(c, err)
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *CreditCardHandler) ListInstallments(c echo.Context) error {
	user, ok := httpx.GetUser(c)
	if !ok {
		return httpx.WriteError(c, http.StatusUnauthorized, "AUTH_INVALID_CREDENTIALS", "Credenciais invalidas", nil)
	}
	ledgerID := c.Param("ledgerId")
	cardAccountID, err := httpx.RequireUUIDParam(c, "cardAccountId")
	if err != nil {
		return err
	}

	var month *time.Time
	if value := strings.TrimSpace(c.QueryParam("month")); value != "" {
		parsed, err := httpx.ParseMonth(value)
		if err != nil {
			return httpx.WriteError(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Parametros invalidos", map[string]string{"month": "invalid"})
		}
		month = &parsed
	}
	var status *string
	if value := strings.TrimSpace(c.QueryParam("status")); value != "" {
		status = &value
	}

	items, err := h.Service.ListInstallments(c.Request().Context(), user.ID, ledgerID, cardAccountID, month, status)
	if err != nil {
		return httpx.WriteAppError(c, err)
	}

	response := make([]installmentResponse, 0, len(items))
	for _, item := range items {
		response = append(response, toInstallmentResponse(item))
	}
	return c.JSON(http.StatusOK, response)
}

func (h *CreditCardHandler) UpdateInstallment(c echo.Context) error {
	user, ok := httpx.GetUser(c)
	if !ok {
		return httpx.WriteError(c, http.StatusUnauthorized, "AUTH_INVALID_CREDENTIALS", "Credenciais invalidas", nil)
	}
	ledgerID := c.Param("ledgerId")
	installmentID, err := httpx.RequireUUIDParam(c, "installmentId")
	if err != nil {
		return err
	}

	var req installmentPatchRequest
	if err := c.Bind(&req); err != nil {
		return httpx.WriteError(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Payload invalido", nil)
	}

	updated, err := h.Service.UpdateInstallment(c.Request().Context(), user.ID, ledgerID, installmentID, req.Status)
	if err != nil {
		return httpx.WriteAppError(c, err)
	}
	return c.JSON(http.StatusOK, toInstallmentResponse(updated))
}

func (h *CreditCardHandler) PostMonth(c echo.Context) error {
	user, ok := httpx.GetUser(c)
	if !ok {
		return httpx.WriteError(c, http.StatusUnauthorized, "AUTH_INVALID_CREDENTIALS", "Credenciais invalidas", nil)
	}
	ledgerID := c.Param("ledgerId")
	cardAccountID, err := httpx.RequireUUIDParam(c, "cardAccountId")
	if err != nil {
		return err
	}
	monthValue := strings.TrimSpace(c.QueryParam("month"))
	if monthValue == "" {
		return httpx.WriteError(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Parametros invalidos", map[string]string{"month": "required"})
	}
	month, err := httpx.ParseMonth(monthValue)
	if err != nil {
		return httpx.WriteError(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Parametros invalidos", map[string]string{"month": "invalid"})
	}

	result, err := h.Service.PostMonth(c.Request().Context(), user.ID, ledgerID, cardAccountID, month)
	if err != nil {
		return httpx.WriteAppError(c, err)
	}
	return c.JSON(http.StatusOK, postingResponse{PostedCount: result.PostedCount, TransactionIDs: result.Transactions})
}

func (h *CreditCardHandler) ListStatements(c echo.Context) error {
	user, ok := httpx.GetUser(c)
	if !ok {
		return httpx.WriteError(c, http.StatusUnauthorized, "AUTH_INVALID_CREDENTIALS", "Credenciais invalidas", nil)
	}
	ledgerID := c.Param("ledgerId")
	cardAccountID, err := httpx.RequireUUIDParam(c, "cardAccountId")
	if err != nil {
		return err
	}

	var month *time.Time
	if value := strings.TrimSpace(c.QueryParam("month")); value != "" {
		parsed, err := httpx.ParseMonth(value)
		if err != nil {
			return httpx.WriteError(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Parametros invalidos", map[string]string{"month": "invalid"})
		}
		month = &parsed
	}

	items, err := h.Service.ListStatements(c.Request().Context(), user.ID, ledgerID, cardAccountID, month)
	if err != nil {
		return httpx.WriteAppError(c, err)
	}

	response := make([]statementResponse, 0, len(items))
	for _, item := range items {
		response = append(response, toStatementResponse(item))
	}
	return c.JSON(http.StatusOK, response)
}

func (h *CreditCardHandler) GetStatement(c echo.Context) error {
	user, ok := httpx.GetUser(c)
	if !ok {
		return httpx.WriteError(c, http.StatusUnauthorized, "AUTH_INVALID_CREDENTIALS", "Credenciais invalidas", nil)
	}
	ledgerID := c.Param("ledgerId")
	cardAccountID, err := httpx.RequireUUIDParam(c, "cardAccountId")
	if err != nil {
		return err
	}
	statementID, err := httpx.RequireUUIDParam(c, "statementId")
	if err != nil {
		return err
	}

	statement, err := h.Service.GetStatement(c.Request().Context(), user.ID, ledgerID, cardAccountID, statementID)
	if err != nil {
		return httpx.WriteAppError(c, err)
	}
	return c.JSON(http.StatusOK, toStatementResponse(statement))
}

func (h *CreditCardHandler) CloseStatement(c echo.Context) error {
	user, ok := httpx.GetUser(c)
	if !ok {
		return httpx.WriteError(c, http.StatusUnauthorized, "AUTH_INVALID_CREDENTIALS", "Credenciais invalidas", nil)
	}
	ledgerID := c.Param("ledgerId")
	cardAccountID, err := httpx.RequireUUIDParam(c, "cardAccountId")
	if err != nil {
		return err
	}
	monthValue := strings.TrimSpace(c.QueryParam("month"))
	if monthValue == "" {
		return httpx.WriteError(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Parametros invalidos", map[string]string{"month": "required"})
	}
	month, err := httpx.ParseMonth(monthValue)
	if err != nil {
		return httpx.WriteError(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Parametros invalidos", map[string]string{"month": "invalid"})
	}

	statement, err := h.Service.CloseStatement(c.Request().Context(), user.ID, ledgerID, cardAccountID, month)
	if err != nil {
		return httpx.WriteAppError(c, err)
	}
	return c.JSON(http.StatusOK, toStatementResponse(statement))
}

func (h *CreditCardHandler) PayStatement(c echo.Context) error {
	user, ok := httpx.GetUser(c)
	if !ok {
		return httpx.WriteError(c, http.StatusUnauthorized, "AUTH_INVALID_CREDENTIALS", "Credenciais invalidas", nil)
	}
	ledgerID := c.Param("ledgerId")
	cardAccountID, err := httpx.RequireUUIDParam(c, "cardAccountId")
	if err != nil {
		return err
	}

	var req statementPayRequest
	if err := c.Bind(&req); err != nil {
		return httpx.WriteError(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Payload invalido", nil)
	}
	if strings.TrimSpace(req.StatementID) == "" {
		return httpx.WriteError(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Parametros invalidos", map[string]string{"statement_id": "required"})
	}
	if strings.TrimSpace(req.CashAccountID) == "" {
		return httpx.WriteError(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Parametros invalidos", map[string]string{"cash_account_id": "required"})
	}
	paymentDate, err := httpx.ParseDateTime(req.PaymentDate)
	if err != nil {
		return httpx.WriteError(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Parametros invalidos", map[string]string{"payment_date": "invalid"})
	}

	statement, err := h.Service.PayStatement(c.Request().Context(), user.ID, ledgerID, cardAccountID, req.StatementID, req.CashAccountID, req.PayAmountCents, paymentDate)
	if err != nil {
		return httpx.WriteAppError(c, err)
	}
	return c.JSON(http.StatusOK, toStatementResponse(statement))
}

func (h *CreditCardHandler) UpdateStatement(c echo.Context) error {
	return httpx.WriteError(c, http.StatusNotImplemented, "NOT_IMPLEMENTED", "Funcionalidade nao disponivel", nil)
}

func toCardNetworkResponse(item creditcard.CardNetwork) cardNetworkResponse {
	return cardNetworkResponse{
		Code:        item.Code,
		DisplayName: item.DisplayName,
		CreatedAt:   item.CreatedAt,
		UpdatedAt:   item.UpdatedAt,
	}
}

func toCreditCardResponse(card creditcard.CreditCard) creditCardResponse {
	return creditCardResponse{
		AccountID:        card.AccountID,
		LedgerID:         card.LedgerID,
		IssuerName:       card.IssuerName,
		Network:          card.Network,
		Nickname:         card.Nickname,
		Last4:            card.Last4,
		CreditLimitCents: card.CreditLimitCents,
		ClosingDay:       card.ClosingDay,
		DueDay:           card.DueDay,
		CreatedAt:        card.CreatedAt,
		UpdatedAt:        card.UpdatedAt,
	}
}

func toInstallmentPlanResponse(plan creditcard.InstallmentPlan) installmentPlanResponse {
	return installmentPlanResponse{
		ID:                     plan.ID,
		LedgerID:               plan.LedgerID,
		CardAccountID:          plan.CardAccountID,
		PurchaseOccurredAt:     plan.PurchaseOccurredAt,
		Merchant:               plan.Merchant,
		Description:            plan.Description,
		CategoryID:             plan.CategoryID,
		TotalAmountCents:       plan.TotalAmountCents,
		InstallmentsCount:      plan.InstallmentsCount,
		InstallmentAmountCents: plan.InstallmentAmountCents,
		FirstDueMonth:          plan.FirstDueMonth,
		Status:                 plan.Status,
		CreatedByUserID:        plan.CreatedByUserID,
		CreatedAt:              plan.CreatedAt,
		UpdatedAt:              plan.UpdatedAt,
	}
}

func toInstallmentResponse(inst creditcard.Installment) installmentResponse {
	return installmentResponse{
		ID:                  inst.ID,
		LedgerID:            inst.LedgerID,
		PlanID:              inst.PlanID,
		InstallmentNo:       inst.InstallmentNo,
		DueMonth:            inst.DueMonth,
		AmountCents:         inst.AmountCents,
		Status:              inst.Status,
		PostedTransactionID: inst.PostedTransactionID,
		PaidStatementID:     inst.PaidStatementID,
		CreatedAt:           inst.CreatedAt,
		UpdatedAt:           inst.UpdatedAt,
	}
}

func toStatementResponse(statement creditcard.Statement) statementResponse {
	return statementResponse{
		ID:                   statement.ID,
		LedgerID:             statement.LedgerID,
		CardAccountID:        statement.CardAccountID,
		StatementMonth:       statement.StatementMonth,
		ClosingDate:          statement.ClosingDate,
		DueDate:              statement.DueDate,
		TotalChargesCents:    statement.TotalChargesCents,
		TotalPaymentsCents:   statement.TotalPaymentsCents,
		Status:               statement.Status,
		PaymentTransactionID: statement.PaymentTransactionID,
		CreatedAt:            statement.CreatedAt,
		UpdatedAt:            statement.UpdatedAt,
	}
}
