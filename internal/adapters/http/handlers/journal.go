package handlers

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/LucasSiedschlag/HausHaltsMeister/internal/adapters/http/dto"
	"github.com/LucasSiedschlag/HausHaltsMeister/internal/adapters/http/httpx"
	"github.com/LucasSiedschlag/HausHaltsMeister/internal/domain/audit"
	"github.com/LucasSiedschlag/HausHaltsMeister/internal/domain/journal"
	"github.com/labstack/echo/v4"
)

type JournalHandler struct {
	Service JournalService
	Audit   audit.Recorder
}

type JournalService interface {
	CreateTransaction(ctx context.Context, userID, ledgerID string, params journal.CreateTransactionParams) (journal.Transaction, error)
	ListTransactions(ctx context.Context, userID, ledgerID string, params journal.ListTransactionsParams) (journal.ListResult, error)
	GetTransaction(ctx context.Context, userID, ledgerID, transactionID string) (journal.Transaction, error)
	UpdateTransaction(ctx context.Context, userID, ledgerID, transactionID string, params journal.UpdateTransactionParams) (journal.Transaction, error)
	DeleteTransaction(ctx context.Context, userID, ledgerID, transactionID string) error
}

type transactionRequest = dto.TransactionRequest
type transactionEntry = dto.TransactionEntry
type transactionPatchRequest = dto.TransactionPatchRequest
type transactionResponse = dto.TransactionResponse
type transactionEntryResponse = dto.TransactionEntryResponse
type transactionListResponse = dto.TransactionListResponse
type transactionCursor = dto.TransactionCursor

func (h *JournalHandler) Register(g *echo.Group) {
	g.POST("", h.Create)
	g.GET("", h.List)
	g.GET("/:transactionId", h.Get)
	g.PATCH("/:transactionId", h.Update)
	g.DELETE("/:transactionId", h.Delete)
}

