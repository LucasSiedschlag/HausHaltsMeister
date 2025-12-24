package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/seuuser/cashflow/internal/adapters/postgres/sqlc"
	"github.com/seuuser/cashflow/internal/domain/category"
)

type CategoryRepository struct {
	q *sqlc.Queries
}

func NewCategoryRepository(db *pgxpool.Pool) *CategoryRepository {
	return &CategoryRepository{
		q: sqlc.New(db),
	}
}

// Support pgxpool connection as well (since sqlc.New accepts DBTX interface usually or we pass the specialized interface)
// Ideally sqlc generated code expects an interface. Let's see what sqlc generated.
// Usually it generates `New(db DBTX)`.
// We will assume `db` satisfies the interface. But for `pgxpool`, we might need to adjust.
// Let's stick to `*pgxpool.Pool` or similar if we can, but simpler is to use the generated interface if possible.
// For now, let's assume we pass something compatible.

func NewCategoryRepositoryWithQuerier(q *sqlc.Queries) *CategoryRepository {
	return &CategoryRepository{q: q}
}

func (r *CategoryRepository) Create(ctx context.Context, c *category.Category) (*category.Category, error) {
	params := sqlc.CreateCategoryParams{
		Name:             c.Name,
		Direction:        c.Direction,
		IsBudgetRelevant: c.IsBudgetRelevant,
		IsActive:         c.IsActive,
	}

	row, err := r.q.CreateCategory(ctx, params)
	if err != nil {
		return nil, err
	}

	return &category.Category{
		ID:               row.CategoryID,
		Name:             row.Name,
		Direction:        row.Direction,
		IsBudgetRelevant: row.IsBudgetRelevant,
		IsActive:         row.IsActive,
	}, nil
}

func (r *CategoryRepository) List(ctx context.Context, activeOnly bool) ([]*category.Category, error) {
	// If activeOnly is true, we pass activeOnly=false IS FALSE -> which is confusing.
	// Looking at query: WHERE ($1::boolean = false OR is_active = true)
	// If activeOnly is true: $1=true. WHERE (true=false OR is_active=true) -> WHERE is_active=true. Correct.
	// If activeOnly is false: $1=false. WHERE (false=false OR is_active=true) -> WHERE true. Correct.

	rows, err := r.q.ListCategories(ctx, activeOnly)
	if err != nil {
		return nil, err
	}

	cats := make([]*category.Category, len(rows))
	for i, row := range rows {
		cats[i] = &category.Category{
			ID:               row.CategoryID,
			Name:             row.Name,
			Direction:        row.Direction,
			IsBudgetRelevant: row.IsBudgetRelevant,
			IsActive:         row.IsActive,
		}
	}
	return cats, nil
}

func (r *CategoryRepository) GetByID(ctx context.Context, id int32) (*category.Category, error) {
	row, err := r.q.GetCategoryByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil // Or explicit not found error
		}
		return nil, err
	}

	return &category.Category{
		ID:               row.CategoryID,
		Name:             row.Name,
		Direction:        row.Direction,
		IsBudgetRelevant: row.IsBudgetRelevant,
		IsActive:         row.IsActive,
	}, nil
}
