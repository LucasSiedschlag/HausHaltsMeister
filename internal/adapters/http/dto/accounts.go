package dto

import "time"

type AccountRequest struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Nature   string `json:"nature"`
	IsActive *bool  `json:"is_active"`
}

type AccountResponse struct {
	ID        string     `json:"id"`
	LedgerID  string     `json:"ledger_id"`
	Name      string     `json:"name"`
	Type      string     `json:"type"`
	Nature    string     `json:"nature"`
	IsActive  bool       `json:"is_active"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt *time.Time `json:"updated_at,omitempty"`
}
