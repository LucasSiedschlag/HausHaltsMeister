package installment

import (
	"errors"
	"time"
)

var (
	ErrInvalidTotalAmount = errors.New("total amount must be greater than zero")
	ErrInvalidCount       = errors.New("installment count must be at least 1")
)

type InstallmentPlan struct {
	ID                int32
	Description       string
	TotalAmount       float64
	InstallmentCount  int32
	InstallmentAmount float64
	StartMonth        time.Time
	PaymentMethodID   int32
}

func NewPlan(description string, totalAmount float64, count int32, startMonth time.Time, paymentMethodID int32) (*InstallmentPlan, error) {
	if totalAmount <= 0 {
		return nil, ErrInvalidTotalAmount
	}
	if count < 1 {
		return nil, ErrInvalidCount
	}

	// Simple Division
	amountPerInstallment := totalAmount / float64(count)

	return &InstallmentPlan{
		Description:       description,
		TotalAmount:       totalAmount,
		InstallmentCount:  count,
		InstallmentAmount: amountPerInstallment,
		StartMonth:        startMonth,
		PaymentMethodID:   paymentMethodID,
	}, nil
}
