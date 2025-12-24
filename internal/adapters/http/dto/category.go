package dto

type CreateCategoryRequest struct {
	Name             string `json:"name"`
	Direction        string `json:"direction"` // IN or OUT
	IsBudgetRelevant bool   `json:"is_budget_relevant"`
}

type CategoryResponse struct {
	ID               int32  `json:"id"`
	Name             string `json:"name"`
	Direction        string `json:"direction"`
	IsBudgetRelevant bool   `json:"is_budget_relevant"`
	IsActive         bool   `json:"is_active"`
}
