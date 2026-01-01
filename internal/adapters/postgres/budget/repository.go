package budget

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/LucasSiedschlag/HausHaltsMeister/internal/adapters/postgres"
	"github.com/LucasSiedschlag/HausHaltsMeister/internal/domain/budget"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(store *postgres.Store) *Repository {
	return &Repository{pool: store.Pool()}
}

func (r *Repository) GetPlanByLedger(ctx context.Context, ledgerID string) (budget.Plan, error) {
	var plan budget.Plan
	row := r.pool.QueryRow(ctx, `
		SELECT id, ledger_id, name, created_at, updated_at
		FROM budget_plans
		WHERE ledger_id = $1
	`, ledgerID)
	if err := row.Scan(&plan.ID, &plan.LedgerID, &plan.Name, &plan.CreatedAt, &plan.UpdatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return budget.Plan{}, budget.ErrNotFound
		}
		return budget.Plan{}, err
	}
	return plan, nil
}

func (r *Repository) CreatePlan(ctx context.Context, ledgerID, name string) (budget.Plan, error) {
	var plan budget.Plan
	row := r.pool.QueryRow(ctx, `
		INSERT INTO budget_plans (ledger_id, name)
		VALUES ($1, $2)
		RETURNING id, ledger_id, name, created_at, updated_at
	`, ledgerID, name)
	if err := row.Scan(&plan.ID, &plan.LedgerID, &plan.Name, &plan.CreatedAt, &plan.UpdatedAt); err != nil {
		return budget.Plan{}, err
	}
	return plan, nil
}

func (r *Repository) UpdatePlan(ctx context.Context, ledgerID, name string, updatedAt time.Time) (budget.Plan, error) {
	var plan budget.Plan
	row := r.pool.QueryRow(ctx, `
		UPDATE budget_plans
		SET name = $2, updated_at = $3
		WHERE ledger_id = $1
		RETURNING id, ledger_id, name, created_at, updated_at
	`, ledgerID, name, updatedAt)
	if err := row.Scan(&plan.ID, &plan.LedgerID, &plan.Name, &plan.CreatedAt, &plan.UpdatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return budget.Plan{}, budget.ErrNotFound
		}
		return budget.Plan{}, err
	}
	return plan, nil
}

func (r *Repository) ListVersions(ctx context.Context, ledgerID string, from, to *time.Time) ([]budget.Version, error) {
	query := `
		SELECT v.id, v.plan_id, v.effective_from_month, v.created_by_user_id, v.created_at, v.updated_at
		FROM budget_plan_versions v
		JOIN budget_plans p ON p.id = v.plan_id
		WHERE p.ledger_id = $1
	`
	args := []interface{}{ledgerID}
	idx := 2
	if from != nil {
		query += " AND v.effective_from_month >= $" + strconv.Itoa(idx)
		args = append(args, *from)
		idx++
	}
	if to != nil {
		query += " AND v.effective_from_month <= $" + strconv.Itoa(idx)
		args = append(args, *to)
		idx++
	}
	query += " ORDER BY v.effective_from_month DESC"

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var versions []budget.Version
	for rows.Next() {
		var version budget.Version
		if err := rows.Scan(&version.ID, &version.PlanID, &version.EffectiveFromMonth, &version.CreatedByUserID, &version.CreatedAt, &version.UpdatedAt); err != nil {
			return nil, err
		}
		versions = append(versions, version)
	}
	return versions, nil
}

func (r *Repository) GetVersion(ctx context.Context, ledgerID, versionID string) (budget.Version, error) {
	var version budget.Version
	row := r.pool.QueryRow(ctx, `
		SELECT v.id, v.plan_id, v.effective_from_month, v.created_by_user_id, v.created_at, v.updated_at
		FROM budget_plan_versions v
		JOIN budget_plans p ON p.id = v.plan_id
		WHERE p.ledger_id = $1 AND v.id = $2
	`, ledgerID, versionID)
	if err := row.Scan(&version.ID, &version.PlanID, &version.EffectiveFromMonth, &version.CreatedByUserID, &version.CreatedAt, &version.UpdatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return budget.Version{}, budget.ErrNotFound
		}
		return budget.Version{}, err
	}
	return version, nil
}

