package cashflow

import "context"

type Repository interface {
	Create(ctx context.Context, cf *CashFlow) (*CashFlow, error)
	ListByMonth(ctx context.Context, monthYear string) ([]CashFlow, error)
}

