package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/LucasSiedschlag/HausHaltsMeister/internal/domain/investments"
	"github.com/jackc/pgx/v5"
)

func (s *Store) FindAccountByType(ctx context.Context, ledgerID, accountType string) (string, error) {
	var id string
	row := s.pool.QueryRow(ctx, `
		SELECT id FROM accounts
		WHERE ledger_id = $1 AND type = $2 AND is_active = true
		ORDER BY created_at ASC
		LIMIT 1
	`, ledgerID, accountType)
	if err := row.Scan(&id); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", investments.ErrNotFound
		}
		return "", err
	}
	return id, nil
}

func (s *Store) FindCategoryByName(ctx context.Context, ledgerID, name string) (string, error) {
	var id string
	row := s.pool.QueryRow(ctx, `
		SELECT id FROM categories
		WHERE ledger_id = $1 AND name = $2
		LIMIT 1
	`, ledgerID, name)
	if err := row.Scan(&id); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", investments.ErrNotFound
		}
		return "", err
	}
	return id, nil
}

func (s *Store) GetCategoryDirection(ctx context.Context, ledgerID, categoryID string) (string, error) {
	var direction string
	row := s.pool.QueryRow(ctx, `
		SELECT direction FROM categories
		WHERE ledger_id = $1 AND id = $2
	`, ledgerID, categoryID)
	if err := row.Scan(&direction); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", investments.ErrNotFound
		}
		return "", err
	}
	return direction, nil
}

func (s *Store) SumByCategory(ctx context.Context, ledgerID, categoryID string, from, to time.Time) (int64, error) {
	var total int64
	row := s.pool.QueryRow(ctx, `
		SELECT COALESCE(SUM(e.amount_cents), 0)
		FROM entries e
		JOIN transactions t ON t.id = e.transaction_id
		WHERE e.ledger_id = $1
			AND e.category_id = $2
			AND t.occurred_at >= $3 AND t.occurred_at <= $4
	`, ledgerID, categoryID, from, to)
	if err := row.Scan(&total); err != nil {
		return 0, err
	}
	return total, nil
}
