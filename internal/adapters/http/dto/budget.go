package dto

type SetBudgetItemRequest struct {
	CategoryID    int32    `json:"category_id"`
	Mode          string   `json:"mode"`
	PlannedAmount *float64 `json:"planned_amount,omitempty"`
	TargetPercent *float64 `json:"target_percent,omitempty"`
}

type UpdateBudgetItemRequest struct {
	Mode          string   `json:"mode"`
	PlannedAmount *float64 `json:"planned_amount,omitempty"`
	TargetPercent *float64 `json:"target_percent,omitempty"`
}

type BudgetItemResponse struct {
	ID             int32   `json:"id"`
	BudgetPeriodID int32   `json:"budget_period_id"`
	CategoryID     int32   `json:"category_id"`
	CategoryName   string  `json:"category_name,omitempty"`
	Mode           string  `json:"mode"`
	PlannedAmount  float64 `json:"planned_amount"`
	ActualAmount   float64 `json:"actual_amount"`
	TargetPercent  float64 `json:"target_percent"`
}

type BudgetSummaryResponse struct {
	Month       string               `json:"month"`
	TotalIncome float64              `json:"total_income"`
	Items       []BudgetItemResponse `json:"items"`
}

type SetBudgetBatchRequest struct {
	StartMonth    string   `json:"start_month"`
	EndMonth      string   `json:"end_month"`
	CategoryID    int32    `json:"category_id"`
	Mode          string   `json:"mode"`
	PlannedAmount *float64 `json:"planned_amount,omitempty"`
	TargetPercent *float64 `json:"target_percent,omitempty"`
}

type BulkBudgetItemRequest struct {
	CategoryID    int32   `json:"category_id"`
	TargetPercent float64 `json:"target_percent"`
}

type BulkBudgetItemsRequest struct {
	Items []BulkBudgetItemRequest `json:"items"`
}
