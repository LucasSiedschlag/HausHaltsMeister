package dto

import "time"

type TransactionRequest struct {
	OccurredAt       string             `json:"occurred_at"`
	Description      string             `json:"description"`
	Notes            *string            `json:"notes"`
	InvestmentAction *string            `json:"investment_action,omitempty"`
	Entries          []TransactionEntry `json:"entries"`
}

type TransactionEntry struct {
	AccountID   string  `json:"account_id"`
	CategoryID  *string `json:"category_id"`
	Kind        string  `json:"kind"`
	AmountCents int64   `json:"amount_cents"`
	Memo        *string `json:"memo"`
}

type TransactionPatchRequest struct {
	OccurredAt  *string             `json:"occurred_at"`
	Description *string             `json:"description"`
	Notes       *string             `json:"notes"`
	Entries     *[]TransactionEntry `json:"entries"`
}

type TransactionResponse struct {
	ID               string                     `json:"id"`
	LedgerID         string                     `json:"ledger_id"`
	OccurredAt       time.Time                  `json:"occurred_at"`
	Description      string                     `json:"description"`
	Notes            *string                    `json:"notes,omitempty"`
	CreditCardID     *string                    `json:"credit_card_id,omitempty"`
	InvestmentAction *string                    `json:"investment_action,omitempty"`
	CreatedAt        time.Time                  `json:"created_at"`
	UpdatedAt        *time.Time                 `json:"updated_at,omitempty"`
	Entries          []TransactionEntryResponse `json:"entries"`
}

type TransactionEntryResponse struct {
	ID          string  `json:"id"`
	AccountID   string  `json:"account_id"`
	CategoryID  *string `json:"category_id,omitempty"`
	Kind        string  `json:"kind"`
	AmountCents int64   `json:"amount_cents"`
	Memo        *string `json:"memo,omitempty"`
}

type TransactionListResponse struct {
	Items      []TransactionResponse `json:"items"`
	NextCursor *TransactionCursor    `json:"next_cursor,omitempty"`
}

type TransactionCursor struct {
	CursorOccurredAt time.Time `json:"cursor_occurred_at"`
	CursorID         string    `json:"cursor_id"`
}
