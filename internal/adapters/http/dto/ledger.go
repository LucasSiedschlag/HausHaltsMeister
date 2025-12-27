package dto

type CreateLedgerAccountRequest struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Currency string `json:"currency"`
	IsActive *bool  `json:"is_active,omitempty"`
}

type LedgerAccountResponse struct {
	ID       int32  `json:"id"`
	Name     string `json:"name"`
	Type     string `json:"type"`
	Currency string `json:"currency"`
	IsActive bool   `json:"is_active"`
}

type CreateLedgerTransactionRequest struct {
	OccurredAt  string                       `json:"occurred_at"`
	Description string                       `json:"description"`
	Reference   string                       `json:"reference,omitempty"`
	Notes       string                       `json:"notes,omitempty"`
	Postings    []CreateLedgerPostingRequest `json:"postings"`
}

type CreateLedgerPostingRequest struct {
	AccountID  int32   `json:"account_id"`
	CategoryID *int32  `json:"category_id,omitempty"`
	PartyID    *int32  `json:"party_id,omitempty"`
	Side       string  `json:"side"`
	Amount     float64 `json:"amount"`
	Memo       string  `json:"memo,omitempty"`
}

type LedgerTransactionResponse struct {
	ID          int32                   `json:"id"`
	OccurredAt  string                  `json:"occurred_at"`
	Description string                  `json:"description"`
	Reference   string                  `json:"reference,omitempty"`
	Notes       string                  `json:"notes,omitempty"`
	Postings    []LedgerPostingResponse `json:"postings"`
}

type LedgerPostingResponse struct {
	ID         int32   `json:"id"`
	AccountID  int32   `json:"account_id"`
	CategoryID *int32  `json:"category_id,omitempty"`
	PartyID    *int32  `json:"party_id,omitempty"`
	Side       string  `json:"side"`
	Amount     float64 `json:"amount"`
	Memo       string  `json:"memo,omitempty"`
}
