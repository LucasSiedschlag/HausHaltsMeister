package investments

import (
	"context"
	"errors"
	"time"

	"github.com/LucasSiedschlag/HausHaltsMeister/internal/adapters/postgres"
	journalrepo "github.com/LucasSiedschlag/HausHaltsMeister/internal/adapters/postgres/journal"
	"github.com/LucasSiedschlag/HausHaltsMeister/internal/domain/investments"
	"github.com/LucasSiedschlag/HausHaltsMeister/internal/domain/journal"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	pool        *pgxpool.Pool
	journalRepo *journalrepo.Repository
}

func NewRepository(store *postgres.Store) *Repository {
	return &Repository{
		pool:        store.Pool(),
		journalRepo: journalrepo.NewRepository(store),
	}
}

func (r *Repository) FindAccountByType(ctx context.Context, ledgerID, accountType string) (string, error) {
	var id string
	row := r.pool.QueryRow(ctx, `
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

func (r *Repository) FindCategoryByName(ctx context.Context, ledgerID, name string) (string, error) {
	var id string
	row := r.pool.QueryRow(ctx, `
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

func (r *Repository) GetCategoryDirection(ctx context.Context, ledgerID, categoryID string) (string, error) {
	var direction string
	row := r.pool.QueryRow(ctx, `
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

func (r *Repository) CreateTransaction(ctx context.Context, params journal.CreateTransactionParams) (journal.Transaction, error) {
	return r.journalRepo.CreateTransaction(ctx, params)
}

func (r *Repository) SumByCategory(ctx context.Context, ledgerID, categoryID string, from, to time.Time) (int64, error) {
	var total int64
	row := r.pool.QueryRow(ctx, `
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

func (r *Repository) GetLedgerRole(ctx context.Context, ledgerID, userID string) (string, error) {
	return postgres.GetLedgerRole(ctx, r.pool, ledgerID, userID)
}

func (r *Repository) LedgerExists(ctx context.Context, ledgerID string) (bool, error) {
	return postgres.LedgerExists(ctx, r.pool, ledgerID)
}
