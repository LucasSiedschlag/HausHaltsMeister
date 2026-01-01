package httpapi

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/LucasSiedschlag/HausHaltsMeister/internal/domain/journal"
	"github.com/labstack/echo/v4"
)

type JournalHandler struct {
	Service JournalService
}

type JournalService interface {
	CreateTransaction(ctx context.Context, userID, ledgerID string, params journal.CreateTransactionParams) (journal.Transaction, error)
	ListTransactions(ctx context.Context, userID, ledgerID string, params journal.ListTransactionsParams) (journal.ListResult, error)
	GetTransaction(ctx context.Context, userID, ledgerID, transactionID string) (journal.Transaction, error)
	UpdateTransaction(ctx context.Context, userID, ledgerID, transactionID string, params journal.UpdateTransactionParams) (journal.Transaction, error)
	DeleteTransaction(ctx context.Context, userID, ledgerID, transactionID string) error
}

type transactionRequest struct {
	OccurredAt  string             `json:"occurred_at"`
	Description string             `json:"description"`
	Notes       *string            `json:"notes"`
	Entries     []transactionEntry `json:"entries"`
}

type transactionEntry struct {
	AccountID   string  `json:"account_id"`
	CategoryID  *string `json:"category_id"`
	Kind        string  `json:"kind"`
	AmountCents int64   `json:"amount_cents"`
	Memo        *string `json:"memo"`
}

type transactionPatchRequest struct {
	OccurredAt  *string             `json:"occurred_at"`
	Description *string             `json:"description"`
	Notes       *string             `json:"notes"`
	Entries     *[]transactionEntry `json:"entries"`
}

type transactionResponse struct {
	ID          string                     `json:"id"`
	LedgerID    string                     `json:"ledger_id"`
	OccurredAt  time.Time                  `json:"occurred_at"`
	Description string                     `json:"description"`
	Notes       *string                    `json:"notes,omitempty"`
	CreatedAt   time.Time                  `json:"created_at"`
	UpdatedAt   *time.Time                 `json:"updated_at,omitempty"`
	Entries     []transactionEntryResponse `json:"entries"`
}

type transactionEntryResponse struct {
	ID          string  `json:"id"`
	AccountID   string  `json:"account_id"`
	CategoryID  *string `json:"category_id,omitempty"`
	Kind        string  `json:"kind"`
	AmountCents int64   `json:"amount_cents"`
	Memo        *string `json:"memo,omitempty"`
}

type transactionListResponse struct {
	Items      []transactionResponse `json:"items"`
	NextCursor *transactionCursor    `json:"next_cursor,omitempty"`
}

type transactionCursor struct {
	CursorOccurredAt time.Time `json:"cursor_occurred_at"`
	CursorID         string    `json:"cursor_id"`
}

func (h *JournalHandler) Register(g *echo.Group) {
	g.POST("", h.Create)
	g.GET("", h.List)
	g.GET("/:transactionId", h.Get)
	g.PATCH("/:transactionId", h.Update)
	g.DELETE("/:transactionId", h.Delete)
}

