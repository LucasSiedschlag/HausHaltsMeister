package cashflow

import (
	"context"
	"time"
)

type Repository interface {
	Create(ctx context.Context, flow *CashFlow) (*CashFlow, error)
	Update(ctx context.Context, flow *CashFlow) (*CashFlow, error)
	Delete(ctx context.Context, id int32) error
	GetByID(ctx context.Context, id int32) (*CashFlow, error)
	ListByMonth(ctx context.Context, month time.Time, direction *string, isFixed *bool) ([]*CashFlow, error)
	GetMonthlySummary(ctx context.Context, month time.Time) (*MonthlySummary, error)
	GetCategorySummary(ctx context.Context, month time.Time) ([]CategorySummary, error)
}

type Service interface {
	CreateCashFlow(ctx context.Context, date time.Time, categoryID int32, paymentMethodID int32, direction, title string, amount float64, isFixed bool) (*CashFlow, error)
	UpdateCashFlow(ctx context.Context, id int32, date time.Time, categoryID int32, paymentMethodID int32, direction, title string, amount float64, isFixed bool) (*CashFlow, error)
	DeleteCashFlow(ctx context.Context, id int32) error
	ReverseCashFlow(ctx context.Context, id int32) (*CashFlow, error)
	ListCashFlows(ctx context.Context, month time.Time, direction *string, isFixed *bool) ([]*CashFlow, error)
	CopyFixedExpenses(ctx context.Context, fromMonth, toMonth time.Time) (int, error)
	GetMonthlySummary(ctx context.Context, month time.Time) (*MonthlySummary, error)
	GetCategorySummary(ctx context.Context, month time.Time) ([]CategorySummary, error)
}
