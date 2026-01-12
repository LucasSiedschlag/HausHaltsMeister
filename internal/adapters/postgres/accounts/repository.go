package accounts

import (
	"context"
	"errors"
	"time"

	"github.com/LucasSiedschlag/HausHaltsMeister/internal/adapters/postgres"
	"github.com/LucasSiedschlag/HausHaltsMeister/internal/domain/accounts"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(store *postgres.Store) *Repository {
	return &Repository{pool: store.Pool()}
}

func (r *Repository) ListAccounts(ctx context.Context, ledgerID string) ([]accounts.Account, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, ledger_id, name, type, nature, is_active, created_at, updated_at
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
			&account.Nature,
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

func (r *Repository) GetAccount(ctx context.Context, ledgerID, accountID string) (accounts.Account, error) {
	var account accounts.Account
	row := r.pool.QueryRow(ctx, `
		SELECT id, ledger_id, name, type, nature, is_active, created_at, updated_at
		FROM accounts
		WHERE ledger_id = $1 AND id = $2
	`, ledgerID, accountID)
	if err := row.Scan(
		&account.ID,
		&account.LedgerID,
		&account.Name,
		&account.Type,
		&account.Nature,
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

func (r *Repository) CreateAccount(ctx context.Context, ledgerID, name, accountType, nature string, isActive bool) (accounts.Account, error) {
	var account accounts.Account
	row := r.pool.QueryRow(ctx, `
		INSERT INTO accounts (ledger_id, name, type, nature, is_active)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, ledger_id, name, type, nature, is_active, created_at, updated_at
	`, ledgerID, name, accountType, nature, isActive)
	if err := row.Scan(
		&account.ID,
		&account.LedgerID,
		&account.Name,
		&account.Type,
		&account.Nature,
		&account.IsActive,
		&account.CreatedAt,
		&account.UpdatedAt,
	); err != nil {
		if postgres.IsUniqueViolation(err) {
			return accounts.Account{}, accounts.ErrDuplicateName
		}
		return accounts.Account{}, err
	}
	return account, nil
}

func (r *Repository) UpdateAccount(ctx context.Context, ledgerID, accountID, name string, isActive bool, updatedAt time.Time) (accounts.Account, error) {
	var account accounts.Account
	row := r.pool.QueryRow(ctx, `
		UPDATE accounts
		SET name = $3, is_active = $4, updated_at = $5
		WHERE ledger_id = $1 AND id = $2
		RETURNING id, ledger_id, name, type, nature, is_active, created_at, updated_at
	`, ledgerID, accountID, name, isActive, updatedAt)
	if err := row.Scan(
		&account.ID,
		&account.LedgerID,
		&account.Name,
		&account.Type,
		&account.Nature,
		&account.IsActive,
		&account.CreatedAt,
		&account.UpdatedAt,
	); err != nil {
		if postgres.IsUniqueViolation(err) {
			return accounts.Account{}, accounts.ErrDuplicateName
		}
		if errors.Is(err, pgx.ErrNoRows) {
			return accounts.Account{}, accounts.ErrNotFound
		}
		return accounts.Account{}, err
	}
	return account, nil
}

func (r *Repository) DeactivateAccount(ctx context.Context, ledgerID, accountID string, updatedAt time.Time) error {
	cmd, err := r.pool.Exec(ctx, `
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

func (r *Repository) GetLedgerRole(ctx context.Context, ledgerID, userID string) (string, error) {
	return postgres.GetLedgerRole(ctx, r.pool, ledgerID, userID)
}

func (r *Repository) LedgerExists(ctx context.Context, ledgerID string) (bool, error) {
	return postgres.LedgerExists(ctx, r.pool, ledgerID)
}
