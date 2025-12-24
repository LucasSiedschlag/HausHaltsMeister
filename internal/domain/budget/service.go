package budget

import (
	"context"
	"fmt"
	"time"

	"github.com/LucasSiedschlag/HausHaltsMeister/internal/domain/cashflow"
	"github.com/LucasSiedschlag/HausHaltsMeister/internal/domain/category"
)

type BudgetService struct {
	repo    Repository
	catRepo category.Repository
	cfRepo  cashflow.Repository
}

func NewService(repo Repository, catRepo category.Repository, cfRepo cashflow.Repository) *BudgetService {
	return &BudgetService{
		repo:    repo,
		catRepo: catRepo,
		cfRepo:  cfRepo,
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
	// 1. Get Base Plan
	period, err := s.GetOrCreatePeriod(ctx, month)
	if err != nil {
		return nil, err
	}

	// 2. Get Actuals (CashFlows)
	// Assuming month is the 1st of the month
	flows, err := s.cfRepo.ListByMonth(ctx, month)
	if err != nil {
		return nil, err
	}

	// 3. Aggregate Actuals by Category
	actuals := make(map[int32]float64)
	for _, f := range flows {
		if f.Direction == "OUT" {
			actuals[f.CategoryID] += f.Amount
		}
	}

	// 4. Enrich Items
	for i := range period.Items {
		period.Items[i].ActualAmount = actuals[period.Items[i].CategoryID]
	}

	return period, nil
}

func (s *BudgetService) SetBudgetBatch(ctx context.Context, startMonth, endMonth time.Time, categoryID int32, plannedAmount float64) error {
	// Normalize to 1st of month
	current := time.Date(startMonth.Year(), startMonth.Month(), 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(endMonth.Year(), endMonth.Month(), 1, 0, 0, 0, 0, time.UTC)

	for !current.After(end) {
		_, err := s.SetBudgetItem(ctx, current, categoryID, plannedAmount)
		if err != nil {
			return fmt.Errorf("failed at month %s: %w", current.Format("2006-01"), err)
		}
		current = current.AddDate(0, 1, 0)
	}
	return nil
}
