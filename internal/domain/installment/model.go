package installment

import (
	"errors"
	"time"
)

var (
	ErrInvalidTotalAmount = errors.New("total amount must be greater than zero")
	ErrInvalidCount       = errors.New("installment count must be at least 1")
	ErrInvalidCategory    = errors.New("category must be OUT")
)

type InstallmentPlan struct {
	ID                int32
	Description       string
	PlanType          string
	TotalAmount       float64
	InstallmentCount  int32
	InstallmentAmount float64
	StartDate         time.Time
	PaymentMethodID   int32
	CategoryID        int32
}

func NewPlan(description string, totalAmount float64, count int32, startDate time.Time, paymentMethodID int32, categoryID int32) (*InstallmentPlan, error) {
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
		PlanType:          "CARD_INSTALLMENT",
		TotalAmount:       totalAmount,
		InstallmentCount:  count,
		InstallmentAmount: amountPerInstallment,
		StartDate:         startDate,
		PaymentMethodID:   paymentMethodID,
		CategoryID:        categoryID,
	}, nil
}
