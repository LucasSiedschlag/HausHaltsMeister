package dto

type CreateCashFlowRequest struct {
	Date       string  `json:"date"` // YYYY-MM-DD
	CategoryID int32   `json:"category_id"`
	Direction  string  `json:"direction"`
	Title      string  `json:"title"`
	Amount     float64 `json:"amount"`
	IsFixed    bool    `json:"is_fixed"`
}

type CashFlowResponse struct {
	ID         int32   `json:"id"`
	Date       string  `json:"date"`
	CategoryID int32   `json:"category_id"`
	Direction  string  `json:"direction"`
	Title      string  `json:"title"`
	Amount     float64 `json:"amount"`
	IsFixed    bool    `json:"is_fixed"`
}

type MonthlySummaryResponse struct {
	TotalIncome  float64 `json:"total_income"`
	TotalExpense float64 `json:"total_expense"`
	Balance      float64 `json:"balance"`
}

type CategorySummaryResponse struct {
	CategoryName string  `json:"category_name"`
	Direction    string  `json:"direction"`
	TotalAmount  float64 `json:"total_amount"`
}

type CopyFixedRequest struct {
	FromMonth string `json:"from_month"`
	ToMonth   string `json:"to_month"`
}
