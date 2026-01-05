package categories

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/LucasSiedschlag/HausHaltsMeister/internal/adapters/postgres"
	"github.com/LucasSiedschlag/HausHaltsMeister/internal/domain/categories"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(store *postgres.Store) *Repository {
	return &Repository{pool: store.Pool()}
}

func (r *Repository) ListCategories(ctx context.Context, ledgerID string, direction *string, active *bool) ([]categories.Category, error) {
	query := `
		SELECT id, ledger_id, parent_id, name, direction, is_budget_base, is_budget_relevant, is_active, created_at, updated_at
		FROM categories
		WHERE ledger_id = $1
	`
	args := []interface{}{ledgerID}
	idx := 2
	if direction != nil {
		query += " AND direction = $" + strconv.Itoa(idx)
		args = append(args, *direction)
		idx++
	}
	if active != nil {
		query += " AND is_active = $" + strconv.Itoa(idx)
		args = append(args, *active)
		idx++
	}
	query += " ORDER BY created_at"

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []categories.Category
	for rows.Next() {
		var item categories.Category
		if err := rows.Scan(
			&item.ID,
			&item.LedgerID,
			&item.ParentID,
			&item.Name,
			&item.Direction,
			&item.IsBudgetBase,
			&item.IsBudgetRelevant,
			&item.IsActive,
			&item.CreatedAt,
			&item.UpdatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}

func (r *Repository) GetCategory(ctx context.Context, ledgerID, categoryID string) (categories.Category, error) {
	var item categories.Category
	row := r.pool.QueryRow(ctx, `
		SELECT id, ledger_id, parent_id, name, direction, is_budget_base, is_budget_relevant, is_active, created_at, updated_at
		FROM categories
		WHERE ledger_id = $1 AND id = $2
	`, ledgerID, categoryID)
	if err := row.Scan(
		&item.ID,
		&item.LedgerID,
		&item.ParentID,
		&item.Name,
		&item.Direction,
		&item.IsBudgetBase,
		&item.IsBudgetRelevant,
		&item.IsActive,
		&item.CreatedAt,
		&item.UpdatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return categories.Category{}, categories.ErrNotFound
		}
		return categories.Category{}, err
	}
	return item, nil
}

func (r *Repository) CreateCategory(ctx context.Context, params categories.CreateCategoryParams) (categories.Category, error) {
	var item categories.Category
	row := r.pool.QueryRow(ctx, `
		INSERT INTO categories (ledger_id, parent_id, name, direction, is_budget_base, is_budget_relevant, is_active)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, ledger_id, parent_id, name, direction, is_budget_base, is_budget_relevant, is_active, created_at, updated_at
	`, params.LedgerID, params.ParentID, params.Name, params.Direction, params.IsBudgetBase, params.IsBudgetRelevant, params.IsActive)
	if err := row.Scan(
		&item.ID,
		&item.LedgerID,
		&item.ParentID,
		&item.Name,
		&item.Direction,
		&item.IsBudgetBase,
		&item.IsBudgetRelevant,
		&item.IsActive,
		&item.CreatedAt,
		&item.UpdatedAt,
	); err != nil {
		if postgres.IsUniqueViolation(err) {
			return categories.Category{}, categories.ErrDuplicateName
		}
		return categories.Category{}, err
	}
	return item, nil
}

func (r *Repository) UpdateCategory(ctx context.Context, params categories.UpdateCategoryParams) (categories.Category, error) {
	var item categories.Category
	row := r.pool.QueryRow(ctx, `
		UPDATE categories
		SET parent_id = $3, name = $4, is_budget_base = $5, is_budget_relevant = $6, is_active = $7, updated_at = $8
		WHERE ledger_id = $1 AND id = $2
		RETURNING id, ledger_id, parent_id, name, direction, is_budget_base, is_budget_relevant, is_active, created_at, updated_at
	`, params.LedgerID, params.CategoryID, params.ParentID, params.Name, params.IsBudgetBase, params.IsBudgetRelevant, params.IsActive, params.UpdatedAt)
	if err := row.Scan(
		&item.ID,
		&item.LedgerID,
		&item.ParentID,
		&item.Name,
		&item.Direction,
		&item.IsBudgetBase,
		&item.IsBudgetRelevant,
		&item.IsActive,
		&item.CreatedAt,
		&item.UpdatedAt,
	); err != nil {
		if postgres.IsUniqueViolation(err) {
			return categories.Category{}, categories.ErrDuplicateName
		}
		if errors.Is(err, pgx.ErrNoRows) {
			return categories.Category{}, categories.ErrNotFound
		}
		return categories.Category{}, err
	}
	return item, nil
}

func (r *Repository) DeactivateCategory(ctx context.Context, ledgerID, categoryID string, updatedAt time.Time) error {
	cmd, err := r.pool.Exec(ctx, `
		UPDATE categories
		SET is_active = false, updated_at = $3
		WHERE ledger_id = $1 AND id = $2
	`, ledgerID, categoryID, updatedAt)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		return categories.ErrNotFound
	}
	return nil
}

func (r *Repository) GetLedgerRole(ctx context.Context, ledgerID, userID string) (string, error) {
	return postgres.GetLedgerRole(ctx, r.pool, ledgerID, userID)
}

func (r *Repository) LedgerExists(ctx context.Context, ledgerID string) (bool, error) {
	return postgres.LedgerExists(ctx, r.pool, ledgerID)
}

func (r *Repository) IsCategoryUsed(ctx context.Context, ledgerID, categoryID string) (bool, error) {
	var exists bool
	row := r.pool.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM entries
			WHERE ledger_id = $1 AND category_id = $2
		)
	`, ledgerID, categoryID)
	if err := row.Scan(&exists); err != nil {
		return false, err
	}
	return exists, nil
}
