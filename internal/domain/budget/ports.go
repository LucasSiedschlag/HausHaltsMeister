package budget

import (
	"context"
	"time"
)

type Repository interface {
	GetPeriodByMonth(ctx context.Context, month time.Time) (*BudgetPeriod, error)
	GetLatestPeriodWithItemsBefore(ctx context.Context, month time.Time) (*BudgetPeriod, error)
	CreatePeriod(ctx context.Context, period *BudgetPeriod) (*BudgetPeriod, error)
	UpsertItem(ctx context.Context, item *BudgetItem) (*BudgetItem, error)
	GetItemsByPeriod(ctx context.Context, periodID int32) ([]BudgetItem, error)
	GetItemByID(ctx context.Context, id int32) (*BudgetItem, error)
	UpdateItem(ctx context.Context, item *BudgetItem) (*BudgetItem, error)
}

type Service interface {
	GetOrCreatePeriod(ctx context.Context, month time.Time) (*BudgetPeriod, error)
	SetBudgetItem(ctx context.Context, month time.Time, categoryID int32, mode string, plannedAmount float64, targetPercent float64) (*BudgetItem, error)
	GetBudgetSummary(ctx context.Context, month time.Time) (*BudgetPeriod, error)
	SetBudgetBatch(ctx context.Context, startMonth, endMonth time.Time, categoryID int32, mode string, plannedAmount float64, targetPercent float64) error
	UpdateBudgetItem(ctx context.Context, id int32, mode string, plannedAmount float64, targetPercent float64) (*BudgetItem, error)
}
