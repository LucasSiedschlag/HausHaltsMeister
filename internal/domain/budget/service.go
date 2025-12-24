package budget

import (
	"context"
	"fmt"
	"time"

	"github.com/seuuser/cashflow/internal/domain/category"
)

type BudgetService struct {
	repo    Repository
	catRepo category.Repository
}

func NewService(repo Repository, catRepo category.Repository) *BudgetService {
	return &BudgetService{
		repo:    repo,
		catRepo: catRepo,
	}
}

func (s *BudgetService) GetOrCreatePeriod(ctx context.Context, month time.Time) (*BudgetPeriod, error) {
	// 1. Try to get existing
	period, err := s.repo.GetPeriodByMonth(ctx, month)
	if err != nil {
		return nil, err
	}
	if period != nil {
		// Fetch items
		items, err := s.repo.GetItemsByPeriod(ctx, period.ID)
		if err != nil {
			return nil, err
		}
		period.Items = items
		return period, nil
	}

	// 2. Create new
	newPeriod := NewPeriod(month)
	created, err := s.repo.CreatePeriod(ctx, newPeriod)
	if err != nil {
		return nil, err
	}
	created.Items = []BudgetItem{}
	return created, nil
}

func (s *BudgetService) SetBudgetItem(ctx context.Context, month time.Time, categoryID int32, plannedAmount float64) (*BudgetItem, error) {
	// 1. Validate Category (Must be OUT and Active, potentially)
	cat, err := s.catRepo.GetByID(ctx, categoryID)
	if err != nil {
		return nil, err
	}
	if cat == nil {
		return nil, ErrInvalidCategory
	}
	if cat.Direction != "OUT" {
		return nil, fmt.Errorf("%w: only OUT categories allowed in budget", ErrInvalidCategory)
	}

	// 2. Ensure Period Exists
	period, err := s.GetOrCreatePeriod(ctx, month)
	if err != nil {
		return nil, err
	}

	if period.IsClosed {
		return nil, fmt.Errorf("budget period is closed")
	}

	// 3. Upsert Item
	// For now supporting only ABSOLUTE value logic from simple input
	item := &BudgetItem{
		BudgetPeriodID: period.ID,
		CategoryID:     categoryID,
		Mode:           ModeAbsolute,
		PlannedAmount:  plannedAmount,
		TargetPercent:  0,
		Notes:          "",
	}

	return s.repo.UpsertItem(ctx, item)
}

func (s *BudgetService) GetBudgetSummary(ctx context.Context, month time.Time) (*BudgetPeriod, error) {
	// Reusing GetOrCreate for now, but in future would calculate actuals
	return s.GetOrCreatePeriod(ctx, month)
}
