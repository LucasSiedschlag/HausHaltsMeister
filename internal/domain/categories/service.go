package categories

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/LucasSiedschlag/HausHaltsMeister/internal/domain/ledger"
)

type Repository interface {
	ListCategories(ctx context.Context, ledgerID string, direction *string, active *bool) ([]Category, error)
	GetCategory(ctx context.Context, ledgerID, categoryID string) (Category, error)
	CreateCategory(ctx context.Context, params CreateCategoryParams) (Category, error)
	UpdateCategory(ctx context.Context, params UpdateCategoryParams) (Category, error)
	DeactivateCategory(ctx context.Context, ledgerID, categoryID string, updatedAt time.Time) error
	LedgerExists(ctx context.Context, ledgerID string) (bool, error)
	GetLedgerRole(ctx context.Context, ledgerID, userID string) (string, error)
	IsCategoryUsed(ctx context.Context, ledgerID, categoryID string) (bool, error)
}

type CreateCategoryParams struct {
	LedgerID         string
	ParentID         *string
	Name             string
	Direction        string
	IsBudgetBase     bool
	IsBudgetRelevant bool
	IsActive         bool
}

type UpdateCategoryParams struct {
	LedgerID         string
	CategoryID       string
	ParentID         *string
	Name             string
	IsBudgetBase     bool
	IsBudgetRelevant bool
	IsActive         bool
	UpdatedAt        time.Time
}

type Service struct {
	repo Repository
	now  func() time.Time
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo, now: time.Now().UTC}
}

func (s *Service) ListCategories(ctx context.Context, userID, ledgerID string, direction *string, active *bool) ([]Category, error) {
	if err := s.requireRole(ctx, ledgerID, userID, "viewer"); err != nil {
		return nil, err
	}
	return s.repo.ListCategories(ctx, ledgerID, direction, active)
}

func (s *Service) GetCategory(ctx context.Context, userID, ledgerID, categoryID string) (Category, error) {
	if err := s.requireRole(ctx, ledgerID, userID, "viewer"); err != nil {
		return Category{}, err
	}
	category, err := s.repo.GetCategory(ctx, ledgerID, categoryID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return Category{}, ErrCategoryNotFound
		}
		return Category{}, s.mapAccessError(ctx, ledgerID, err)
	}
	return category, nil
}

func (s *Service) CreateCategory(ctx context.Context, userID, ledgerID string, input CreateCategoryParams) (Category, error) {
	if err := s.requireRole(ctx, ledgerID, userID, "editor"); err != nil {
		return Category{}, err
	}

	name := strings.TrimSpace(input.Name)
	if name == "" {
		return Category{}, NewError("VALIDATION_ERROR", "Validacao falhou", map[string]string{"name": "required"})
	}

	direction := strings.TrimSpace(input.Direction)
	if !isValidDirection(direction) {
		return Category{}, NewError("VALIDATION_ERROR", "Validacao falhou", map[string]string{"direction": "invalid"})
	}

	parentID := input.ParentID
	if parentID != nil && strings.TrimSpace(*parentID) == "" {
		parentID = nil
	}

	if parentID != nil {
		parent, err := s.repo.GetCategory(ctx, ledgerID, *parentID)
		if err != nil {
			return Category{}, s.mapAccessError(ctx, ledgerID, err)
		}
		if parent.LedgerID != ledgerID {
			return Category{}, ErrAccessDenied
		}
	}

	flags := applyDirectionFlags(direction, input.IsBudgetBase, input.IsBudgetRelevant)

	created, err := s.repo.CreateCategory(ctx, CreateCategoryParams{
		LedgerID:         ledgerID,
		ParentID:         parentID,
		Name:             name,
		Direction:        direction,
		IsBudgetBase:     flags.IsBudgetBase,
		IsBudgetRelevant: flags.IsBudgetRelevant,
		IsActive:         input.IsActive,
	})
	if err != nil {
		if errors.Is(err, ErrDuplicateName) {
			return Category{}, ErrDuplicateName
		}
		return Category{}, err
	}
	return created, nil
}

func (s *Service) UpdateCategory(ctx context.Context, userID, ledgerID, categoryID string, input UpdateCategoryParams) (Category, error) {
	if err := s.requireRole(ctx, ledgerID, userID, "editor"); err != nil {
		return Category{}, err
	}

	name := strings.TrimSpace(input.Name)
	if name == "" {
		return Category{}, NewError("VALIDATION_ERROR", "Validacao falhou", map[string]string{"name": "required"})
	}

	parentID := input.ParentID
	if parentID != nil && strings.TrimSpace(*parentID) == "" {
		parentID = nil
	}

	current, err := s.repo.GetCategory(ctx, ledgerID, categoryID)
	if err != nil {
		return Category{}, s.mapAccessError(ctx, ledgerID, err)
	}

	if parentID != nil {
		if *parentID == categoryID {
			return Category{}, NewError("VALIDATION_ERROR", "Validacao falhou", map[string]string{"parent_id": "self"})
		}
		parent, err := s.repo.GetCategory(ctx, ledgerID, *parentID)
		if err != nil {
			return Category{}, s.mapAccessError(ctx, ledgerID, err)
		}
		if parent.LedgerID != ledgerID {
			return Category{}, ErrAccessDenied
		}
	}

	flags := applyDirectionFlags(current.Direction, input.IsBudgetBase, input.IsBudgetRelevant)

	updated, err := s.repo.UpdateCategory(ctx, UpdateCategoryParams{
		LedgerID:         ledgerID,
		CategoryID:       categoryID,
		ParentID:         parentID,
		Name:             name,
		IsBudgetBase:     flags.IsBudgetBase,
		IsBudgetRelevant: flags.IsBudgetRelevant,
		IsActive:         input.IsActive,
		UpdatedAt:        s.now(),
	})
	if err != nil {
		if errors.Is(err, ErrDuplicateName) {
			return Category{}, ErrDuplicateName
		}
		if errors.Is(err, ErrNotFound) {
			return Category{}, ErrCategoryNotFound
		}
		return Category{}, err
	}
	return updated, nil
}

func (s *Service) DeleteCategory(ctx context.Context, userID, ledgerID, categoryID string) error {
	if err := s.requireRole(ctx, ledgerID, userID, "editor"); err != nil {
		return err
	}

	used, err := s.repo.IsCategoryUsed(ctx, ledgerID, categoryID)
	if err != nil {
		return err
	}
	if used {
		return NewError("VALIDATION_ERROR", "Validacao falhou", map[string]string{"category_id": "in_use"})
	}

	if err := s.repo.DeactivateCategory(ctx, ledgerID, categoryID, s.now()); err != nil {
		if errors.Is(err, ErrNotFound) {
			return ErrCategoryNotFound
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

func isValidDirection(direction string) bool {
	switch direction {
	case "in", "out":
		return true
	default:
		return false
	}
}

func applyDirectionFlags(direction string, isBudgetBase bool, isBudgetRelevant bool) struct {
	IsBudgetBase     bool
	IsBudgetRelevant bool
} {
	flags := struct {
		IsBudgetBase     bool
		IsBudgetRelevant bool
	}{IsBudgetBase: isBudgetBase, IsBudgetRelevant: isBudgetRelevant}

	if direction == "in" {
		flags.IsBudgetRelevant = false
	}
	if direction == "out" {
		flags.IsBudgetBase = false
	}
	return flags
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
