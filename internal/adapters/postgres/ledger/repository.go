package ledger

import (
	"context"
	"errors"
	"time"

	"github.com/LucasSiedschlag/HausHaltsMeister/internal/adapters/postgres"
	"github.com/LucasSiedschlag/HausHaltsMeister/internal/domain/ledger"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(store *postgres.Store) *Repository {
	return &Repository{pool: store.Pool()}
}

func (r *Repository) ListLedgersForUser(ctx context.Context, userID string) ([]ledger.LedgerWithRole, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT l.id, l.owner_user_id, l.name, l.currency_code, l.created_at, l.updated_at,
			CASE WHEN l.owner_user_id = $1 THEN 'owner' ELSE lm.role END AS role
		FROM ledgers l
		LEFT JOIN ledger_members lm ON lm.ledger_id = l.id AND lm.user_id = $1 AND lm.removed_at IS NULL
		WHERE l.owner_user_id = $1 OR lm.user_id = $1
		ORDER BY l.created_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []ledger.LedgerWithRole
	for rows.Next() {
		var item ledger.LedgerWithRole
		if err := rows.Scan(
			&item.Ledger.ID,
			&item.Ledger.OwnerUserID,
			&item.Ledger.Name,
			&item.Ledger.CurrencyCode,
			&item.Ledger.CreatedAt,
			&item.Ledger.UpdatedAt,
			&item.Role,
		); err != nil {
			return nil, err
		}
		results = append(results, item)
	}
	return results, nil
}

func (r *Repository) CreateLedger(ctx context.Context, params ledger.CreateLedgerParams) (ledger.Ledger, error) {
	var created ledger.Ledger
	err := postgres.WithTx(ctx, r.pool, func(tx pgx.Tx) error {
		row := tx.QueryRow(ctx, `
			INSERT INTO ledgers (owner_user_id, name, currency_code)
			VALUES ($1, $2, $3)
			RETURNING id, owner_user_id, name, currency_code, created_at, updated_at
		`, params.OwnerUserID, params.Name, params.CurrencyCode)
		if err := scanLedger(row, &created); err != nil {
			return err
		}

		_, err := tx.Exec(ctx, `
			INSERT INTO ledger_members (ledger_id, user_id, role)
			VALUES ($1, $2, 'owner')
		`, created.ID, params.OwnerUserID)
		return err
	})
	if err != nil {
		return ledger.Ledger{}, err
	}
	return created, nil
}

func (r *Repository) GetLedgerForUser(ctx context.Context, ledgerID, userID string) (ledger.LedgerWithRole, error) {
	var result ledger.LedgerWithRole
	row := r.pool.QueryRow(ctx, `
		SELECT l.id, l.owner_user_id, l.name, l.currency_code, l.created_at, l.updated_at,
			CASE WHEN l.owner_user_id = $1 THEN 'owner' ELSE lm.role END AS role
		FROM ledgers l
		LEFT JOIN ledger_members lm ON lm.ledger_id = l.id AND lm.user_id = $1 AND lm.removed_at IS NULL
		WHERE l.id = $2 AND (l.owner_user_id = $1 OR lm.user_id = $1)
	`, userID, ledgerID)
	if err := scanLedgerWithRole(row, &result); err != nil {
		return ledger.LedgerWithRole{}, err
	}
	return result, nil
}

func (r *Repository) GetLedgerByID(ctx context.Context, ledgerID string) (ledger.Ledger, error) {
	var result ledger.Ledger
	row := r.pool.QueryRow(ctx, `
		SELECT id, owner_user_id, name, currency_code, created_at, updated_at
		FROM ledgers
		WHERE id = $1
	`, ledgerID)
	if err := scanLedger(row, &result); err != nil {
		return ledger.Ledger{}, err
	}
	return result, nil
}

func (r *Repository) UpdateLedger(ctx context.Context, ledgerID, name string, updatedAt time.Time) (ledger.Ledger, error) {
	var result ledger.Ledger
	row := r.pool.QueryRow(ctx, `
		UPDATE ledgers
		SET name = $2, updated_at = $3
		WHERE id = $1
		RETURNING id, owner_user_id, name, currency_code, created_at, updated_at
	`, ledgerID, name, updatedAt)
	if err := scanLedger(row, &result); err != nil {
		return ledger.Ledger{}, err
	}
	return result, nil
}

func (r *Repository) LedgerExists(ctx context.Context, ledgerID string) (bool, error) {
	var exists bool
	row := r.pool.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM ledgers WHERE id = $1)`, ledgerID)
	if err := row.Scan(&exists); err != nil {
		return false, err
	}
	return exists, nil
}

func (r *Repository) GetLedgerRole(ctx context.Context, ledgerID, userID string) (string, error) {
	var role string
	row := r.pool.QueryRow(ctx, `
		SELECT CASE WHEN l.owner_user_id = $1 THEN 'owner' ELSE lm.role END AS role
		FROM ledgers l
		LEFT JOIN ledger_members lm ON lm.ledger_id = l.id AND lm.user_id = $1 AND lm.removed_at IS NULL
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

func (r *Repository) ListMembers(ctx context.Context, ledgerID string) ([]ledger.Member, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT lm.ledger_id, lm.user_id, lm.role, lm.created_at, lm.updated_at,
			COALESCE(u.display_name, '') AS display_name,
			u.email,
			u.avatar_url
		FROM ledger_members lm
		JOIN users u ON u.id = lm.user_id
		WHERE lm.ledger_id = $1 AND lm.removed_at IS NULL
		ORDER BY lm.created_at
	`, ledgerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []ledger.Member
	for rows.Next() {
		var member ledger.Member
		if err := scanMember(rows, &member); err != nil {
			return nil, err
		}
		members = append(members, member)
	}
	return members, nil
}

func (r *Repository) AddMember(ctx context.Context, ledgerID, userID, role string, updatedAt time.Time) (ledger.Member, error) {
	var member ledger.Member
	row := r.pool.QueryRow(ctx, `
		WITH upserted AS (
			INSERT INTO ledger_members (ledger_id, user_id, role, updated_at)
			VALUES ($1, $2, $3, $4)
			ON CONFLICT (ledger_id, user_id) DO UPDATE
			SET role = EXCLUDED.role,
				removed_at = NULL,
				updated_at = EXCLUDED.updated_at
			WHERE ledger_members.removed_at IS NOT NULL
			RETURNING ledger_id, user_id, role, created_at, updated_at
		)
		SELECT upserted.ledger_id, upserted.user_id, upserted.role, upserted.created_at, upserted.updated_at,
			COALESCE(u.display_name, '') AS display_name,
			u.email,
			u.avatar_url
		FROM upserted
		JOIN users u ON u.id = upserted.user_id
	`, ledgerID, userID, role, updatedAt)
	if err := scanMember(row, &member); err != nil {
		if errors.Is(err, pgx.ErrNoRows) || postgres.IsUniqueViolation(err) {
			return ledger.Member{}, ledger.ErrMemberExists
		}
		return ledger.Member{}, err
	}
	return member, nil
}

func (r *Repository) UpdateMemberRole(ctx context.Context, ledgerID, userID, role string, updatedAt time.Time) (ledger.Member, error) {
	var member ledger.Member
	row := r.pool.QueryRow(ctx, `
		WITH updated AS (
			UPDATE ledger_members
			SET role = $3, updated_at = $4
			WHERE ledger_id = $1 AND user_id = $2 AND removed_at IS NULL
			RETURNING ledger_id, user_id, role, created_at, updated_at
		)
		SELECT updated.ledger_id, updated.user_id, updated.role, updated.created_at, updated.updated_at,
			COALESCE(u.display_name, '') AS display_name,
			u.email,
			u.avatar_url
		FROM updated
		JOIN users u ON u.id = updated.user_id
	`, ledgerID, userID, role, updatedAt)
	if err := scanMember(row, &member); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ledger.Member{}, ledger.ErrNotFound
		}
		return ledger.Member{}, err
	}
	return member, nil
}

func (r *Repository) RemoveMember(ctx context.Context, ledgerID, userID string) error {
	cmd, err := r.pool.Exec(ctx, `
		UPDATE ledger_members
		SET removed_at = now(), updated_at = now()
		WHERE ledger_id = $1 AND user_id = $2 AND removed_at IS NULL
	`, ledgerID, userID)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		return ledger.ErrNotFound
	}
	return nil
}

func (r *Repository) GetUserIDByEmail(ctx context.Context, email string) (string, error) {
	var userID string
	row := r.pool.QueryRow(ctx, `SELECT id FROM users WHERE email = $1`, email)
	if err := row.Scan(&userID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", ledger.ErrUserNotFound
		}
		return "", err
	}
	return userID, nil
}

func scanLedger(row pgx.Row, ledgerItem *ledger.Ledger) error {
	if err := row.Scan(
		&ledgerItem.ID,
		&ledgerItem.OwnerUserID,
		&ledgerItem.Name,
		&ledgerItem.CurrencyCode,
		&ledgerItem.CreatedAt,
		&ledgerItem.UpdatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ledger.ErrNotFound
		}
		return err
	}
	return nil
}

func scanMember(row pgx.Row, member *ledger.Member) error {
	if err := row.Scan(
		&member.LedgerID,
		&member.UserID,
		&member.Role,
		&member.CreatedAt,
		&member.UpdatedAt,
		&member.DisplayName,
		&member.Email,
		&member.AvatarURL,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return pgx.ErrNoRows
		}
		return err
	}
	return nil
}

func scanLedgerWithRole(row pgx.Row, result *ledger.LedgerWithRole) error {
	if err := row.Scan(
		&result.Ledger.ID,
		&result.Ledger.OwnerUserID,
		&result.Ledger.Name,
		&result.Ledger.CurrencyCode,
		&result.Ledger.CreatedAt,
		&result.Ledger.UpdatedAt,
		&result.Role,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ledger.ErrNotFound
		}
		return err
	}
	return nil
}
