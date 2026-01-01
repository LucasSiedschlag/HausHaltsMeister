package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/LucasSiedschlag/HausHaltsMeister/internal/domain/accounts"
	"github.com/jackc/pgx/v5"
)

func (s *Store) ListAccounts(ctx context.Context, ledgerID string) ([]accounts.Account, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, ledger_id, name, type, is_active, created_at, updated_at
		FROM accounts
		WHERE ledger_id = $1
		ORDER BY created_at
	`, ledgerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []accounts.Account
	for rows.Next() {
		var account accounts.Account
		if err := rows.Scan(
			&account.ID,
			&account.LedgerID,
			&account.Name,
			&account.Type,
			&account.IsActive,
			&account.CreatedAt,
			&account.UpdatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, account)
	}
	return items, nil
}

func (s *Store) GetAccount(ctx context.Context, ledgerID, accountID string) (accounts.Account, error) {
	var account accounts.Account
	row := s.pool.QueryRow(ctx, `
		SELECT id, ledger_id, name, type, is_active, created_at, updated_at
		FROM accounts
		WHERE ledger_id = $1 AND id = $2
	`, ledgerID, accountID)
	if err := row.Scan(
		&account.ID,
		&account.LedgerID,
		&account.Name,
		&account.Type,
		&account.IsActive,
		&account.CreatedAt,
		&account.UpdatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return accounts.Account{}, accounts.ErrNotFound
		}
		return accounts.Account{}, err
	}
	return account, nil
}

func (s *Store) CreateAccount(ctx context.Context, ledgerID, name, accountType string, isActive bool) (accounts.Account, error) {
	var account accounts.Account
	row := s.pool.QueryRow(ctx, `
		INSERT INTO accounts (ledger_id, name, type, is_active)
		VALUES ($1, $2, $3, $4)
		RETURNING id, ledger_id, name, type, is_active, created_at, updated_at
	`, ledgerID, name, accountType, isActive)
	if err := row.Scan(
		&account.ID,
		&account.LedgerID,
		&account.Name,
		&account.Type,
		&account.IsActive,
		&account.CreatedAt,
		&account.UpdatedAt,
	); err != nil {
		if isUniqueViolation(err) {
			return accounts.Account{}, accounts.ErrDuplicateName
		}
		return accounts.Account{}, err
	}
	return account, nil
}

func (s *Store) UpdateAccount(ctx context.Context, ledgerID, accountID, name string, isActive bool, updatedAt time.Time) (accounts.Account, error) {
	var account accounts.Account
	row := s.pool.QueryRow(ctx, `
		UPDATE accounts
		SET name = $3, is_active = $4, updated_at = $5
		WHERE ledger_id = $1 AND id = $2
		RETURNING id, ledger_id, name, type, is_active, created_at, updated_at
	`, ledgerID, accountID, name, isActive, updatedAt)
	if err := row.Scan(
		&account.ID,
		&account.LedgerID,
		&account.Name,
		&account.Type,
		&account.IsActive,
		&account.CreatedAt,
		&account.UpdatedAt,
	); err != nil {
		if isUniqueViolation(err) {
			return accounts.Account{}, accounts.ErrDuplicateName
		}
		if errors.Is(err, pgx.ErrNoRows) {
			return accounts.Account{}, accounts.ErrNotFound
		}
		return accounts.Account{}, err
	}
	return account, nil
}

func (s *Store) DeactivateAccount(ctx context.Context, ledgerID, accountID string, updatedAt time.Time) error {
	cmd, err := s.pool.Exec(ctx, `
		UPDATE accounts
		SET is_active = false, updated_at = $3
		WHERE ledger_id = $1 AND id = $2
	`, ledgerID, accountID, updatedAt)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		return accounts.ErrNotFound
	}
	return nil
}