func (h *JournalHandler) Create(c echo.Context) error {
	user, ok := GetUser(c)
	if !ok {
		return WriteError(c, http.StatusUnauthorized, "AUTH_INVALID_CREDENTIALS", "Credenciais invalidas", nil)
	}
	ledgerID := c.Param("ledgerId")
	if ledgerID == "" {
		return WriteError(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Parametros invalidos", map[string]string{"ledger_id": "required"})
	}

	var req transactionRequest
	if err := c.Bind(&req); err != nil {
		return WriteError(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Payload invalido", nil)
	}

	occurredAt, err := parseDateTime(req.OccurredAt)
	if err != nil {
		return WriteError(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Parametros invalidos", map[string]string{"occurred_at": "invalid"})
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

	created, err := h.Service.CreateTransaction(c.Request().Context(), user.ID, ledgerID, journal.CreateTransactionParams{
		LedgerID:    ledgerID,
		OccurredAt:  occurredAt,
		Description: req.Description,
		Notes:       req.Notes,
		Entries:     entries,
	})
	if err != nil {
		return WriteAppError(c, err)
	}

	return c.JSON(http.StatusCreated, toTransactionResponse(created))
}

func (h *JournalHandler) List(c echo.Context) error {
	user, ok := GetUser(c)
	if !ok {
		return WriteError(c, http.StatusUnauthorized, "AUTH_INVALID_CREDENTIALS", "Credenciais invalidas", nil)
	}
	ledgerID := c.Param("ledgerId")
	if ledgerID == "" {
		return WriteError(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Parametros invalidos", map[string]string{"ledger_id": "required"})
	}

	params := journal.ListTransactionsParams{}
	if value := strings.TrimSpace(c.QueryParam("from")); value != "" {
		parsed, err := parseDateTime(value)
		if err != nil {
			return WriteError(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Parametros invalidos", map[string]string{"from": "invalid"})
		}
		params.From = &parsed
	}
	if value := strings.TrimSpace(c.QueryParam("to")); value != "" {
		parsed, err := parseDateTime(value)
		if err != nil {
			return WriteError(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Parametros invalidos", map[string]string{"to": "invalid"})
		}
		params.To = &parsed
	}
	if value := strings.TrimSpace(c.QueryParam("account_id")); value != "" {
		params.AccountID = &value
	}
	if value := strings.TrimSpace(c.QueryParam("category_id")); value != "" {
		params.CategoryID = &value
	}
	if value := strings.TrimSpace(c.QueryParam("q")); value != "" {
		params.Query = &value
	}
	if value := strings.TrimSpace(c.QueryParam("limit")); value != "" {
		limit, err := strconv.Atoi(value)
		if err != nil {
			return WriteError(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Parametros invalidos", map[string]string{"limit": "invalid"})
		}
		params.Limit = limit
	}

	cursorOccurredAt := strings.TrimSpace(c.QueryParam("cursor_occurred_at"))
	cursorID := strings.TrimSpace(c.QueryParam("cursor_id"))
	if cursorOccurredAt != "" && cursorID != "" {
		parsed, err := parseDateTime(cursorOccurredAt)
		if err != nil {
			return WriteError(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Parametros invalidos", map[string]string{"cursor_occurred_at": "invalid"})
		}
		params.Cursor = &journal.Cursor{OccurredAt: parsed, ID: cursorID}
	}

	result, err := h.Service.ListTransactions(c.Request().Context(), user.ID, ledgerID, params)
	if err != nil {
		return WriteAppError(c, err)
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
	user, ok := GetUser(c)
	if !ok {
		return WriteError(c, http.StatusUnauthorized, "AUTH_INVALID_CREDENTIALS", "Credenciais invalidas", nil)
	}
	ledgerID := c.Param("ledgerId")
	transactionID := c.Param("transactionId")
	if ledgerID == "" || transactionID == "" {
		return WriteError(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Parametros invalidos", map[string]string{"transaction_id": "required"})
	}

	item, err := h.Service.GetTransaction(c.Request().Context(), user.ID, ledgerID, transactionID)
	if err != nil {
		return WriteAppError(c, err)
	}
	return c.JSON(http.StatusOK, toTransactionResponse(item))
}

func (h *JournalHandler) Update(c echo.Context) error {
	user, ok := GetUser(c)
	if !ok {
		return WriteError(c, http.StatusUnauthorized, "AUTH_INVALID_CREDENTIALS", "Credenciais invalidas", nil)
	}
	ledgerID := c.Param("ledgerId")
	transactionID := c.Param("transactionId")
	if ledgerID == "" || transactionID == "" {
		return WriteError(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Parametros invalidos", map[string]string{"transaction_id": "required"})
	}

	var req transactionPatchRequest
	if err := c.Bind(&req); err != nil {
		return WriteError(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Payload invalido", nil)
	}

	var occurredAt *time.Time
	if req.OccurredAt != nil {
		parsed, err := parseDateTime(*req.OccurredAt)
		if err != nil {
			return WriteError(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Parametros invalidos", map[string]string{"occurred_at": "invalid"})
		}
		occurredAt = &parsed
	}

	var entries *[]journal.EntryInput
	if req.Entries != nil {
		if len(*req.Entries) == 0 {
			return WriteError(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Parametros invalidos", map[string]string{"entries": "required"})
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
		return WriteAppError(c, err)
	}

	return c.JSON(http.StatusOK, toTransactionResponse(updated))
}

func (h *JournalHandler) Delete(c echo.Context) error {
	user, ok := GetUser(c)
	if !ok {
		return WriteError(c, http.StatusUnauthorized, "AUTH_INVALID_CREDENTIALS", "Credenciais invalidas", nil)
	}
	ledgerID := c.Param("ledgerId")
	transactionID := c.Param("transactionId")
	if ledgerID == "" || transactionID == "" {
		return WriteError(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Parametros invalidos", map[string]string{"transaction_id": "required"})
	}

	if err := h.Service.DeleteTransaction(c.Request().Context(), user.ID, ledgerID, transactionID); err != nil {
		return WriteAppError(c, err)
	}
	return c.NoContent(http.StatusNoContent)
}

func parseDateTime(value string) (time.Time, error) {
	if strings.Contains(value, "T") {
		return time.Parse(time.RFC3339, value)
	}
	return time.Parse("2006-01-02", value)
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
		ID:          item.ID,
		LedgerID:    item.LedgerID,
		OccurredAt:  item.OccurredAt,
		Description: item.Description,
		Notes:       item.Notes,
		CreatedAt:   item.CreatedAt,
		UpdatedAt:   item.UpdatedAt,
		Entries:     entries,
	}
}
