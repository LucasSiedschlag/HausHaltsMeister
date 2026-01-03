package dto

import "time"

type InvestmentRequest struct {
	AmountCents int64   `json:"amount_cents"`
	OccurredAt  string  `json:"occurred_at"`
	Memo        *string `json:"memo"`
}

type InvestmentSummaryResponse struct {
	LedgerID           string    `json:"ledger_id"`
	From               time.Time `json:"from"`
	To                 time.Time `json:"to"`
	TotalContributions int64     `json:"total_contributions"`
	TotalRedemptions   int64     `json:"total_redemptions"`
	TotalEarnings      int64     `json:"total_earnings"`
	TotalLosses        int64     `json:"total_losses"`
	NetVariation       int64     `json:"net_variation"`
}
