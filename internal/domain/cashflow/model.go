package cashflow

import (
	"errors"
	"time"
)

var (
	ErrInvalidAmount = errors.New("amount must be greater than zero")
	ErrEmptyTitle    = errors.New("title cannot be empty")
	ErrInvalidDate   = errors.New("date is required")
	ErrPaymentMethod = errors.New("payment method is required")
)

type CashFlow struct {
	ID                    int32
	TransactionID         int32
	InstallmentPlanID     *int32
	InstallmentPlanItemID *int32
	Date                  time.Time
	CategoryID            int32
	CategoryName          string
	PaymentMethodID       int32
	PaymentMethodName     string
	Direction             string
	Title                 string
	Amount                float64
	IsFixed               bool
	ReversalOfEntryID     *int32
}

type MonthlySummary struct {
	TotalIncome  float64 `json:"total_income"`
	TotalExpense float64 `json:"total_expense"`
	Balance      float64 `json:"balance"`
}

type CategorySummary struct {
	CategoryName string  `json:"category_name"`
	Direction    string  `json:"direction"`
	TotalAmount  float64 `json:"total_amount"`
}

func New(date time.Time, categoryID int32, paymentMethodID int32, direction, title string, amount float64, isFixed bool) (*CashFlow, error) {
	if amount <= 0 {
		return nil, ErrInvalidAmount
	}
	if title == "" {
		return nil, ErrEmptyTitle
	}
	if date.IsZero() {
		return nil, ErrInvalidDate
	}
	if paymentMethodID <= 0 {
		return nil, ErrPaymentMethod
	}

	// Note: Direction validation against Category happens in Service/Repo or Phase 3

	return &CashFlow{
		Date:       date,
		CategoryID: categoryID,
		PaymentMethodID: paymentMethodID,
		Direction:  direction,
		Title:      title,
		Amount:     amount,
		IsFixed:    isFixed,
	}, nil
}
