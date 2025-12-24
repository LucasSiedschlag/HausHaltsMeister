package cashflow

import (
	"context"
	"errors"
	"time"
)

var (
	ErrInvalidAmount    = errors.New("amount must be positive")
	ErrInvalidDirection = errors.New("invalid direction")
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

type CreateCashFlowInput struct {
	Date       time.Time
	CategoryID int64
	Direction  Direction
	Title      string
	Amount     float64
}

func (s *Service) CreateCashFlow(ctx context.Context, in CreateCashFlowInput) (*CashFlow, error) {
	if in.Amount <= 0 {
		return nil, ErrInvalidAmount
	}

	if in.Direction != DirectionIn && in.Direction != DirectionOut {
		return nil, ErrInvalidDirection
	}

	cf := &CashFlow{
		Date:       in.Date,
		CategoryID: in.CategoryID,
		Direction:  in.Direction,
		Title:      in.Title,
		Amount:     in.Amount,
	}

	return s.repo.Create(ctx, cf)
}

func (s *Service) ListCashFlowsByMonth(ctx context.Context, month time.Time) ([]CashFlow, error) {
	// Ex: passamos "2025-12-01" para a query
	monthStr := month.Format("2006-01-02")
	return s.repo.ListByMonth(ctx, monthStr)
}

