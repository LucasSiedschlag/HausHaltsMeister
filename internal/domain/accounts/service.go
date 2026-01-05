package accounts

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/LucasSiedschlag/HausHaltsMeister/internal/domain/ledger"
)

type Repository interface {
	ListAccounts(ctx context.Context, ledgerID string) ([]Account, error)
	GetAccount(ctx context.Context, ledgerID, accountID string) (Account, error)
	CreateAccount(ctx context.Context, ledgerID, name, accountType string, isActive bool) (Account, error)
	UpdateAccount(ctx context.Context, ledgerID, accountID, name string, isActive bool, updatedAt time.Time) (Account, error)
	DeactivateAccount(ctx context.Context, ledgerID, accountID string, updatedAt time.Time) error
	LedgerExists(ctx context.Context, ledgerID string) (bool, error)
	GetLedgerRole(ctx context.Context, ledgerID, userID string) (string, error)
}

type Service struct {
	repo Repository
	now  func() time.Time
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo, now: time.Now().UTC}
}

func (s *Service) ListAccounts(ctx context.Context, userID, ledgerID string) ([]Account, error) {
	if err := s.requireRole(ctx, ledgerID, userID, "viewer"); err != nil {
		return nil, err
	}
	return s.repo.ListAccounts(ctx, ledgerID)
}

func (s *Service) GetAccount(ctx context.Context, userID, ledgerID, accountID string) (Account, error) {
	if err := s.requireRole(ctx, ledgerID, userID, "viewer"); err != nil {
		return Account{}, err
	}
	account, err := s.repo.GetAccount(ctx, ledgerID, accountID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return Account{}, ErrAccountNotFound
		}
		return Account{}, s.mapAccessError(ctx, ledgerID, err)
	}
	return account, nil
}

func (s *Service) CreateAccount(ctx context.Context, userID, ledgerID, name, accountType string, isActive bool) (Account, error) {
	if err := s.requireRole(ctx, ledgerID, userID, "editor"); err != nil {
		return Account{}, err
	}

	name = strings.TrimSpace(name)
	if name == "" {
		return Account{}, NewError("VALIDATION_ERROR", "Validacao falhou", map[string]string{"name": "required"})
	}
	accountType = strings.TrimSpace(accountType)
	if !isValidType(accountType) {
		return Account{}, NewError("VALIDATION_ERROR", "Validacao falhou", map[string]string{"type": "invalid"})
	}

	account, err := s.repo.CreateAccount(ctx, ledgerID, name, accountType, isActive)
	if err != nil {
		if errors.Is(err, ErrDuplicateName) {
			return Account{}, ErrDuplicateName
		}
		return Account{}, err
	}
	return account, nil
}

func (s *Service) UpdateAccount(ctx context.Context, userID, ledgerID, accountID, name string, isActive bool) (Account, error) {
	if err := s.requireRole(ctx, ledgerID, userID, "editor"); err != nil {
		return Account{}, err
	}

	name = strings.TrimSpace(name)
	if name == "" {
		return Account{}, NewError("VALIDATION_ERROR", "Validacao falhou", map[string]string{"name": "required"})
	}

	updatedAt := s.now()
	account, err := s.repo.UpdateAccount(ctx, ledgerID, accountID, name, isActive, updatedAt)
	if err != nil {
		if errors.Is(err, ErrDuplicateName) {
			return Account{}, ErrDuplicateName
		}
		if errors.Is(err, ErrNotFound) {
			return Account{}, ErrAccountNotFound
		}
		return Account{}, err
	}
	return account, nil
}

func (s *Service) DeleteAccount(ctx context.Context, userID, ledgerID, accountID string) error {
	if err := s.requireRole(ctx, ledgerID, userID, "editor"); err != nil {
		return err
	}

	updatedAt := s.now()
	if err := s.repo.DeactivateAccount(ctx, ledgerID, accountID, updatedAt); err != nil {
		if errors.Is(err, ErrNotFound) {
			return ErrAccountNotFound
		}
		return err
	}
	return nil
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

func isValidType(accountType string) bool {
	switch accountType {
	case "cash", "investment", "credit_card":
		return true
	default:
		return false
	}
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