func (h *JournalHandler) Create(c echo.Context) error {
	user, ok := httpx.GetUser(c)
	if !ok {
		return httpx.WriteError(c, http.StatusUnauthorized, "AUTH_INVALID_CREDENTIALS", "Credenciais invalidas", nil)
	}
	ledgerID, err := httpx.RequireUUIDParam(c, "ledgerId")
	if err != nil {
		return err
	}

	var req transactionRequest
	if err := c.Bind(&req); err != nil {
		return httpx.WriteError(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Payload invalido", nil)
	}

	occurredAt, err := httpx.ParseDateTime(req.OccurredAt)
	if err != nil {
		return httpx.WriteError(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Parametros invalidos", map[string]string{"occurred_at": "invalid"})
	}

	entries := make([]journal.EntryInput, 0, len(req.Entries))
	for _, entry := range req.Entries {
		entries = append(entries, journal.EntryInput{
			AccountID:   entry.AccountID,
			CategoryID:  entry.CategoryID,
			Kind:        entry.Kind,
			AmountCents: entry.AmountCents,
			Memo:        entry.Memo,
		})
	}

	var idempotency *journal.IdempotencyParams
	if value := strings.TrimSpace(c.Request().Header.Get("Idempotency-Key")); value != "" {
		idempotency = &journal.IdempotencyParams{Key: value}
	}

	created, err := h.Service.CreateTransaction(c.Request().Context(), user.ID, ledgerID, journal.CreateTransactionParams{
		LedgerID:         ledgerID,
		OccurredAt:       occurredAt,
		Description:      req.Description,
		Notes:            req.Notes,
		InvestmentAction: req.InvestmentAction,
		Entries:          entries,
		Idempotency:      idempotency,
	})
	if err != nil {
		return httpx.WriteAppError(c, err)
	}

	recordAudit(c, h.Audit, audit.Event{
		LedgerID:  ledgerID,
		UserID:    user.ID,
		Action:    "journal.transaction.create",
		EntityID:  &created.ID,
		IP:        c.RealIP(),
		UserAgent: c.Request().UserAgent(),
	})

	return c.JSON(http.StatusCreated, toTransactionResponse(created))
}

func (h *JournalHandler) List(c echo.Context) error {
	user, ok := httpx.GetUser(c)
	if !ok {
		return httpx.WriteError(c, http.StatusUnauthorized, "AUTH_INVALID_CREDENTIALS", "Credenciais invalidas", nil)
	}
	ledgerID, err := httpx.RequireUUIDParam(c, "ledgerId")
	if err != nil {
		return err
	}

	params := journal.ListTransactionsParams{}
	if value := strings.TrimSpace(c.QueryParam("from")); value != "" {
		parsed, err := httpx.ParseDateTime(value)
		if err != nil {
			return httpx.WriteError(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Parametros invalidos", map[string]string{"from": "invalid"})
		}
		params.From = &parsed
	}
	if value := strings.TrimSpace(c.QueryParam("to")); value != "" {
		parsed, err := httpx.ParseDateTime(value)
		if err != nil {
			return httpx.WriteError(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Parametros invalidos", map[string]string{"to": "invalid"})
		}
		params.To = &parsed
	}
	if value := strings.TrimSpace(c.QueryParam("account_id")); value != "" {
		if !httpx.IsUUID(value) {
			return httpx.WriteError(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Parametros invalidos", map[string]string{"account_id": "invalid"})
		}
		params.AccountID = &value
	}
	if value := strings.TrimSpace(c.QueryParam("category_id")); value != "" {
		if !httpx.IsUUID(value) {
			return httpx.WriteError(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Parametros invalidos", map[string]string{"category_id": "invalid"})
		}
		params.CategoryID = &value
	}
	if value := strings.TrimSpace(c.QueryParam("q")); value != "" {
		params.Query = &value
	}
	if value := strings.TrimSpace(c.QueryParam("limit")); value != "" {
		limit, err := strconv.Atoi(value)
		if err != nil {
			return httpx.WriteError(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Parametros invalidos", map[string]string{"limit": "invalid"})
		}
		params.Limit = limit
	}

	cursorOccurredAt := strings.TrimSpace(c.QueryParam("cursor_occurred_at"))
	cursorID := strings.TrimSpace(c.QueryParam("cursor_id"))
	if cursorOccurredAt != "" && cursorID != "" {
		if !httpx.IsUUID(cursorID) {
			return httpx.WriteError(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Parametros invalidos", map[string]string{"cursor_id": "invalid"})
		}
		parsed, err := httpx.ParseDateTime(cursorOccurredAt)
		if err != nil {
			return httpx.WriteError(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Parametros invalidos", map[string]string{"cursor_occurred_at": "invalid"})
		}
		params.Cursor = &journal.Cursor{OccurredAt: parsed, ID: cursorID}
	}

	result, err := h.Service.ListTransactions(c.Request().Context(), user.ID, ledgerID, params)
	if err != nil {
		return httpx.WriteAppError(c, err)
	}

	items := make([]transactionResponse, 0, len(result.Items))
	for _, item := range result.Items {
		items = append(items, toTransactionResponse(item))
	}

	var next *transactionCursor
	if result.NextCursor != nil {
		next = &transactionCursor{CursorOccurredAt: result.NextCursor.OccurredAt, CursorID: result.NextCursor.ID}
	}

	return c.JSON(http.StatusOK, transactionListResponse{Items: items, NextCursor: next})
}

func (h *JournalHandler) Get(c echo.Context) error {
	user, ok := httpx.GetUser(c)
	if !ok {
		return httpx.WriteError(c, http.StatusUnauthorized, "AUTH_INVALID_CREDENTIALS", "Credenciais invalidas", nil)
	}
	ledgerID, err := httpx.RequireUUIDParam(c, "ledgerId")
	if err != nil {
		return err
	}
	transactionID, err := httpx.RequireUUIDParam(c, "transactionId")
	if err != nil {
		return err
	}

	item, err := h.Service.GetTransaction(c.Request().Context(), user.ID, ledgerID, transactionID)
	if err != nil {
		return httpx.WriteAppError(c, err)
	}
	return c.JSON(http.StatusOK, toTransactionResponse(item))
}

func (h *JournalHandler) Update(c echo.Context) error {
	user, ok := httpx.GetUser(c)
	if !ok {
		return httpx.WriteError(c, http.StatusUnauthorized, "AUTH_INVALID_CREDENTIALS", "Credenciais invalidas", nil)
	}
	ledgerID, err := httpx.RequireUUIDParam(c, "ledgerId")
	if err != nil {
		return err
	}
	transactionID, err := httpx.RequireUUIDParam(c, "transactionId")
	if err != nil {
		return err
	}

	var req transactionPatchRequest
	if err := c.Bind(&req); err != nil {
		return httpx.WriteError(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Payload invalido", nil)
	}

	var occurredAt *time.Time
	if req.OccurredAt != nil {
		parsed, err := httpx.ParseDateTime(*req.OccurredAt)
		if err != nil {
			return httpx.WriteError(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Parametros invalidos", map[string]string{"occurred_at": "invalid"})
		}
		occurredAt = &parsed
	}

	var entries *[]journal.EntryInput
	if req.Entries != nil {
		if len(*req.Entries) == 0 {
			return httpx.WriteError(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Parametros invalidos", map[string]string{"entries": "required"})
		}
		parsedEntries := make([]journal.EntryInput, 0, len(*req.Entries))
		for _, entry := range *req.Entries {
			parsedEntries = append(parsedEntries, journal.EntryInput{
				AccountID:   entry.AccountID,
				CategoryID:  entry.CategoryID,
				Kind:        entry.Kind,
				AmountCents: entry.AmountCents,
				Memo:        entry.Memo,
			})
		}
		entries = &parsedEntries
	}

	updated, err := h.Service.UpdateTransaction(c.Request().Context(), user.ID, ledgerID, transactionID, journal.UpdateTransactionParams{
		OccurredAt:  occurredAt,
		Description: req.Description,
		Notes:       req.Notes,
		Entries:     entries,
	})
	if err != nil {
		return httpx.WriteAppError(c, err)
	}

	recordAudit(c, h.Audit, audit.Event{
		LedgerID:  ledgerID,
		UserID:    user.ID,
		Action:    "journal.transaction.update",
		EntityID:  &updated.ID,
		IP:        c.RealIP(),
		UserAgent: c.Request().UserAgent(),
	})

	return c.JSON(http.StatusOK, toTransactionResponse(updated))
}

func (h *JournalHandler) Delete(c echo.Context) error {
	user, ok := httpx.GetUser(c)
	if !ok {
		return httpx.WriteError(c, http.StatusUnauthorized, "AUTH_INVALID_CREDENTIALS", "Credenciais invalidas", nil)
	}
	ledgerID, err := httpx.RequireUUIDParam(c, "ledgerId")
	if err != nil {
		return err
	}
	transactionID, err := httpx.RequireUUIDParam(c, "transactionId")
	if err != nil {
		return err
	}

	if err := h.Service.DeleteTransaction(c.Request().Context(), user.ID, ledgerID, transactionID); err != nil {
		return httpx.WriteAppError(c, err)
	}

	recordAudit(c, h.Audit, audit.Event{
		LedgerID:  ledgerID,
		UserID:    user.ID,
		Action:    "journal.transaction.delete",
		EntityID:  &transactionID,
		IP:        c.RealIP(),
		UserAgent: c.Request().UserAgent(),
	})
	return c.NoContent(http.StatusNoContent)
}

func toTransactionResponse(item journal.Transaction) transactionResponse {
	entries := make([]transactionEntryResponse, 0, len(item.Entries))
	for _, entry := range item.Entries {
		entries = append(entries, transactionEntryResponse{
			ID:          entry.ID,
			AccountID:   entry.AccountID,
			CategoryID:  entry.CategoryID,
			Kind:        entry.Kind,
			AmountCents: entry.AmountCents,
			Memo:        entry.Memo,
		})
	}

	return transactionResponse{
		ID:               item.ID,
		LedgerID:         item.LedgerID,
		OccurredAt:       item.OccurredAt,
		Description:      item.Description,
		Notes:            item.Notes,
		CreditCardID:     item.CreditCardID,
		InvestmentAction: item.InvestmentAction,
		CreatedAt:        item.CreatedAt,
		UpdatedAt:        item.UpdatedAt,
		Entries:          entries,
	}
}
