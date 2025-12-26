package category

import (
	"context"
	"time"
)

type Repository interface {
	Create(ctx context.Context, category *Category) (*Category, error)
	List(ctx context.Context, activeOnly bool) ([]*Category, error)
	ListByMonth(ctx context.Context, activeOnly bool, month time.Time) ([]*Category, error)
	GetByID(ctx context.Context, id int32) (*Category, error)
	Update(ctx context.Context, category *Category) (*Category, error)
	Deactivate(ctx context.Context, id int32, inactiveFromMonth time.Time) (*Category, error)
}

type Service interface {
	CreateCategory(ctx context.Context, name, direction string, isBudgetRelevant bool) (*Category, error)
	ListCategories(ctx context.Context, activeOnly bool) ([]*Category, error)
	ListCategoriesByMonth(ctx context.Context, activeOnly bool, month time.Time) ([]*Category, error)
	DeactivateCategory(ctx context.Context, id int32, inactiveFromMonth time.Time) error
	UpdateCategory(ctx context.Context, id int32, name, direction string, isBudgetRelevant, isActive bool) (*Category, error)
}
