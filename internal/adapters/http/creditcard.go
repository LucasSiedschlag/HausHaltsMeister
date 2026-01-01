package httpapi

import (
	"context"
	"net/http"
	"strings"
	"time"

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

type cardNetworkRequest struct {
	Code        string `json:"code"`
	DisplayName string `json:"display_name"`
}

type cardNetworkResponse struct {
	Code        string     `json:"code"`
	DisplayName string     `json:"display_name"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   *time.Time `json:"updated_at,omitempty"`
}

type creditCardRequest struct {
	AccountID        string  `json:"account_id"`
	IssuerName       *string `json:"issuer_name"`
	Network          string  `json:"network"`
	Nickname         *string `json:"nickname"`
	Last4            *string `json:"last4"`
	CreditLimitCents *int64  `json:"credit_limit_cents"`
	ClosingDay       int     `json:"closing_day"`
	DueDay           int     `json:"due_day"`
}

type creditCardResponse struct {
	AccountID        string     `json:"account_id"`
	LedgerID         string     `json:"ledger_id"`
	IssuerName       *string    `json:"issuer_name,omitempty"`
	Network          string     `json:"network"`
	Nickname         *string    `json:"nickname,omitempty"`
	Last4            *string    `json:"last4,omitempty"`
	CreditLimitCents *int64     `json:"credit_limit_cents,omitempty"`
	ClosingDay       int        `json:"closing_day"`
	DueDay           int        `json:"due_day"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        *time.Time `json:"updated_at,omitempty"`
}

type installmentPlanRequest struct {
	PurchaseOccurredAt     string  `json:"purchase_occurred_at"`
	Merchant               *string `json:"merchant"`
	Description            string  `json:"description"`
	CategoryID             string  `json:"category_id"`
	TotalAmountCents       int64   `json:"total_amount_cents"`
	InstallmentsCount      int     `json:"installments_count"`
	InstallmentAmountCents int64   `json:"installment_amount_cents"`
	FirstDueMonth          string  `json:"first_due_month"`
}

type installmentPlanPatchRequest struct {
	Status string `json:"status"`
}

type installmentPlanResponse struct {
	ID                     string     `json:"id"`
	LedgerID               string     `json:"ledger_id"`
	CardAccountID          string     `json:"card_account_id"`
	PurchaseOccurredAt     time.Time  `json:"purchase_occurred_at"`
	Merchant               *string    `json:"merchant,omitempty"`
	Description            string     `json:"description"`
	CategoryID             string     `json:"category_id"`
	TotalAmountCents       int64      `json:"total_amount_cents"`
	InstallmentsCount      int        `json:"installments_count"`
	InstallmentAmountCents int64      `json:"installment_amount_cents"`
	FirstDueMonth          time.Time  `json:"first_due_month"`
	Status                 string     `json:"status"`
	CreatedByUserID        string     `json:"created_by_user_id"`
	CreatedAt              time.Time  `json:"created_at"`
	UpdatedAt              *time.Time `json:"updated_at,omitempty"`
}

type installmentResponse struct {
	ID                  string     `json:"id"`
	LedgerID            string     `json:"ledger_id"`
	PlanID              string     `json:"plan_id"`
	InstallmentNo       int        `json:"installment_no"`
	DueMonth            time.Time  `json:"due_month"`
	AmountCents         int64      `json:"amount_cents"`
	Status              string     `json:"status"`
	PostedTransactionID *string    `json:"posted_transaction_id,omitempty"`
	PaidStatementID     *string    `json:"paid_statement_id,omitempty"`
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           *time.Time `json:"updated_at,omitempty"`
}

type installmentPatchRequest struct {
	Status string `json:"status"`
}

type postingResponse struct {
	PostedCount    int      `json:"posted_count"`
	TransactionIDs []string `json:"transaction_ids"`
}

type statementResponse struct {
	ID                   string     `json:"id"`
	LedgerID             string     `json:"ledger_id"`
	CardAccountID        string     `json:"card_account_id"`
	StatementMonth       time.Time  `json:"statement_month"`
	ClosingDate          time.Time  `json:"closing_date"`
	DueDate              time.Time  `json:"due_date"`
	TotalChargesCents    int64      `json:"total_charges_cents"`
	TotalPaymentsCents   int64      `json:"total_payments_cents"`
	Status               string     `json:"status"`
	PaymentTransactionID *string    `json:"payment_transaction_id,omitempty"`
	CreatedAt            time.Time  `json:"created_at"`
	UpdatedAt            *time.Time `json:"updated_at,omitempty"`
}

type statementPayRequest struct {
	StatementID    string `json:"statement_id"`
	PaymentDate    string `json:"payment_date"`
	PayAmountCents int64  `json:"pay_amount_cents"`
	CashAccountID  string `json:"cash_account_id"`
}

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
	if _, ok := GetUser(c); !ok {
		return WriteError(c, http.StatusUnauthorized, "AUTH_INVALID_CREDENTIALS", "Credenciais invalidas", nil)
	}

	items, err := h.Service.ListCardNetworks(c.Request().Context())
	if err != nil {
		return WriteAppError(c, err)
	}

	response := make([]cardNetworkResponse, 0, len(items))
	for _, item := range items {
		response = append(response, toCardNetworkResponse(item))
	}
	return c.JSON(http.StatusOK, response)
}

