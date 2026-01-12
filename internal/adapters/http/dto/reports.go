package dto

import "time"

type BalanceResponse struct {
	LedgerID string                `json:"ledger_id"`
	Month    time.Time             `json:"month"`
	Items    []BalanceItemResponse `json:"items"`
}

type BalanceItemResponse struct {
	AccountID     string `json:"account_id"`
	AccountName   string `json:"account_name"`
	AccountType   string `json:"account_type"`
	AccountNature string `json:"account_nature"`
	BalanceCents  int64  `json:"balance_cents"`
}

type CategorySummaryResponse struct {
	LedgerID string                        `json:"ledger_id"`
	From     time.Time                     `json:"from"`
	To       time.Time                     `json:"to"`
	Items    []CategorySummaryItemResponse `json:"items"`
}

type CategorySummaryItemResponse struct {
	CategoryID string `json:"category_id"`
	Name       string `json:"name"`
	Direction  string `json:"direction"`
	TotalCents int64  `json:"total_cents"`
}

type CashflowResponse struct {
	LedgerID string                 `json:"ledger_id"`
	From     time.Time              `json:"from"`
	To       time.Time              `json:"to"`
	Items    []CashflowItemResponse `json:"items"`
}

type CashflowItemResponse struct {
	Month         time.Time `json:"month"`
	TotalInCents  int64     `json:"total_in_cents"`
	TotalOutCents int64     `json:"total_out_cents"`
	NetCents      int64     `json:"net_cents"`
}
