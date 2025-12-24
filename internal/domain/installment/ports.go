package installment

import (
	"context"
	"time"
)

type Repository interface {
	CreatePlan(ctx context.Context, plan *InstallmentPlan) (*InstallmentPlan, error)
	CreateExpenseDetail(ctx context.Context, cashFlowID int32, paymentMethodID int32, planID int32, affectsCardInvoice bool) error
}

type Service interface {
	CreateInstallmentPurchase(ctx context.Context, description string, totalAmount float64, count int32, categoryID int32, paymentMethodID int32, purchaseDate time.Time) (*InstallmentPlan, error)
}
