package dto

type CreateInstallmentRequest struct {
	Description     string  `json:"description"`
	TotalAmount     float64 `json:"total_amount"`
	Count           int32   `json:"count"`
	CategoryID      int32   `json:"category_id"`
	PaymentMethodID int32   `json:"payment_method_id"`
	PurchaseDate    string  `json:"purchase_date"` // YYYY-MM-DD
}

type InstallmentPlanResponse struct {
	ID                int32   `json:"id"`
	Description       string  `json:"description"`
	TotalAmount       float64 `json:"total_amount"`
	InstallmentCount  int32   `json:"installment_count"`
	InstallmentAmount float64 `json:"installment_amount"`
	StartMonth        string  `json:"start_month"`
	PaymentMethodID   int32   `json:"payment_method_id"`
}