func (h *CreditCardHandler) CreateNetwork(c echo.Context) error {
	if _, ok := GetUser(c); !ok {
		return WriteError(c, http.StatusUnauthorized, "AUTH_INVALID_CREDENTIALS", "Credenciais invalidas", nil)
	}

	var req cardNetworkRequest
	if err := c.Bind(&req); err != nil {
		return WriteError(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Payload invalido", nil)
	}

	created, err := h.Service.CreateCardNetwork(c.Request().Context(), req.Code, req.DisplayName)
	if err != nil {
		return WriteAppError(c, err)
	}

	return c.JSON(http.StatusCreated, toCardNetworkResponse(created))
}

func (h *CreditCardHandler) UpdateNetwork(c echo.Context) error {
	if _, ok := GetUser(c); !ok {
		return WriteError(c, http.StatusUnauthorized, "AUTH_INVALID_CREDENTIALS", "Credenciais invalidas", nil)
	}

	code := strings.TrimSpace(c.Param("code"))
	if code == "" {
		return WriteError(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Parametros invalidos", map[string]string{"code": "required"})
	}

	var req cardNetworkRequest
	if err := c.Bind(&req); err != nil {
		return WriteError(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Payload invalido", nil)
	}

	updated, err := h.Service.UpdateCardNetwork(c.Request().Context(), code, req.DisplayName)
	if err != nil {
		return WriteAppError(c, err)
	}
	return c.JSON(http.StatusOK, toCardNetworkResponse(updated))
}

func (h *CreditCardHandler) DeleteNetwork(c echo.Context) error {
	if _, ok := GetUser(c); !ok {
		return WriteError(c, http.StatusUnauthorized, "AUTH_INVALID_CREDENTIALS", "Credenciais invalidas", nil)
	}
	code := strings.TrimSpace(c.Param("code"))
	if code == "" {
		return WriteError(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Parametros invalidos", map[string]string{"code": "required"})
	}

	if err := h.Service.DeleteCardNetwork(c.Request().Context(), code); err != nil {
		return WriteAppError(c, err)
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *CreditCardHandler) ListCards(c echo.Context) error {
	user, ok := GetUser(c)
	if !ok {
		return WriteError(c, http.StatusUnauthorized, "AUTH_INVALID_CREDENTIALS", "Credenciais invalidas", nil)
	}
	ledgerID := c.Param("ledgerId")
	if ledgerID == "" {
		return WriteError(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Parametros invalidos", map[string]string{"ledger_id": "required"})
	}

	items, err := h.Service.ListCreditCards(c.Request().Context(), user.ID, ledgerID)
	if err != nil {
		return WriteAppError(c, err)
	}

	response := make([]creditCardResponse, 0, len(items))
	for _, item := range items {
		response = append(response, toCreditCardResponse(item))
	}
	return c.JSON(http.StatusOK, response)
}

func (h *CreditCardHandler) GetCard(c echo.Context) error {
	user, ok := GetUser(c)
	if !ok {
		return WriteError(c, http.StatusUnauthorized, "AUTH_INVALID_CREDENTIALS", "Credenciais invalidas", nil)
	}
	ledgerID := c.Param("ledgerId")
	cardAccountID := c.Param("cardAccountId")
	if ledgerID == "" || cardAccountID == "" {
		return WriteError(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Parametros invalidos", map[string]string{"card_account_id": "required"})
	}

	card, err := h.Service.GetCreditCard(c.Request().Context(), user.ID, ledgerID, cardAccountID)
	if err != nil {
		return WriteAppError(c, err)
	}
	return c.JSON(http.StatusOK, toCreditCardResponse(card))
}

func (h *CreditCardHandler) CreateCard(c echo.Context) error {
	user, ok := GetUser(c)
	if !ok {
		return WriteError(c, http.StatusUnauthorized, "AUTH_INVALID_CREDENTIALS", "Credenciais invalidas", nil)
	}
	ledgerID := c.Param("ledgerId")
	if ledgerID == "" {
		return WriteError(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Parametros invalidos", map[string]string{"ledger_id": "required"})
	}

	var req creditCardRequest
	if err := c.Bind(&req); err != nil {
		return WriteError(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Payload invalido", nil)
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
		return WriteAppError(c, err)
	}

	return c.JSON(http.StatusCreated, toCreditCardResponse(created))
}

func (h *CreditCardHandler) UpdateCard(c echo.Context) error {
	user, ok := GetUser(c)
	if !ok {
		return WriteError(c, http.StatusUnauthorized, "AUTH_INVALID_CREDENTIALS", "Credenciais invalidas", nil)
	}
	ledgerID := c.Param("ledgerId")
	cardAccountID := c.Param("cardAccountId")
	if ledgerID == "" || cardAccountID == "" {
		return WriteError(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Parametros invalidos", map[string]string{"card_account_id": "required"})
	}

	var req creditCardRequest
	if err := c.Bind(&req); err != nil {
		return WriteError(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Payload invalido", nil)
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
		return WriteAppError(c, err)
	}

	return c.JSON(http.StatusOK, toCreditCardResponse(updated))
}

func (h *CreditCardHandler) DeleteCard(c echo.Context) error {
	user, ok := GetUser(c)
	if !ok {
		return WriteError(c, http.StatusUnauthorized, "AUTH_INVALID_CREDENTIALS", "Credenciais invalidas", nil)
	}
	ledgerID := c.Param("ledgerId")
	cardAccountID := c.Param("cardAccountId")
	if ledgerID == "" || cardAccountID == "" {
		return WriteError(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Parametros invalidos", map[string]string{"card_account_id": "required"})
	}

	if err := h.Service.DeleteCreditCard(c.Request().Context(), user.ID, ledgerID, cardAccountID); err != nil {
		return WriteAppError(c, err)
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *CreditCardHandler) CreatePlan(c echo.Context) error {
	user, ok := GetUser(c)
	if !ok {
		return WriteError(c, http.StatusUnauthorized, "AUTH_INVALID_CREDENTIALS", "Credenciais invalidas", nil)
	}
	ledgerID := c.Param("ledgerId")
	cardAccountID := c.Param("cardAccountId")
	if ledgerID == "" || cardAccountID == "" {
		return WriteError(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Parametros invalidos", map[string]string{"card_account_id": "required"})
	}

	var req installmentPlanRequest
	if err := c.Bind(&req); err != nil {
		return WriteError(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Payload invalido", nil)
	}
	purchaseAt, err := parseDateTime(req.PurchaseOccurredAt)
	if err != nil {
		return WriteError(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Parametros invalidos", map[string]string{"purchase_occurred_at": "invalid"})
	}
	firstDue, err := parseMonth(req.FirstDueMonth)
	if err != nil {
		return WriteError(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Parametros invalidos", map[string]string{"first_due_month": "invalid"})
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
		return WriteAppError(c, err)
	}
	return c.JSON(http.StatusCreated, toInstallmentPlanResponse(created))
}

func (h *CreditCardHandler) ListPlans(c echo.Context) error {
	user, ok := GetUser(c)
	if !ok {
		return WriteError(c, http.StatusUnauthorized, "AUTH_INVALID_CREDENTIALS", "Credenciais invalidas", nil)
	}
	ledgerID := c.Param("ledgerId")
	cardAccountID := c.Param("cardAccountId")
	if ledgerID == "" || cardAccountID == "" {
		return WriteError(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Parametros invalidos", map[string]string{"card_account_id": "required"})
	}

	var status *string
	if value := strings.TrimSpace(c.QueryParam("status")); value != "" {
		status = &value
	}

	items, err := h.Service.ListPlans(c.Request().Context(), user.ID, ledgerID, cardAccountID, status)
	if err != nil {
		return WriteAppError(c, err)
	}

	response := make([]installmentPlanResponse, 0, len(items))
	for _, item := range items {
		response = append(response, toInstallmentPlanResponse(item))
	}
	return c.JSON(http.StatusOK, response)
}

func (h *CreditCardHandler) GetPlan(c echo.Context) error {
	user, ok := GetUser(c)
	if !ok {
		return WriteError(c, http.StatusUnauthorized, "AUTH_INVALID_CREDENTIALS", "Credenciais invalidas", nil)
	}
	ledgerID := c.Param("ledgerId")
	cardAccountID := c.Param("cardAccountId")
	planID := c.Param("planId")
	if ledgerID == "" || cardAccountID == "" || planID == "" {
		return WriteError(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Parametros invalidos", map[string]string{"plan_id": "required"})
	}

	plan, err := h.Service.GetPlan(c.Request().Context(), user.ID, ledgerID, cardAccountID, planID)
	if err != nil {
		return WriteAppError(c, err)
	}
	return c.JSON(http.StatusOK, toInstallmentPlanResponse(plan))
}

func (h *CreditCardHandler) UpdatePlan(c echo.Context) error {
	user, ok := GetUser(c)
	if !ok {
		return WriteError(c, http.StatusUnauthorized, "AUTH_INVALID_CREDENTIALS", "Credenciais invalidas", nil)
	}
	ledgerID := c.Param("ledgerId")
	cardAccountID := c.Param("cardAccountId")
	planID := c.Param("planId")
	if ledgerID == "" || cardAccountID == "" || planID == "" {
		return WriteError(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Parametros invalidos", map[string]string{"plan_id": "required"})
	}

	var req installmentPlanPatchRequest
	if err := c.Bind(&req); err != nil {
		return WriteError(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Payload invalido", nil)
	}
	if strings.ToLower(strings.TrimSpace(req.Status)) != "cancelled" {
		return WriteError(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Parametros invalidos", map[string]string{"status": "invalid"})
	}

	if err := h.Service.CancelPlan(c.Request().Context(), user.ID, ledgerID, cardAccountID, planID); err != nil {
		return WriteAppError(c, err)
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *CreditCardHandler) DeletePlan(c echo.Context) error {
	user, ok := GetUser(c)
	if !ok {
		return WriteError(c, http.StatusUnauthorized, "AUTH_INVALID_CREDENTIALS", "Credenciais invalidas", nil)
	}
	ledgerID := c.Param("ledgerId")
	cardAccountID := c.Param("cardAccountId")
	planID := c.Param("planId")
	if ledgerID == "" || cardAccountID == "" || planID == "" {
		return WriteError(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Parametros invalidos", map[string]string{"plan_id": "required"})
	}

	if err := h.Service.CancelPlan(c.Request().Context(), user.ID, ledgerID, cardAccountID, planID); err != nil {
		return WriteAppError(c, err)
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *CreditCardHandler) ListInstallments(c echo.Context) error {
	user, ok := GetUser(c)
	if !ok {
		return WriteError(c, http.StatusUnauthorized, "AUTH_INVALID_CREDENTIALS", "Credenciais invalidas", nil)
	}
	ledgerID := c.Param("ledgerId")
	cardAccountID := c.Param("cardAccountId")
	if ledgerID == "" || cardAccountID == "" {
		return WriteError(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Parametros invalidos", map[string]string{"card_account_id": "required"})
	}

	var month *time.Time
	if value := strings.TrimSpace(c.QueryParam("month")); value != "" {
		parsed, err := parseMonth(value)
		if err != nil {
			return WriteError(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Parametros invalidos", map[string]string{"month": "invalid"})
		}
		month = &parsed
	}
	var status *string
	if value := strings.TrimSpace(c.QueryParam("status")); value != "" {
		status = &value
	}

	items, err := h.Service.ListInstallments(c.Request().Context(), user.ID, ledgerID, cardAccountID, month, status)
	if err != nil {
		return WriteAppError(c, err)
	}

	response := make([]installmentResponse, 0, len(items))
	for _, item := range items {
		response = append(response, toInstallmentResponse(item))
	}
	return c.JSON(http.StatusOK, response)
}

func (h *CreditCardHandler) UpdateInstallment(c echo.Context) error {
	user, ok := GetUser(c)
	if !ok {
		return WriteError(c, http.StatusUnauthorized, "AUTH_INVALID_CREDENTIALS", "Credenciais invalidas", nil)
	}
	ledgerID := c.Param("ledgerId")
	installmentID := c.Param("installmentId")
	if ledgerID == "" || installmentID == "" {
		return WriteError(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Parametros invalidos", map[string]string{"installment_id": "required"})
	}

	var req installmentPatchRequest
	if err := c.Bind(&req); err != nil {
		return WriteError(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Payload invalido", nil)
	}

	updated, err := h.Service.UpdateInstallment(c.Request().Context(), user.ID, ledgerID, installmentID, req.Status)
	if err != nil {
		return WriteAppError(c, err)
	}
	return c.JSON(http.StatusOK, toInstallmentResponse(updated))
}

func (h *CreditCardHandler) PostMonth(c echo.Context) error {
	user, ok := GetUser(c)
	if !ok {
		return WriteError(c, http.StatusUnauthorized, "AUTH_INVALID_CREDENTIALS", "Credenciais invalidas", nil)
	}
	ledgerID := c.Param("ledgerId")
	cardAccountID := c.Param("cardAccountId")
	if ledgerID == "" || cardAccountID == "" {
		return WriteError(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Parametros invalidos", map[string]string{"card_account_id": "required"})
	}
	monthValue := strings.TrimSpace(c.QueryParam("month"))
	if monthValue == "" {
		return WriteError(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Parametros invalidos", map[string]string{"month": "required"})
	}
	month, err := parseMonth(monthValue)
	if err != nil {
		return WriteError(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Parametros invalidos", map[string]string{"month": "invalid"})
	}

	result, err := h.Service.PostMonth(c.Request().Context(), user.ID, ledgerID, cardAccountID, month)
	if err != nil {
		return WriteAppError(c, err)
	}
	return c.JSON(http.StatusOK, postingResponse{PostedCount: result.PostedCount, TransactionIDs: result.Transactions})
}

func (h *CreditCardHandler) ListStatements(c echo.Context) error {
	user, ok := GetUser(c)
	if !ok {
		return WriteError(c, http.StatusUnauthorized, "AUTH_INVALID_CREDENTIALS", "Credenciais invalidas", nil)
	}
	ledgerID := c.Param("ledgerId")
	cardAccountID := c.Param("cardAccountId")
	if ledgerID == "" || cardAccountID == "" {
		return WriteError(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Parametros invalidos", map[string]string{"card_account_id": "required"})
	}

	var month *time.Time
	if value := strings.TrimSpace(c.QueryParam("month")); value != "" {
		parsed, err := parseMonth(value)
		if err != nil {
			return WriteError(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Parametros invalidos", map[string]string{"month": "invalid"})
		}
		month = &parsed
	}

	items, err := h.Service.ListStatements(c.Request().Context(), user.ID, ledgerID, cardAccountID, month)
	if err != nil {
		return WriteAppError(c, err)
	}

	response := make([]statementResponse, 0, len(items))
	for _, item := range items {
		response = append(response, toStatementResponse(item))
	}
	return c.JSON(http.StatusOK, response)
}

func (h *CreditCardHandler) GetStatement(c echo.Context) error {
	user, ok := GetUser(c)
	if !ok {
		return WriteError(c, http.StatusUnauthorized, "AUTH_INVALID_CREDENTIALS", "Credenciais invalidas", nil)
	}
	ledgerID := c.Param("ledgerId")
	cardAccountID := c.Param("cardAccountId")
	statementID := c.Param("statementId")
	if ledgerID == "" || cardAccountID == "" || statementID == "" {
		return WriteError(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Parametros invalidos", map[string]string{"statement_id": "required"})
	}

	statement, err := h.Service.GetStatement(c.Request().Context(), user.ID, ledgerID, cardAccountID, statementID)
	if err != nil {
		return WriteAppError(c, err)
	}
	return c.JSON(http.StatusOK, toStatementResponse(statement))
}

func (h *CreditCardHandler) CloseStatement(c echo.Context) error {
	user, ok := GetUser(c)
	if !ok {
		return WriteError(c, http.StatusUnauthorized, "AUTH_INVALID_CREDENTIALS", "Credenciais invalidas", nil)
	}
	ledgerID := c.Param("ledgerId")
	cardAccountID := c.Param("cardAccountId")
	if ledgerID == "" || cardAccountID == "" {
		return WriteError(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Parametros invalidos", map[string]string{"card_account_id": "required"})
	}
	monthValue := strings.TrimSpace(c.QueryParam("month"))
	if monthValue == "" {
		return WriteError(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Parametros invalidos", map[string]string{"month": "required"})
	}
	month, err := parseMonth(monthValue)
	if err != nil {
		return WriteError(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Parametros invalidos", map[string]string{"month": "invalid"})
	}

	statement, err := h.Service.CloseStatement(c.Request().Context(), user.ID, ledgerID, cardAccountID, month)
	if err != nil {
		return WriteAppError(c, err)
	}
	return c.JSON(http.StatusOK, toStatementResponse(statement))
}

func (h *CreditCardHandler) PayStatement(c echo.Context) error {
	user, ok := GetUser(c)
	if !ok {
		return WriteError(c, http.StatusUnauthorized, "AUTH_INVALID_CREDENTIALS", "Credenciais invalidas", nil)
	}
	ledgerID := c.Param("ledgerId")
	cardAccountID := c.Param("cardAccountId")
	if ledgerID == "" || cardAccountID == "" {
		return WriteError(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Parametros invalidos", map[string]string{"card_account_id": "required"})
	}

	var req statementPayRequest
	if err := c.Bind(&req); err != nil {
		return WriteError(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Payload invalido", nil)
	}
	if strings.TrimSpace(req.StatementID) == "" {
		return WriteError(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Parametros invalidos", map[string]string{"statement_id": "required"})
	}
	if strings.TrimSpace(req.CashAccountID) == "" {
		return WriteError(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Parametros invalidos", map[string]string{"cash_account_id": "required"})
	}
	paymentDate, err := parseDateTime(req.PaymentDate)
	if err != nil {
		return WriteError(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Parametros invalidos", map[string]string{"payment_date": "invalid"})
	}

	statement, err := h.Service.PayStatement(c.Request().Context(), user.ID, ledgerID, cardAccountID, req.StatementID, req.CashAccountID, req.PayAmountCents, paymentDate)
	if err != nil {
		return WriteAppError(c, err)
	}
	return c.JSON(http.StatusOK, toStatementResponse(statement))
}

func (h *CreditCardHandler) UpdateStatement(c echo.Context) error {
	return WriteError(c, http.StatusNotImplemented, "NOT_IMPLEMENTED", "Funcionalidade nao disponivel", nil)
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
