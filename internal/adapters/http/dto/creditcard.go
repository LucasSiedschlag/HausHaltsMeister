package dto

import "time"

type CardNetworkRequest struct {
	Code        string `json:"code"`
	DisplayName string `json:"display_name"`
}

type CardNetworkResponse struct {
	Code        string     `json:"code"`
	DisplayName string     `json:"display_name"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   *time.Time `json:"updated_at,omitempty"`
}

type CreditCardRequest struct {
	Label      *string `json:"label"`
	Brand      string  `json:"brand"`
	Last4      *string `json:"last4"`
	CVV        *string `json:"cvv"`
	HolderName *string `json:"holder_name"`
	Active     *bool   `json:"active"`
	Color      *string `json:"color"`
	Style      *string `json:"style"`
	ClosingDay int     `json:"closing_day"`
	DueDay     int     `json:"due_day"`
}

type CreditCardResponse struct {
	ID                 string     `json:"id"`
	LedgerID           string     `json:"ledger_id"`
	ParentAccountID    string     `json:"parent_account_id"`
	LiabilityAccountID string     `json:"liability_account_id"`
	Label              *string    `json:"label,omitempty"`
	Brand              string     `json:"brand"`
	Last4              *string    `json:"last4,omitempty"`
	CVV                *string    `json:"cvv,omitempty"`
	HolderName         *string    `json:"holder_name,omitempty"`
	Active             bool       `json:"active"`
	Color              *string    `json:"color,omitempty"`
	Style              *string    `json:"style,omitempty"`
	ClosingDay         int        `json:"closing_day"`
	DueDay             int        `json:"due_day"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          *time.Time `json:"updated_at,omitempty"`
}

type InstallmentPlanRequest struct {
	PurchaseOccurredAt     string  `json:"purchase_occurred_at"`
	Merchant               *string `json:"merchant"`
	Description            string  `json:"description"`
	CategoryID             string  `json:"category_id"`
	TotalAmountCents       int64   `json:"total_amount_cents"`
	InstallmentsCount      int     `json:"installments_count"`
	InstallmentAmountCents int64   `json:"installment_amount_cents"`
	FirstDueMonth          string  `json:"first_due_month"`
}

type InstallmentPlanPatchRequest struct {
	Status string `json:"status"`
}

type InstallmentPlanResponse struct {
	ID                     string     `json:"id"`
	LedgerID               string     `json:"ledger_id"`
	CreditCardID           string     `json:"credit_card_id"`
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

type InstallmentResponse struct {
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

type InstallmentPatchRequest struct {
	Status string `json:"status"`
}

type PostingResponse struct {
	PostedCount    int      `json:"posted_count"`
	TransactionIDs []string `json:"transaction_ids"`
}

type StatementResponse struct {
	ID                   string     `json:"id"`
	LedgerID             string     `json:"ledger_id"`
	CreditCardID         string     `json:"credit_card_id"`
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

type StatementPayRequest struct {
	StatementID     string `json:"statement_id"`
	PaymentDate     string `json:"payment_date"`
	PayAmountCents  int64  `json:"pay_amount_cents"`
	PayingAccountID string `json:"paying_account_id"`
}
