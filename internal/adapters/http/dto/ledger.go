package dto

import "time"

type LedgerCreateRequest struct {
	Name                  string `json:"name"`
	CurrencyCode          string `json:"currency_code"`
	CreateDefaultAccounts *bool  `json:"create_default_accounts"`
	IncludeInvestment     *bool  `json:"include_investment"`
}

type LedgerUpdateRequest struct {
	Name string `json:"name"`
}

type LedgerMemberRequest struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
}

type LedgerResponse struct {
	ID           string     `json:"id"`
	OwnerUserID  string     `json:"owner_user_id"`
	Name         string     `json:"name"`
	CurrencyCode string     `json:"currency_code"`
	Role         string     `json:"role,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    *time.Time `json:"updated_at,omitempty"`
}

type LedgerMemberResponse struct {
	LedgerID    string     `json:"ledger_id"`
	UserID      string     `json:"user_id"`
	Role        string     `json:"role"`
	DisplayName string     `json:"display_name"`
	Email       string     `json:"email"`
	AvatarURL   *string    `json:"avatar_url,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   *time.Time `json:"updated_at,omitempty"`
}

type LedgerMeResponse struct {
	LedgerID string `json:"ledger_id"`
	Role     string `json:"role"`
}
