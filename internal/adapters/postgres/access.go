package postgres

import (
	"context"
	"errors"

	"github.com/LucasSiedschlag/HausHaltsMeister/internal/domain/ledger"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func LedgerExists(ctx context.Context, pool *pgxpool.Pool, ledgerID string) (bool, error) {
	var exists bool
	row := pool.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM ledgers WHERE id = $1)`, ledgerID)
	if err := row.Scan(&exists); err != nil {
		return false, err
	}
	return exists, nil
}

func GetLedgerRole(ctx context.Context, pool *pgxpool.Pool, ledgerID, userID string) (string, error) {
	var role string
	row := pool.QueryRow(ctx, `
		SELECT CASE WHEN l.owner_user_id = $1 THEN 'owner' ELSE lm.role END AS role
		FROM ledgers l
		LEFT JOIN ledger_members lm ON lm.ledger_id = l.id AND lm.user_id = $1
		WHERE l.id = $2 AND (l.owner_user_id = $1 OR lm.user_id = $1)
	`, userID, ledgerID)
	if err := row.Scan(&role); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", ledger.ErrNotFound
		}
		return "", err
	}
	return role, nil
}
