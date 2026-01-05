package reports

import (
	"context"
	"errors"
	"time"

	"github.com/LucasSiedschlag/HausHaltsMeister/internal/domain/ledger"
)

type Repository interface {
	GetLedgerRole(ctx context.Context, ledgerID, userID string) (string, error)
	LedgerExists(ctx context.Context, ledgerID string) (bool, error)
	ListAccountBalances(ctx context.Context, ledgerID string, cutoff time.Time) ([]AccountBalance, error)
	ListCategorySummary(ctx context.Context, ledgerID string, from, to time.Time) ([]CategorySummary, error)
	ListCashflow(ctx context.Context, ledgerID string, from, to time.Time) ([]CashflowEntry, error)
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Balances(ctx context.Context, userID, ledgerID string, month time.Time) (BalanceReport, error) {
	if err := s.requireRole(ctx, ledgerID, userID, "viewer"); err != nil {
		return BalanceReport{}, err
	}
	if !isMonthStart(month) {
		return BalanceReport{}, NewError("VALIDATION_ERROR", "Validacao falhou", map[string]string{"month": "invalid"})
	}

	cutoff := month.AddDate(0, 1, 0)
	items, err := s.repo.ListAccountBalances(ctx, ledgerID, cutoff)
	if err != nil {
		return BalanceReport{}, err
	}
	return BalanceReport{Month: month, Items: items}, nil
}

func (s *Service) CategorySummary(ctx context.Context, userID, ledgerID string, from, to time.Time) (CategoryReport, error) {
	if err := s.requireRole(ctx, ledgerID, userID, "viewer"); err != nil {
		return CategoryReport{}, err
	}
	if from.IsZero() || to.IsZero() || to.Before(from) {
		return CategoryReport{}, NewError("VALIDATION_ERROR", "Validacao falhou", map[string]string{"range": "invalid"})
	}

	items, err := s.repo.ListCategorySummary(ctx, ledgerID, from, to)
	if err != nil {
		return CategoryReport{}, err
	}
	return CategoryReport{From: from, To: to, Items: items}, nil
}

func (s *Service) Cashflow(ctx context.Context, userID, ledgerID string, from, to time.Time) (CashflowReport, error) {
	if err := s.requireRole(ctx, ledgerID, userID, "viewer"); err != nil {
		return CashflowReport{}, err
	}
	if from.IsZero() || to.IsZero() || to.Before(from) {
		return CashflowReport{}, NewError("VALIDATION_ERROR", "Validacao falhou", map[string]string{"range": "invalid"})
	}

	items, err := s.repo.ListCashflow(ctx, ledgerID, from, to)
	if err != nil {
		return CashflowReport{}, err
	}
	return CashflowReport{From: from, To: to, Items: items}, nil
}

func (s *Service) requireRole(ctx context.Context, ledgerID, userID, minRole string) error {
	role, err := s.repo.GetLedgerRole(ctx, ledgerID, userID)
	if err != nil {
		return s.mapAccessError(ctx, ledgerID, err)
	}
	if roleRank(role) < roleRank(minRole) {
		return ErrAccessDenied
	}
	return nil
}

func (s *Service) mapAccessError(ctx context.Context, ledgerID string, err error) error {
	if errors.Is(err, ErrNotFound) || errors.Is(err, ledger.ErrNotFound) {
		exists, checkErr := s.repo.LedgerExists(ctx, ledgerID)
		if checkErr == nil && exists {
			return ErrAccessDenied
		}
		return ErrLedgerNotFound
	}
	return err
}

func roleRank(role string) int {
	switch role {
	case "owner":
		return 3
	case "editor":
		return 2
	case "viewer":
		return 1
	default:
		return 0
	}
}

func isMonthStart(value time.Time) bool {
	return value.Day() == 1 && value.Equal(time.Date(value.Year(), value.Month(), 1, 0, 0, 0, 0, value.Location()))
}
