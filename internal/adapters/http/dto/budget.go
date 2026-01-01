package dto

import "time"

type BudgetPlanRequest struct {
	Name string `json:"name"`
}

type BudgetPlanResponse struct {
	ID        string     `json:"id"`
	LedgerID  string     `json:"ledger_id"`
	Name      string     `json:"name"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt *time.Time `json:"updated_at,omitempty"`
}

type BudgetVersionRequest struct {
	EffectiveFromMonth string                     `json:"effective_from_month"`
	Lines              []BudgetVersionLineRequest `json:"lines"`
}

type BudgetVersionLineRequest struct {
	CategoryID      string  `json:"category_id"`
	Percent         float64 `json:"percent"`
	IncludeChildren bool    `json:"include_children"`
}

type BudgetVersionResponse struct {
	ID                 string               `json:"id"`
	PlanID             string               `json:"plan_id"`
	EffectiveFromMonth time.Time            `json:"effective_from_month"`
	CreatedByUserID    string               `json:"created_by_user_id"`
	CreatedAt          time.Time            `json:"created_at"`
	UpdatedAt          *time.Time           `json:"updated_at,omitempty"`
	Lines              []BudgetLineResponse `json:"lines,omitempty"`
}

type BudgetLineResponse struct {
	ID              string     `json:"id"`
	VersionID       string     `json:"version_id"`
	CategoryID      string     `json:"category_id"`
	Percent         float64    `json:"percent"`
	IncludeChildren bool       `json:"include_children"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       *time.Time `json:"updated_at,omitempty"`
}

type BudgetMonthlyResponse struct {
	Month              time.Time                   `json:"month"`
	IncomeBaseCents    int64                       `json:"income_base_cents"`
	OutsideBudgetCents int64                       `json:"outside_budget_cents"`
	Version            *BudgetVersionResponse      `json:"version,omitempty"`
	Lines              []BudgetMonthlyLineResponse `json:"lines"`
}

type BudgetMonthlyLineResponse struct {
	CategoryID       string  `json:"category_id"`
	Percent          float64 `json:"percent"`
	IncludeChildren  bool    `json:"include_children"`
	BudgetLimitCents int64   `json:"budget_limit_cents"`
	SpentActualCents int64   `json:"spent_actual_cents"`
	DeltaCents       int64   `json:"delta_cents"`
	UsagePct         float64 `json:"usage_pct"`
}
