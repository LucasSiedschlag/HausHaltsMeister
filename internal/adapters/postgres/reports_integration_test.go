package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

func TestListAccountBalances(t *testing.T) {
	ctx := context.Background()
	skipIfDockerUnavailable(t)

	pgContainer, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("hhm"),
		postgres.WithUsername("hhm"),
		postgres.WithPassword("hhm"),
	)
	require.NoError(t, err)
	defer func() { _ = pgContainer.Terminate(ctx) }()

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	require.NoError(t, waitForPostgres(ctx, connStr))
	require.NoError(t, applyMigrations(ctx, connStr))

	store, err := NewStore(ctx, connStr)
	require.NoError(t, err)
	defer store.Close()

	conn, err := store.pool.Acquire(ctx)
	require.NoError(t, err)
	defer conn.Release()

	_, err = conn.Exec(ctx, `INSERT INTO users (id, email) VALUES ($1, $2)`, "user-1", "user@example.com")
	require.NoError(t, err)

	var ledgerID string
	err = conn.QueryRow(ctx, `INSERT INTO ledgers (owner_user_id, name, currency_code) VALUES ($1, $2, $3) RETURNING id`, "user-1", "Pessoal", "BRL").Scan(&ledgerID)
	require.NoError(t, err)

	_, err = conn.Exec(ctx, `INSERT INTO ledger_members (ledger_id, user_id, role) VALUES ($1, $2, 'owner')`, ledgerID, "user-1")
	require.NoError(t, err)

	var cashAccountID string
	err = conn.QueryRow(ctx, `INSERT INTO accounts (ledger_id, name, type, is_active) VALUES ($1, $2, 'cash', true) RETURNING id`, ledgerID, "Pessoal").Scan(&cashAccountID)
	require.NoError(t, err)

	var cardAccountID string
	err = conn.QueryRow(ctx, `INSERT INTO accounts (ledger_id, name, type, is_active) VALUES ($1, $2, 'credit_card', true) RETURNING id`, ledgerID, "Cartao").Scan(&cardAccountID)
	require.NoError(t, err)

	var catInID string
	err = conn.QueryRow(ctx, `INSERT INTO categories (ledger_id, name, direction, is_budget_base, is_budget_relevant, is_active) VALUES ($1, $2, 'in', true, false, true) RETURNING id`, ledgerID, "Salario").Scan(&catInID)
	require.NoError(t, err)

	var catOutID string
	err = conn.QueryRow(ctx, `INSERT INTO categories (ledger_id, name, direction, is_budget_base, is_budget_relevant, is_active) VALUES ($1, $2, 'out', false, true, true) RETURNING id`, ledgerID, "Mercado").Scan(&catOutID)
	require.NoError(t, err)

	occurredAt := time.Date(2026, 2, 10, 0, 0, 0, 0, time.UTC)
	var txID string
	err = conn.QueryRow(ctx, `INSERT INTO transactions (ledger_id, occurred_at, description, created_by_user_id) VALUES ($1, $2, $3, $4) RETURNING id`, ledgerID, occurredAt, "Salario", "user-1").Scan(&txID)
	require.NoError(t, err)
	_, err = conn.Exec(ctx, `INSERT INTO entries (transaction_id, ledger_id, account_id, category_id, kind, amount_cents) VALUES ($1, $2, $3, $4, 'normal', $5)`, txID, ledgerID, cashAccountID, catInID, 10000)
	require.NoError(t, err)

	occurredAt = time.Date(2026, 2, 15, 0, 0, 0, 0, time.UTC)
	err = conn.QueryRow(ctx, `INSERT INTO transactions (ledger_id, occurred_at, description, created_by_user_id) VALUES ($1, $2, $3, $4) RETURNING id`, ledgerID, occurredAt, "Mercado", "user-1").Scan(&txID)
	require.NoError(t, err)
	_, err = conn.Exec(ctx, `INSERT INTO entries (transaction_id, ledger_id, account_id, category_id, kind, amount_cents) VALUES ($1, $2, $3, $4, 'normal', $5)`, txID, ledgerID, cashAccountID, catOutID, 3000)
	require.NoError(t, err)

	occurredAt = time.Date(2026, 2, 20, 0, 0, 0, 0, time.UTC)
	err = conn.QueryRow(ctx, `INSERT INTO transactions (ledger_id, occurred_at, description, created_by_user_id) VALUES ($1, $2, $3, $4) RETURNING id`, ledgerID, occurredAt, "Compra Cartao", "user-1").Scan(&txID)
	require.NoError(t, err)
	_, err = conn.Exec(ctx, `INSERT INTO entries (transaction_id, ledger_id, account_id, category_id, kind, amount_cents) VALUES ($1, $2, $3, $4, 'normal', $5)`, txID, ledgerID, cardAccountID, catOutID, 2000)
	require.NoError(t, err)

	cutoff := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
	items, err := store.ListAccountBalances(ctx, ledgerID, cutoff)
	require.NoError(t, err)

	var cashBalance int64
	var cardBalance int64
	for _, item := range items {
		if item.AccountID == cashAccountID {
			cashBalance = item.BalanceCents
		}
		if item.AccountID == cardAccountID {
			cardBalance = item.BalanceCents
		}
	}

	require.Equal(t, int64(7000), cashBalance)
	require.Equal(t, int64(2000), cardBalance)
}
