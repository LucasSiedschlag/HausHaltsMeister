package ledger

import (
	"context"
	"errors"
	"strings"
	"time"
)

type Repository interface {
	ListLedgersForUser(ctx context.Context, userID string) ([]LedgerWithRole, error)
	CreateLedger(ctx context.Context, params CreateLedgerParams) (Ledger, error)
	GetLedgerForUser(ctx context.Context, ledgerID, userID string) (LedgerWithRole, error)
	GetLedgerByID(ctx context.Context, ledgerID string) (Ledger, error)
	UpdateLedger(ctx context.Context, ledgerID, name string, updatedAt time.Time) (Ledger, error)
	LedgerExists(ctx context.Context, ledgerID string) (bool, error)
	GetLedgerRole(ctx context.Context, ledgerID, userID string) (string, error)
	ListMembers(ctx context.Context, ledgerID string) ([]Member, error)
	AddMember(ctx context.Context, ledgerID, userID, role string, updatedAt time.Time) (Member, error)
	UpdateMemberRole(ctx context.Context, ledgerID, userID, role string, updatedAt time.Time) (Member, error)
	RemoveMember(ctx context.Context, ledgerID, userID string) error
}

type CreateLedgerParams struct {
	OwnerUserID  string
	Name         string
	CurrencyCode string
}

type Service struct {
	repo Repository
	now  func() time.Time
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo, now: time.Now().UTC}
}

func (s *Service) ListLedgers(ctx context.Context, userID string) ([]LedgerWithRole, error) {
	return s.repo.ListLedgersForUser(ctx, userID)
}

func (s *Service) GetMembership(ctx context.Context, userID, ledgerID string) (string, error) {
	role, err := s.repo.GetLedgerRole(ctx, ledgerID, userID)
	if err != nil {
		return "", s.mapAccessError(ctx, ledgerID, err)
	}
	return role, nil
}

func (s *Service) CreateLedger(ctx context.Context, userID, name, currencyCode string) (Ledger, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return Ledger{}, NewError("VALIDATION_ERROR", "Validacao falhou", map[string]string{"name": "required"})
	}
	currencyCode = strings.TrimSpace(currencyCode)
	if currencyCode == "" {
		currencyCode = "BRL"
	}
	return s.repo.CreateLedger(ctx, CreateLedgerParams{
		OwnerUserID:  userID,
		Name:         name,
		CurrencyCode: currencyCode,
	})
}

func (s *Service) GetLedger(ctx context.Context, userID, ledgerID string) (LedgerWithRole, error) {
	ledger, err := s.repo.GetLedgerForUser(ctx, ledgerID, userID)
	if err == nil {
		return ledger, nil
	}
	return LedgerWithRole{}, s.mapAccessError(ctx, ledgerID, err)
}

func (s *Service) UpdateLedger(ctx context.Context, userID, ledgerID, name string) (Ledger, error) {
	if err := s.requireRole(ctx, ledgerID, userID, "owner"); err != nil {
		return Ledger{}, err
	}

	name = strings.TrimSpace(name)
	if name == "" {
		return Ledger{}, NewError("VALIDATION_ERROR", "Validacao falhou", map[string]string{"name": "required"})
	}

	updatedAt := s.now()
	return s.repo.UpdateLedger(ctx, ledgerID, name, updatedAt)
}

func (s *Service) DeleteLedger(ctx context.Context, userID, ledgerID string) error {
	if err := s.requireRole(ctx, ledgerID, userID, "owner"); err != nil {
		return err
	}
	return ErrNotImplemented
}

func (s *Service) ListMembers(ctx context.Context, userID, ledgerID string) ([]Member, error) {
	if err := s.requireRole(ctx, ledgerID, userID, "owner"); err != nil {
		return nil, err
	}
	return s.repo.ListMembers(ctx, ledgerID)
}

func (s *Service) AddMember(ctx context.Context, userID, ledgerID, memberUserID, role string) (Member, error) {
	if err := s.requireRole(ctx, ledgerID, userID, "owner"); err != nil {
		return Member{}, err
	}
	role = strings.TrimSpace(role)
	if role == "" {
		role = "viewer"
	}
	if !isValidRole(role) || role == "owner" {
		return Member{}, NewError("VALIDATION_ERROR", "Validacao falhou", map[string]string{"role": "invalid"})
	}

	updatedAt := s.now()
	member, err := s.repo.AddMember(ctx, ledgerID, memberUserID, role, updatedAt)
	if err != nil {
		if errors.Is(err, ErrMemberExists) {
			return Member{}, ErrMemberExists
		}
		return Member{}, err
	}
	return member, nil
}

func (s *Service) UpdateMember(ctx context.Context, userID, ledgerID, memberUserID, role string) (Member, error) {
	if err := s.requireRole(ctx, ledgerID, userID, "owner"); err != nil {
		return Member{}, err
	}
	role = strings.TrimSpace(role)
	if role == "" || !isValidRole(role) || role == "owner" {
		return Member{}, NewError("VALIDATION_ERROR", "Validacao falhou", map[string]string{"role": "invalid"})
	}

	ledger, err := s.repo.GetLedgerByID(ctx, ledgerID)
	if err != nil {
		return Member{}, s.mapAccessError(ctx, ledgerID, err)
	}
	if memberUserID == ledger.OwnerUserID {
		return Member{}, NewError("VALIDATION_ERROR", "Validacao falhou", map[string]string{"user_id": "owner"})
	}

	updatedAt := s.now()
	member, err := s.repo.UpdateMemberRole(ctx, ledgerID, memberUserID, role, updatedAt)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return Member{}, NewError("VALIDATION_ERROR", "Validacao falhou", map[string]string{"user_id": "not_found"})
		}
		return Member{}, err
	}
	return member, nil
}

func (s *Service) RemoveMember(ctx context.Context, userID, ledgerID, memberUserID string) error {
	if err := s.requireRole(ctx, ledgerID, userID, "owner"); err != nil {
		return err
	}

	ledger, err := s.repo.GetLedgerByID(ctx, ledgerID)
	if err != nil {
		return s.mapAccessError(ctx, ledgerID, err)
	}
	if memberUserID == ledger.OwnerUserID {
		return NewError("VALIDATION_ERROR", "Validacao falhou", map[string]string{"user_id": "owner"})
	}

	if err := s.repo.RemoveMember(ctx, ledgerID, memberUserID); err != nil {
		if errors.Is(err, ErrNotFound) {
			return NewError("VALIDATION_ERROR", "Validacao falhou", map[string]string{"user_id": "not_found"})
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
	if errors.Is(err, ErrNotFound) {
		exists, checkErr := s.repo.LedgerExists(ctx, ledgerID)
		if checkErr == nil && exists {
			return ErrAccessDenied
		}
		return ErrLedgerNotFound
	}
	return err
}

func isValidRole(role string) bool {
	switch role {
	case "owner", "editor", "viewer":
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