func (r *Repository) CreateVersionWithLines(ctx context.Context, ledgerID, planID, userID string, effectiveFrom time.Time, lines []budget.LineInput) (budget.Version, error) {
	var version budget.Version
	err := postgres.WithTx(ctx, r.pool, func(tx pgx.Tx) error {
		row := tx.QueryRow(ctx, `
			INSERT INTO budget_plan_versions (plan_id, effective_from_month, created_by_user_id)
			VALUES ($1, $2, $3)
			RETURNING id, plan_id, effective_from_month, created_by_user_id, created_at, updated_at
		`, planID, effectiveFrom, userID)
		if err := row.Scan(&version.ID, &version.PlanID, &version.EffectiveFromMonth, &version.CreatedByUserID, &version.CreatedAt, &version.UpdatedAt); err != nil {
			return err
		}

		for _, line := range lines {
			_, err := tx.Exec(ctx, `
				INSERT INTO budget_plan_lines (version_id, category_id, percent, include_children)
				VALUES ($1, $2, $3, $4)
			`, version.ID, line.CategoryID, line.Percent, line.IncludeChildren)
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return budget.Version{}, err
	}

	return version, nil
}

func (r *Repository) AddLine(ctx context.Context, versionID string, line budget.LineInput) (budget.Line, error) {
	var created budget.Line
	row := r.pool.QueryRow(ctx, `
		INSERT INTO budget_plan_lines (version_id, category_id, percent, include_children)
		VALUES ($1, $2, $3, $4)
		RETURNING id, version_id, category_id, percent, include_children, created_at, updated_at
	`, versionID, line.CategoryID, line.Percent, line.IncludeChildren)
	if err := row.Scan(&created.ID, &created.VersionID, &created.CategoryID, &created.Percent, &created.IncludeChildren, &created.CreatedAt, &created.UpdatedAt); err != nil {
		return budget.Line{}, err
	}
	return created, nil
}

func (r *Repository) UpdateLine(ctx context.Context, lineID string, percent float64, includeChildren bool, updatedAt time.Time) (budget.Line, error) {
	var updated budget.Line
	row := r.pool.QueryRow(ctx, `
		UPDATE budget_plan_lines
		SET percent = $2, include_children = $3, updated_at = $4
		WHERE id = $1
		RETURNING id, version_id, category_id, percent, include_children, created_at, updated_at
	`, lineID, percent, includeChildren, updatedAt)
	if err := row.Scan(&updated.ID, &updated.VersionID, &updated.CategoryID, &updated.Percent, &updated.IncludeChildren, &updated.CreatedAt, &updated.UpdatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return budget.Line{}, budget.ErrNotFound
		}
		return budget.Line{}, err
	}
	return updated, nil
}

func (r *Repository) DeleteLine(ctx context.Context, lineID string) error {
	cmd, err := r.pool.Exec(ctx, `DELETE FROM budget_plan_lines WHERE id = $1`, lineID)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		return budget.ErrNotFound
	}
	return nil
}

func (r *Repository) GetLinesByVersion(ctx context.Context, versionID string) ([]budget.Line, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, version_id, category_id, percent, include_children, created_at, updated_at
		FROM budget_plan_lines
		WHERE version_id = $1
		ORDER BY created_at
	`, versionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var lines []budget.Line
	for rows.Next() {
		var line budget.Line
		if err := rows.Scan(&line.ID, &line.VersionID, &line.CategoryID, &line.Percent, &line.IncludeChildren, &line.CreatedAt, &line.UpdatedAt); err != nil {
			return nil, err
		}
		lines = append(lines, line)
	}
	return lines, nil
}

func (r *Repository) GetApplicableVersion(ctx context.Context, ledgerID string, month time.Time) (budget.Version, error) {
	var version budget.Version
	row := r.pool.QueryRow(ctx, `
		SELECT v.id, v.plan_id, v.effective_from_month, v.created_by_user_id, v.created_at, v.updated_at
		FROM budget_plan_versions v
		JOIN budget_plans p ON p.id = v.plan_id
		WHERE p.ledger_id = $1 AND v.effective_from_month <= $2
		ORDER BY v.effective_from_month DESC
		LIMIT 1
	`, ledgerID, month)
	if err := row.Scan(&version.ID, &version.PlanID, &version.EffectiveFromMonth, &version.CreatedByUserID, &version.CreatedAt, &version.UpdatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return budget.Version{}, budget.ErrNotFound
		}
		return budget.Version{}, err
	}
	return version, nil
}

func (r *Repository) GetCategoryInfo(ctx context.Context, ledgerID string, categoryIDs []string) (map[string]budget.CategoryInfo, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, direction, is_budget_relevant
		FROM categories
		WHERE ledger_id = $1 AND id = ANY($2)
	`, ledgerID, categoryIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := map[string]budget.CategoryInfo{}
	for rows.Next() {
		var id string
		var direction string
		var relevant bool
		if err := rows.Scan(&id, &direction, &relevant); err != nil {
			return nil, err
		}
		result[id] = budget.CategoryInfo{Direction: direction, IsBudgetRelevant: relevant}
	}
	return result, nil
}

