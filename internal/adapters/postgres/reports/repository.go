package reports

import (
	"context"
	"time"

	"github.com/LucasSiedschlag/HausHaltsMeister/internal/adapters/postgres"
	"github.com/LucasSiedschlag/HausHaltsMeister/internal/domain/reports"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(store *postgres.Store) *Repository {
	return &Repository{pool: store.Pool()}
}

func (r *Repository) ListAccountBalances(ctx context.Context, ledgerID string, cutoff time.Time) ([]reports.AccountBalance, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT a.id, a.name, a.type,
			COALESCE(SUM(
				CASE
					WHEN t.id IS NULL THEN 0
					WHEN c.direction = 'in' THEN e.amount_cents
					WHEN c.direction = 'out' THEN -e.amount_cents
					ELSE 0
				END
			), 0) AS net_cents
		FROM accounts a
		LEFT JOIN entries e ON e.account_id = a.id AND e.ledger_id = a.ledger_id
		LEFT JOIN transactions t ON t.id = e.transaction_id AND t.occurred_at < $2
		LEFT JOIN categories c ON c.id = e.category_id
		WHERE a.ledger_id = $1
		GROUP BY a.id, a.name, a.type
		ORDER BY a.created_at
	`, ledgerID, cutoff)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []reports.AccountBalance{}
	for rows.Next() {
		var item reports.AccountBalance
		if err := rows.Scan(&item.AccountID, &item.AccountName, &item.AccountType, &item.BalanceCents); err != nil {
			return nil, err
		}
		if item.AccountType == "credit_card" {
			item.BalanceCents = -item.BalanceCents
		}
		items = append(items, item)
	}
	return items, nil
}

func (r *Repository) ListCategorySummary(ctx context.Context, ledgerID string, from, to time.Time) ([]reports.CategorySummary, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT c.id, c.name, c.direction,
			COALESCE(SUM(
				CASE
					WHEN t.id IS NULL THEN 0
					ELSE e.amount_cents
				END
			), 0) AS total_cents
		FROM categories c
		LEFT JOIN entries e ON e.category_id = c.id AND e.ledger_id = c.ledger_id
		LEFT JOIN transactions t ON t.id = e.transaction_id AND t.occurred_at >= $2 AND t.occurred_at <= $3
		WHERE c.ledger_id = $1
		GROUP BY c.id, c.name, c.direction
		ORDER BY c.name
	`, ledgerID, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []reports.CategorySummary{}
	for rows.Next() {
		var item reports.CategorySummary
		if err := rows.Scan(&item.CategoryID, &item.Name, &item.Direction, &item.TotalCents); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}

func (r *Repository) ListCashflow(ctx context.Context, ledgerID string, from, to time.Time) ([]reports.CashflowEntry, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT date_trunc('month', t.occurred_at)::date AS month,
			COALESCE(SUM(CASE WHEN c.direction = 'in' THEN e.amount_cents ELSE 0 END), 0) AS total_in,
			COALESCE(SUM(CASE WHEN c.direction = 'out' THEN e.amount_cents ELSE 0 END), 0) AS total_out
		FROM entries e
		JOIN transactions t ON t.id = e.transaction_id
		JOIN categories c ON c.id = e.category_id
		WHERE e.ledger_id = $1 AND t.occurred_at >= $2 AND t.occurred_at <= $3
		GROUP BY month
		ORDER BY month
	`, ledgerID, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []reports.CashflowEntry{}
	for rows.Next() {
		var item reports.CashflowEntry
		if err := rows.Scan(&item.Month, &item.TotalInCents, &item.TotalOutCents); err != nil {
			return nil, err
		}
		item.NetCents = item.TotalInCents - item.TotalOutCents
		items = append(items, item)
	}
	return items, nil
}

func (r *Repository) GetLedgerRole(ctx context.Context, ledgerID, userID string) (string, error) {
	return postgres.GetLedgerRole(ctx, r.pool, ledgerID, userID)
}

func (r *Repository) LedgerExists(ctx context.Context, ledgerID string) (bool, error) {
	return postgres.LedgerExists(ctx, r.pool, ledgerID)
}
