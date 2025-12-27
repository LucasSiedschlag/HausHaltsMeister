package http

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/LucasSiedschlag/HausHaltsMeister/internal/adapters/http/dto"
	"github.com/LucasSiedschlag/HausHaltsMeister/internal/domain/ledger"
	"github.com/labstack/echo/v4"
)

type LedgerHandler struct {
	service ledger.Service
}

func NewLedgerHandler(service ledger.Service) *LedgerHandler {
	return &LedgerHandler{service: service}
}

// CreateAccount creates a new ledger account.
// @Summary Criar Conta
// @Description Creates a new ledger account.
// @Tags Ledger
// @Accept json
// @Produce json
// @Param payload body dto.CreateLedgerAccountRequest true "Account Payload"
// @Success 201 {object} dto.LedgerAccountResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Router /ledger/accounts [post]
func (h *LedgerHandler) CreateAccount(c echo.Context) error {
	var req dto.CreateLedgerAccountRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid payload"})
	}

	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	created, err := h.service.CreateAccount(c.Request().Context(), req.Name, req.Type, req.Currency, isActive)
	if err != nil {
		if isLedgerValidationError(err) {
			return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		}
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "failed to create account"})
	}

	return c.JSON(http.StatusCreated, toLedgerAccountResponse(created))
}

// ListAccounts lists ledger accounts.
// @Summary Listar Contas
// @Description Returns all ledger accounts.
// @Tags Ledger
// @Accept json
// @Produce json
// @Success 200 {array} dto.LedgerAccountResponse
// @Failure 401 {object} dto.ErrorResponse
// @Router /ledger/accounts [get]
func (h *LedgerHandler) ListAccounts(c echo.Context) error {
	list, err := h.service.ListAccounts(c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "failed to list accounts"})
	}

	resp := make([]dto.LedgerAccountResponse, len(list))
	for i, acc := range list {
		resp[i] = toLedgerAccountResponse(acc)
	}

	return c.JSON(http.StatusOK, resp)
}

// CreateTransaction creates a new ledger transaction.
// @Summary Criar Transacao
// @Description Creates a new ledger transaction with postings.
// @Tags Ledger
// @Accept json
// @Produce json
// @Param payload body dto.CreateLedgerTransactionRequest true "Transaction Payload"
// @Success 201 {object} dto.LedgerTransactionResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Router /ledger/transactions [post]
func (h *LedgerHandler) CreateTransaction(c echo.Context) error {
	var req dto.CreateLedgerTransactionRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid payload"})
	}

	occurredAt, err := time.Parse("2006-01-02", req.OccurredAt)
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid occurred_at format, use YYYY-MM-DD"})
	}

	postings := make([]*ledger.Posting, len(req.Postings))
	for i, posting := range req.Postings {
		postings[i] = &ledger.Posting{
			AccountID:  posting.AccountID,
			CategoryID: posting.CategoryID,
			PartyID:    posting.PartyID,
			Side:       posting.Side,
			Amount:     posting.Amount,
			Memo:       posting.Memo,
		}
	}

	created, err := h.service.CreateTransaction(
		c.Request().Context(),
		occurredAt,
		req.Description,
		req.Reference,
		req.Notes,
		postings,
	)
	if err != nil {
		if isLedgerValidationError(err) {
			return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		}
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: fmt.Sprintf("failed to create transaction: %v", err)})
	}

	return c.JSON(http.StatusCreated, toLedgerTransactionResponse(created))
}

// ListTransactionsByMonth lists transactions for a month.
// @Summary Listar Transacoes por Mes
// @Description Returns ledger transactions for a reference month.
// @Tags Ledger
// @Accept json
// @Produce json
// @Param month query string true "Reference Month (YYYY-MM-DD)" format(date) example(2024-03-01)
// @Success 200 {array} dto.LedgerTransactionResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Router /ledger/transactions [get]
func (h *LedgerHandler) ListTransactionsByMonth(c echo.Context) error {
	monthStr := c.QueryParam("month")
	if monthStr == "" {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "month parameter is required (YYYY-MM-DD)"})
	}

	month, err := time.Parse("2006-01-02", monthStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid month format, use YYYY-MM-DD"})
	}

	list, err := h.service.ListTransactionsByMonth(c.Request().Context(), month)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "failed to list transactions"})
	}

	resp := make([]dto.LedgerTransactionResponse, len(list))
	for i, tx := range list {
		resp[i] = toLedgerTransactionResponse(tx)
	}

	return c.JSON(http.StatusOK, resp)
}

func RegisterLedgerRoutes(e *echo.Echo, h *LedgerHandler) {
	g := e.Group("/ledger")
	g.POST("/accounts", h.CreateAccount)
	g.GET("/accounts", h.ListAccounts)
	g.POST("/transactions", h.CreateTransaction)
	g.GET("/transactions", h.ListTransactionsByMonth)
}

func toLedgerAccountResponse(account *ledger.Account) dto.LedgerAccountResponse {
	return dto.LedgerAccountResponse{
		ID:       account.ID,
		Name:     account.Name,
		Type:     account.Type,
		Currency: account.Currency,
		IsActive: account.IsActive,
	}
}

func toLedgerTransactionResponse(tx *ledger.Transaction) dto.LedgerTransactionResponse {
	postings := make([]dto.LedgerPostingResponse, len(tx.Postings))
	for i, posting := range tx.Postings {
		postings[i] = dto.LedgerPostingResponse{
			ID:         posting.ID,
			AccountID:  posting.AccountID,
			CategoryID: posting.CategoryID,
			PartyID:    posting.PartyID,
			Side:       posting.Side,
			Amount:     posting.Amount,
			Memo:       posting.Memo,
		}
	}

	return dto.LedgerTransactionResponse{
		ID:          tx.ID,
		OccurredAt:  tx.OccurredAt.Format("2006-01-02"),
		Description: tx.Description,
		Reference:   tx.Reference,
		Notes:       tx.Notes,
		Postings:    postings,
	}
}

func isLedgerValidationError(err error) bool {
	return errors.Is(err, ledger.ErrInvalidAccountName) ||
		errors.Is(err, ledger.ErrInvalidAccountType) ||
		errors.Is(err, ledger.ErrInvalidCurrency) ||
		errors.Is(err, ledger.ErrInvalidTransaction) ||
		errors.Is(err, ledger.ErrEmptyDescription) ||
		errors.Is(err, ledger.ErrNoPostings) ||
		errors.Is(err, ledger.ErrInvalidPostingSide) ||
		errors.Is(err, ledger.ErrInvalidPostingAmount) ||
		errors.Is(err, ledger.ErrInvalidPostingAccount) ||
		errors.Is(err, ledger.ErrUnbalancedTransaction)
}