func (r *Repository) GetCategoryDescendants(ctx context.Context, ledgerID, categoryID string) ([]string, error) {
	rows, err := r.pool.Query(ctx, `
		WITH RECURSIVE tree AS (
			SELECT id FROM categories WHERE ledger_id = $1 AND id = $2
			UNION ALL
			SELECT c.id FROM categories c JOIN tree t ON c.parent_id = t.id WHERE c.ledger_id = $1
		)
		SELECT id FROM tree WHERE id <> $2
	`, ledgerID, categoryID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, nil
}

func (r *Repository) IncomeBaseForMonth(ctx context.Context, ledgerID string, month time.Time) (int64, error) {
	start := month
	end := month.AddDate(0, 1, 0)
	var total int64
	row := r.pool.QueryRow(ctx, `
		SELECT COALESCE(SUM(e.amount_cents), 0)
		FROM entries e
		JOIN categories c ON c.id = e.category_id
		JOIN transactions t ON t.id = e.transaction_id
		WHERE e.ledger_id = $1
			AND t.occurred_at >= $2 AND t.occurred_at < $3
			AND c.direction = 'in'
			AND c.is_budget_base = true
	`, ledgerID, start, end)
	if err := row.Scan(&total); err != nil {
		return 0, err
	}
	return total, nil
}

func (r *Repository) SpentActualForMonth(ctx context.Context, ledgerID string, categoryIDs []string, month time.Time) (int64, error) {
	start := month
	end := month.AddDate(0, 1, 0)
	var total int64
	row := r.pool.QueryRow(ctx, `
		SELECT COALESCE(SUM(e.amount_cents), 0)
		FROM entries e
		JOIN categories c ON c.id = e.category_id
		JOIN transactions t ON t.id = e.transaction_id
		WHERE e.ledger_id = $1
			AND t.occurred_at >= $2 AND t.occurred_at < $3
			AND c.direction = 'out'
			AND c.is_budget_relevant = true
			AND c.id = ANY($4)
	`, ledgerID, start, end, categoryIDs)
	if err := row.Scan(&total); err != nil {
		return 0, err
	}
	return total, nil
}

func (r *Repository) OutsideBudgetForMonth(ctx context.Context, ledgerID string, month time.Time) (int64, error) {
	start := month
	end := month.AddDate(0, 1, 0)
	var total int64
	row := r.pool.QueryRow(ctx, `
		SELECT COALESCE(SUM(e.amount_cents), 0)
		FROM entries e
		JOIN categories c ON c.id = e.category_id
		JOIN transactions t ON t.id = e.transaction_id
		WHERE e.ledger_id = $1
			AND t.occurred_at >= $2 AND t.occurred_at < $3
			AND c.direction = 'out'
			AND c.is_budget_relevant = false
	`, ledgerID, start, end)
	if err := row.Scan(&total); err != nil {
		return 0, err
	}
	return total, nil
}

func (r *Repository) GetLedgerRole(ctx context.Context, ledgerID, userID string) (string, error) {
	return postgres.GetLedgerRole(ctx, r.pool, ledgerID, userID)
}

func (r *Repository) LedgerExists(ctx context.Context, ledgerID string) (bool, error) {
	return postgres.LedgerExists(ctx, r.pool, ledgerID)
}
