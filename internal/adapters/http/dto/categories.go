package dto

import "time"

type CategoryRequest struct {
	ParentID         *string `json:"parent_id"`
	Name             string  `json:"name"`
	Direction        string  `json:"direction"`
	IsBudgetBase     *bool   `json:"is_budget_base"`
	IsBudgetRelevant *bool   `json:"is_budget_relevant"`
	IsActive         *bool   `json:"is_active"`
}

type CategoryResponse struct {
	ID               string     `json:"id"`
	LedgerID         string     `json:"ledger_id"`
	ParentID         *string    `json:"parent_id,omitempty"`
	Name             string     `json:"name"`
	Direction        string     `json:"direction"`
	IsBudgetBase     bool       `json:"is_budget_base"`
	IsBudgetRelevant bool       `json:"is_budget_relevant"`
	IsActive         bool       `json:"is_active"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        *time.Time `json:"updated_at,omitempty"`
}
