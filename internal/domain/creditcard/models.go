package creditcard

import "time"

type CardNetwork struct {
	Code        string
	DisplayName string
	CreatedAt   time.Time
	UpdatedAt   *time.Time
}

type AccountInfo struct {
	ID       string
	LedgerID string
	Type     string
	Nature   string
	IsActive bool
}

type CreditCard struct {
	ID                 string
	LedgerID           string
	ParentAccountID    string
	LiabilityAccountID string
	Label              *string
	Brand              string
	Last4              *string
	CVV                *string
	HolderName         *string
	Active             bool
	Color              *string
	Style              *string
	ClosingDay         int
	DueDay             int
	CreatedAt          time.Time
	UpdatedAt          *time.Time
}

type InstallmentPlan struct {
	ID                     string
	LedgerID               string
	CreditCardID           string
	PurchaseOccurredAt     time.Time
	Merchant               *string
	Description            string
	CategoryID             string
	TotalAmountCents       int64
	InstallmentsCount      int
	InstallmentAmountCents int64
	FirstDueMonth          time.Time
	Status                 string
	CreatedByUserID        string
	CreatedAt              time.Time
	UpdatedAt              *time.Time
}

type Installment struct {
	ID                  string
	LedgerID            string
	PlanID              string
	InstallmentNo       int
	DueMonth            time.Time
	AmountCents         int64
	Status              string
	PostedTransactionID *string
	PaidStatementID     *string
	CreatedAt           time.Time
	UpdatedAt           *time.Time
}

type Statement struct {
	ID                   string
	LedgerID             string
	CreditCardID         string
	StatementMonth       time.Time
	ClosingDate          time.Time
	DueDate              time.Time
	TotalChargesCents    int64
	TotalPaymentsCents   int64
	Status               string
	PaymentTransactionID *string
	CreatedAt            time.Time
	UpdatedAt            *time.Time
}

type PostingResult struct {
	PostedCount  int
	Transactions []string
}
