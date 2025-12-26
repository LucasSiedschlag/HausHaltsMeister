package dto

type CreatePersonRequest struct {
	Name  string `json:"name"`
	Notes string `json:"notes"`
}

type UpdatePersonRequest struct {
	Name  string `json:"name"`
	Notes string `json:"notes"`
}

type PersonResponse struct {
	ID      int32   `json:"id"`
	Name    string  `json:"name"`
	Notes   string  `json:"notes"`
	Balance float64 `json:"balance"`
}

type AddEntryRequest struct {
	PersonID        int32   `json:"person_id"`
	Kind            string  `json:"kind"` // PLUS or MINUS
	Amount          float64 `json:"amount"`
	CashFlowID      *int32  `json:"cash_flow_id"` // Optional
	AutoCreateFlow  bool    `json:"auto_create_flow"`
	PaymentMethodID *int32  `json:"payment_method_id"`
	CardOwner       string  `json:"card_owner"`
}

type UpdateEntryRequest struct {
	PersonID        int32   `json:"person_id"`
	Kind            string  `json:"kind"` // PLUS or MINUS
	Amount          float64 `json:"amount"`
	AutoCreateFlow  bool    `json:"auto_create_flow"`
	PaymentMethodID *int32  `json:"payment_method_id"`
	CardOwner       string  `json:"card_owner"`
}

type PicuinhaEntryResponse struct {
	ID              int32   `json:"id"`
	PersonID        int32   `json:"person_id"`
	Amount          float64 `json:"amount"`
	Kind            string  `json:"kind"`
	CashFlowID      *int32  `json:"cash_flow_id"`
	PaymentMethodID *int32  `json:"payment_method_id"`
	CardOwner       string  `json:"card_owner"`
	CreatedAt       string  `json:"created_at"`
}

type CreateCaseRequest struct {
	PersonID                 int32    `json:"person_id"`
	Title                    string   `json:"title"`
	CaseType                 string   `json:"case_type"`
	TotalAmount              float64  `json:"total_amount"`
	InstallmentCount         int32    `json:"installment_count"`
	InstallmentAmount        float64  `json:"installment_amount"`
	StartDate                string   `json:"start_date"`
	PaymentMethodID          *int32   `json:"payment_method_id"`
	InstallmentPlanID        *int32   `json:"installment_plan_id"`
	CategoryID               *int32   `json:"category_id"`
	InterestRate             *float64 `json:"interest_rate"`
	InterestRateUnit         string   `json:"interest_rate_unit"`
	RecurrenceIntervalMonths *int32   `json:"recurrence_interval_months"`
}

type CaseResponse struct {
	ID                       int32    `json:"id"`
	PersonID                 int32    `json:"person_id"`
	Title                    string   `json:"title"`
	CaseType                 string   `json:"case_type"`
	TotalAmount              *float64 `json:"total_amount"`
	InstallmentCount         *int32   `json:"installment_count"`
	InstallmentAmount        *float64 `json:"installment_amount"`
	StartDate                string   `json:"start_date"`
	PaymentMethodID          *int32   `json:"payment_method_id"`
	InstallmentPlanID        *int32   `json:"installment_plan_id"`
	CategoryID               *int32   `json:"category_id"`
	InterestRate             *float64 `json:"interest_rate"`
	InterestRateUnit         string   `json:"interest_rate_unit"`
	RecurrenceIntervalMonths *int32   `json:"recurrence_interval_months"`
	InstallmentsTotal        int32    `json:"installments_total"`
	InstallmentsPaid         int32    `json:"installments_paid"`
	AmountPaid               float64  `json:"amount_paid"`
	AmountRemaining          float64  `json:"amount_remaining"`
	Status                   string   `json:"status"`
}

type CaseInstallmentResponse struct {
	ID                int32   `json:"id"`
	CaseID            int32   `json:"case_id"`
	InstallmentNumber int32   `json:"installment_number"`
	DueDate           string  `json:"due_date"`
	Amount            float64 `json:"amount"`
	ExtraAmount       float64 `json:"extra_amount"`
	IsPaid            bool    `json:"is_paid"`
	PaidAt            *string `json:"paid_at"`
}

type UpdateCaseInstallmentRequest struct {
	IsPaid      bool    `json:"is_paid"`
	ExtraAmount float64 `json:"extra_amount"`
}
