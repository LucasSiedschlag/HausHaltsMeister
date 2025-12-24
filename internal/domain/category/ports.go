package category

import "context"

type Repository interface {
	Create(ctx context.Context, category *Category) (*Category, error)
	List(ctx context.Context, activeOnly bool) ([]*Category, error)
	GetByID(ctx context.Context, id int32) (*Category, error)
	Update(ctx context.Context, id int32, isActive bool) (*Category, error)
}

type Service interface {
	CreateCategory(ctx context.Context, name, direction string, isBudgetRelevant bool) (*Category, error)
	ListCategories(ctx context.Context, activeOnly bool) ([]*Category, error)
	DeactivateCategory(ctx context.Context, id int32) error
}
