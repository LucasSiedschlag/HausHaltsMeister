package dto

type CreatePaymentMethodRequest struct {
	Name       string `json:"name"`
	Kind       string `json:"kind"`
	BankName   string `json:"bank_name"`
	ClosingDay *int32 `json:"closing_day"`
	DueDay     *int32 `json:"due_day"`
}

type UpdatePaymentMethodRequest struct {
	Name       *string `json:"name"`
	Kind       *string `json:"kind"`
	BankName   *string `json:"bank_name"`
	ClosingDay *int32  `json:"closing_day"`
	DueDay     *int32  `json:"due_day"`
	IsActive   *bool   `json:"is_active"`
}

type PaymentMethodResponse struct {
	ID         int32  `json:"id"`
	Name       string `json:"name"`
	Kind       string `json:"kind"`
	BankName   string `json:"bank_name"`
	ClosingDay *int32 `json:"closing_day"`
	DueDay     *int32 `json:"due_day"`
	IsActive   bool   `json:"is_active"`
}

type InvoiceEntryResponse struct {
	CashFlowID   int32   `json:"cash_flow_id"`
	Date         string  `json:"date"`
	Title        string  `json:"title"`
	Amount       float64 `json:"amount"`
	CategoryName string  `json:"category_name"`
}

type InvoiceResponse struct {
	PaymentMethodID int32                  `json:"payment_method_id"`
	Month           string                 `json:"month"`
	Total           float64                `json:"total"`
	Entries         []InvoiceEntryResponse `json:"entries"`
}
