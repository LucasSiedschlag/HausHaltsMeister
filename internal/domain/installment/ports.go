package installment

import (
	"context"
	"time"
)

type InstallmentPlanItem struct {
	ID            int32
	PlanID        int32
	Sequence      int32
	DueDate       time.Time
	Amount        float64
	TransactionID *int32
}

type Repository interface {
	CreatePlan(ctx context.Context, plan *InstallmentPlan) (*InstallmentPlan, error)
	CreatePlanItem(ctx context.Context, item *InstallmentPlanItem) (*InstallmentPlanItem, error)
	ListPlanItemsByPlan(ctx context.Context, planID int32) ([]*InstallmentPlanItem, error)
}

type Service interface {
	CreateInstallmentPurchase(ctx context.Context, description string, totalAmount float64, count int32, categoryID int32, paymentMethodID int32, purchaseDate time.Time) (*InstallmentPlan, error)
	ListPlanItems(ctx context.Context, planID int32) ([]*InstallmentPlanItem, error)
}
