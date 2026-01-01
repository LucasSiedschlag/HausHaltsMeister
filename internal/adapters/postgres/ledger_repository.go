package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/LucasSiedschlag/HausHaltsMeister/internal/domain/ledger"
	"github.com/jackc/pgx/v5"
)

func (s *Store) ListLedgersForUser(ctx context.Context, userID string) ([]ledger.LedgerWithRole, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT l.id, l.owner_user_id, l.name, l.currency_code, l.created_at, l.updated_at,
			CASE WHEN l.owner_user_id = $1 THEN 'owner' ELSE lm.role END AS role
		FROM ledgers l
		LEFT JOIN ledger_members lm ON lm.ledger_id = l.id AND lm.user_id = $1
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

func (s *Store) CreateLedger(ctx context.Context, params ledger.CreateLedgerParams) (ledger.Ledger, error) {
	var created ledger.Ledger
	err := withTx(ctx, s.pool, func(tx pgx.Tx) error {
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

func (s *Store) GetLedgerForUser(ctx context.Context, ledgerID, userID string) (ledger.LedgerWithRole, error) {
	var result ledger.LedgerWithRole
	row := s.pool.QueryRow(ctx, `
		SELECT l.id, l.owner_user_id, l.name, l.currency_code, l.created_at, l.updated_at,
			CASE WHEN l.owner_user_id = $1 THEN 'owner' ELSE lm.role END AS role
		FROM ledgers l
		LEFT JOIN ledger_members lm ON lm.ledger_id = l.id AND lm.user_id = $1
		WHERE l.id = $2 AND (l.owner_user_id = $1 OR lm.user_id = $1)
	`, userID, ledgerID)
	if err := scanLedgerWithRole(row, &result); err != nil {
		return ledger.LedgerWithRole{}, err
	}
	return result, nil
}

func (s *Store) GetLedgerByID(ctx context.Context, ledgerID string) (ledger.Ledger, error) {
	var result ledger.Ledger
	row := s.pool.QueryRow(ctx, `
		SELECT id, owner_user_id, name, currency_code, created_at, updated_at
		FROM ledgers
		WHERE id = $1
	`, ledgerID)
	if err := scanLedger(row, &result); err != nil {
		return ledger.Ledger{}, err
	}
	return result, nil
}

func (s *Store) UpdateLedger(ctx context.Context, ledgerID, name string, updatedAt time.Time) (ledger.Ledger, error) {
	var result ledger.Ledger
	row := s.pool.QueryRow(ctx, `
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

func (s *Store) LedgerExists(ctx context.Context, ledgerID string) (bool, error) {
	var exists bool
	row := s.pool.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM ledgers WHERE id = $1)`, ledgerID)
	if err := row.Scan(&exists); err != nil {
		return false, err
	}
	return exists, nil
}

func (s *Store) GetLedgerRole(ctx context.Context, ledgerID, userID string) (string, error) {
	var role string
	row := s.pool.QueryRow(ctx, `
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

func (s *Store) ListMembers(ctx context.Context, ledgerID string) ([]ledger.Member, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT ledger_id, user_id, role, created_at, updated_at
		FROM ledger_members
		WHERE ledger_id = $1
		ORDER BY created_at
	`, ledgerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []ledger.Member
	for rows.Next() {
		var member ledger.Member
		if err := rows.Scan(&member.LedgerID, &member.UserID, &member.Role, &member.CreatedAt, &member.UpdatedAt); err != nil {
			return nil, err
		}
		members = append(members, member)
	}
	return members, nil
}

func (s *Store) AddMember(ctx context.Context, ledgerID, userID, role string, updatedAt time.Time) (ledger.Member, error) {
	var member ledger.Member
	row := s.pool.QueryRow(ctx, `
		INSERT INTO ledger_members (ledger_id, user_id, role, updated_at)
		VALUES ($1, $2, $3, $4)
		RETURNING ledger_id, user_id, role, created_at, updated_at
	`, ledgerID, userID, role, updatedAt)
	if err := row.Scan(&member.LedgerID, &member.UserID, &member.Role, &member.CreatedAt, &member.UpdatedAt); err != nil {
		if isUniqueViolation(err) {
			return ledger.Member{}, ledger.ErrMemberExists
		}
		return ledger.Member{}, err
	}
	return member, nil
}

func (s *Store) UpdateMemberRole(ctx context.Context, ledgerID, userID, role string, updatedAt time.Time) (ledger.Member, error) {
	var member ledger.Member
	row := s.pool.QueryRow(ctx, `
		UPDATE ledger_members
		SET role = $3, updated_at = $4
		WHERE ledger_id = $1 AND user_id = $2
		RETURNING ledger_id, user_id, role, created_at, updated_at
	`, ledgerID, userID, role, updatedAt)
	if err := row.Scan(&member.LedgerID, &member.UserID, &member.Role, &member.CreatedAt, &member.UpdatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ledger.Member{}, ledger.ErrNotFound
		}
		return ledger.Member{}, err
	}
	return member, nil
}

func (s *Store) RemoveMember(ctx context.Context, ledgerID, userID string) error {
	cmd, err := s.pool.Exec(ctx, `DELETE FROM ledger_members WHERE ledger_id = $1 AND user_id = $2`, ledgerID, userID)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		return ledger.ErrNotFound
	}
	return nil
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
