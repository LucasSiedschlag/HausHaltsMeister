package cashflow

import (
	"context"
	"time"
)

type Repository interface {
	Create(ctx context.Context, flow *CashFlow) (*CashFlow, error)
	ListByMonth(ctx context.Context, month time.Time) ([]*CashFlow, error)
}

type Service interface {
	CreateCashFlow(ctx context.Context, date time.Time, categoryID int32, direction, title string, amount float64) (*CashFlow, error)
	ListCashFlows(ctx context.Context, month time.Time) ([]*CashFlow, error)
}
