package cashflow

import (
	"context"
	"time"
)

type Repository interface {
	Create(ctx context.Context, flow *CashFlow) (*CashFlow, error)
	ListByMonth(ctx context.Context, month time.Time) ([]*CashFlow, error)
	GetMonthlySummary(ctx context.Context, month time.Time) (*MonthlySummary, error)
	GetCategorySummary(ctx context.Context, month time.Time) ([]CategorySummary, error)
}

type Service interface {
	CreateCashFlow(ctx context.Context, date time.Time, categoryID int32, direction, title string, amount float64, isFixed bool) (*CashFlow, error)
	ListCashFlows(ctx context.Context, month time.Time) ([]*CashFlow, error)
	CopyFixedExpenses(ctx context.Context, fromMonth, toMonth time.Time) (int, error)
	GetMonthlySummary(ctx context.Context, month time.Time) (*MonthlySummary, error)
	GetCategorySummary(ctx context.Context, month time.Time) ([]CategorySummary, error)
}
