package cashflow

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/seuuser/cashflow/internal/domain/category"
)

var (
	ErrCategoryNotFound  = errors.New("category not found")
	ErrDirectionMismatch = errors.New("cash flow direction does not match category direction")
)

type CashFlowService struct {
	repo    Repository
	catRepo category.Repository
}

func NewService(repo Repository, catRepo category.Repository) *CashFlowService {
	return &CashFlowService{
		repo:    repo,
		catRepo: catRepo,
	}
}

func (s *CashFlowService) CreateCashFlow(ctx context.Context, date time.Time, categoryID int32, direction, title string, amount float64) (*CashFlow, error) {
	newFlow, err := New(date, categoryID, direction, title, amount)
	if err != nil {
		return nil, fmt.Errorf("domain validation failed: %w", err)
	}

	// Validate Category and Direction
	cat, err := s.catRepo.GetByID(ctx, categoryID)
	if err != nil {
		return nil, fmt.Errorf("failed to get category: %w", err)
	}
	if cat == nil {
		return nil, ErrCategoryNotFound
	}

	if cat.Direction != direction {
		return nil, ErrDirectionMismatch
	}

	return s.repo.Create(ctx, newFlow)
}

func (s *CashFlowService) ListCashFlows(ctx context.Context, month time.Time) ([]*CashFlow, error) {
	return s.repo.ListByMonth(ctx, month)
}
