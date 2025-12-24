package category

import (
	"context"
	"fmt"
)

type CategoryService struct {
	repo Repository
}

func NewService(repo Repository) *CategoryService {
	return &CategoryService{repo: repo}
}

func (s *CategoryService) CreateCategory(ctx context.Context, name, direction string, isBudgetRelevant bool) (*Category, error) {
	newCat, err := New(name, direction, isBudgetRelevant)
	if err != nil {
		return nil, fmt.Errorf("validation error: %w", err)
	}

	createdCat, err := s.repo.Create(ctx, newCat)
	if err != nil {
		return nil, fmt.Errorf("repository error: %w", err)
	}
	return createdCat, nil
}

func (s *CategoryService) ListCategories(ctx context.Context, activeOnly bool) ([]*Category, error) {
	return s.repo.List(ctx, activeOnly)
}

func (s *CategoryService) DeactivateCategory(ctx context.Context, id int32) error {
	_, err := s.repo.Update(ctx, id, false)
	if err != nil {
		return fmt.Errorf("failed to deactivate category: %w", err)
	}
	return nil
}
