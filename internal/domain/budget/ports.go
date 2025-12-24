package budget

import (
	"context"
	"time"
)

type Repository interface {
	GetPeriodByMonth(ctx context.Context, month time.Time) (*BudgetPeriod, error)
	CreatePeriod(ctx context.Context, period *BudgetPeriod) (*BudgetPeriod, error)
	UpsertItem(ctx context.Context, item *BudgetItem) (*BudgetItem, error)
	GetItemsByPeriod(ctx context.Context, periodID int32) ([]BudgetItem, error)
}

type Service interface {
	GetOrCreatePeriod(ctx context.Context, month time.Time) (*BudgetPeriod, error)
	SetBudgetItem(ctx context.Context, month time.Time, categoryID int32, plannedAmount float64) (*BudgetItem, error)
	GetBudgetSummary(ctx context.Context, month time.Time) (*BudgetPeriod, error)
	SetBudgetBatch(ctx context.Context, startMonth, endMonth time.Time, categoryID int32, plannedAmount float64) error
}
