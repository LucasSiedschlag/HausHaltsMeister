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
	Month string               `json:"month"`
	Items []BudgetItemResponse `json:"items"`
}

type SetBudgetBatchRequest struct {
	StartMonth    string  `json:"start_month"`
	EndMonth      string  `json:"end_month"`
	CategoryID    int32   `json:"category_id"`
	PlannedAmount float64 `json:"planned_amount"`
}
