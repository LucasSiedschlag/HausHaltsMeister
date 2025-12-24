package dto

type SetBudgetItemRequest struct {
	CategoryID    int32   `json:"category_id"`
	PlannedAmount float64 `json:"planned_amount"`
}

type BudgetItemResponse struct {
	ID             int32   `json:"id"`
	BudgetPeriodID int32   `json:"budget_period_id"`
	CategoryID     int32   `json:"category_id"`
	CategoryName   string  `json:"category_name,omitempty"`
	Mode           string  `json:"mode"`
	PlannedAmount  float64 `json:"planned_amount"`
	ActualAmount   float64 `json:"actual_amount"`
}

type BudgetSummaryResponse struct {
	Month        string               `json:"month"`
	TotalPlanned float64              `json:"total_planned"`
	TotalActual  float64              `json:"total_actual"`
	Items        []BudgetItemResponse `json:"items"`
}
