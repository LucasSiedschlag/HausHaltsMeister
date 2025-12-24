package cashflow

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/LucasSiedschlag/HausHaltsMeister/internal/domain/category"
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

func (s *CashFlowService) CreateCashFlow(ctx context.Context, date time.Time, categoryID int32, direction, title string, amount float64, isFixed bool) (*CashFlow, error) {
	newFlow, err := New(date, categoryID, direction, title, amount, isFixed)
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

func (s *CashFlowService) CopyFixedExpenses(ctx context.Context, fromMonth, toMonth time.Time) (int, error) {
	// 1. List from previous month
	sourceFlows, err := s.repo.ListByMonth(ctx, fromMonth)
	if err != nil {
		return 0, fmt.Errorf("failed to list source month expenses: %w", err)
	}

	count := 0
	for _, flow := range sourceFlows {
		if !flow.IsFixed {
			continue
		}

		// 2. Clone to new month
		// Adjust date to same day in target month
		_, _, day := flow.Date.Date()

		// Simple Date Construction: toMonth Year/Month, same Day.
		// Handle overflow (e.g. Jan 31 -> Feb 31 doesn't exist)
		// time.Date automaticaly normalizes Feb 31 to March 3.
		// We might want to clamp to end of month.
		// However, for MVP/Agent: Standard normalization is acceptable or logic to clamp.
		// Let's rely on standard time.Date normalization for now, or use a helper if we wanted perfection.
		// For financial fixed expenses (rent 10th), usually safe. For 31st, it might jump.
		// Let's clamp to last day of month if necessary.

		targetDate := time.Date(toMonth.Year(), toMonth.Month(), day, 0, 0, 0, 0, toMonth.Location())

		// If targetDate jumped to next month (e.g. was 31st, became 1st of next), rollback to end of current month
		if targetDate.Month() != toMonth.Month() {
			// Set to day 0 of date's month -> Last day of previous (target) month
			targetDate = time.Date(targetDate.Year(), targetDate.Month(), 0, 0, 0, 0, 0, targetDate.Location())
		}

		_, err := s.CreateCashFlow(
			ctx,
			targetDate,
			flow.CategoryID,
			flow.Direction,
			flow.Title,
			flow.Amount,
			true, // Keep it fixed for next month too
		)
		if err != nil {
			return count, fmt.Errorf("failed to copy flow %d: %w", flow.ID, err)
		}
		count++
	}
	return count, nil
}

func (s *CashFlowService) GetMonthlySummary(ctx context.Context, month time.Time) (*MonthlySummary, error) {
	return s.repo.GetMonthlySummary(ctx, month)
}

func (s *CashFlowService) GetCategorySummary(ctx context.Context, month time.Time) ([]CategorySummary, error) {
	return s.repo.GetCategorySummary(ctx, month)
}
