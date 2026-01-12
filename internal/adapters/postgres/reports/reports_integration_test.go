package reports

import (
	"context"
	"testing"
	"time"

	pgstore "github.com/LucasSiedschlag/HausHaltsMeister/internal/adapters/postgres"
	"github.com/LucasSiedschlag/HausHaltsMeister/internal/adapters/postgres/postgrestest"
	"github.com/LucasSiedschlag/HausHaltsMeister/internal/domain/reports"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

func TestListAccountBalances(t *testing.T) {
	ctx := context.Background()
	postgrestest.SkipIfDockerUnavailable(t)

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

	require.NoError(t, postgrestest.WaitForPostgres(ctx, connStr))
	require.NoError(t, postgrestest.ApplyMigrations(ctx, connStr))

	store, err := pgstore.NewStore(ctx, connStr)
	require.NoError(t, err)
	defer store.Close()

	conn, err := pgx.Connect(ctx, connStr)
	require.NoError(t, err)
	defer conn.Close(ctx)

	_, err = conn.Exec(ctx, `INSERT INTO users (id, email) VALUES ($1, $2)`, "00000000-0000-0000-0000-000000000001", "user@example.com")
	require.NoError(t, err)

	var ledgerID string
	err = conn.QueryRow(ctx, `INSERT INTO ledgers (owner_user_id, name, currency_code) VALUES ($1, $2, $3) RETURNING id`, "00000000-0000-0000-0000-000000000001", "Pessoal", "BRL").Scan(&ledgerID)
	require.NoError(t, err)

	_, err = conn.Exec(ctx, `INSERT INTO ledger_members (ledger_id, user_id, role) VALUES ($1, $2, 'owner')`, ledgerID, "00000000-0000-0000-0000-000000000001")
	require.NoError(t, err)

	var cashAccountID string
	err = conn.QueryRow(ctx, `INSERT INTO accounts (ledger_id, name, type, nature, is_active) VALUES ($1, $2, 'wallet', 'asset', true) RETURNING id`, ledgerID, "Pessoal").Scan(&cashAccountID)
	require.NoError(t, err)

	var cardAccountID string
	err = conn.QueryRow(ctx, `INSERT INTO accounts (ledger_id, name, type, nature, is_active) VALUES ($1, $2, 'current', 'liability', true) RETURNING id`, ledgerID, "Cartao").Scan(&cardAccountID)
	require.NoError(t, err)

	var catInID string
	err = conn.QueryRow(ctx, `INSERT INTO categories (ledger_id, name, direction, is_budget_base, is_budget_relevant, is_active) VALUES ($1, $2, 'in', true, false, true) RETURNING id`, ledgerID, "Salario").Scan(&catInID)
	require.NoError(t, err)

	var catOutID string
	err = conn.QueryRow(ctx, `INSERT INTO categories (ledger_id, name, direction, is_budget_base, is_budget_relevant, is_active) VALUES ($1, $2, 'out', false, true, true) RETURNING id`, ledgerID, "Mercado").Scan(&catOutID)
	require.NoError(t, err)

	occurredAt := time.Date(2026, 2, 10, 0, 0, 0, 0, time.UTC)
	var txID string
	err = conn.QueryRow(ctx, `INSERT INTO transactions (ledger_id, occurred_at, description, created_by_user_id) VALUES ($1, $2, $3, $4) RETURNING id`, ledgerID, occurredAt, "Salario", "00000000-0000-0000-0000-000000000001").Scan(&txID)
	require.NoError(t, err)
	_, err = conn.Exec(ctx, `INSERT INTO entries (transaction_id, ledger_id, account_id, category_id, kind, amount_cents) VALUES ($1, $2, $3, $4, 'normal', $5)`, txID, ledgerID, cashAccountID, catInID, 10000)
	require.NoError(t, err)

	occurredAt = time.Date(2026, 2, 15, 0, 0, 0, 0, time.UTC)
	err = conn.QueryRow(ctx, `INSERT INTO transactions (ledger_id, occurred_at, description, created_by_user_id) VALUES ($1, $2, $3, $4) RETURNING id`, ledgerID, occurredAt, "Mercado", "00000000-0000-0000-0000-000000000001").Scan(&txID)
	require.NoError(t, err)
	_, err = conn.Exec(ctx, `INSERT INTO entries (transaction_id, ledger_id, account_id, category_id, kind, amount_cents) VALUES ($1, $2, $3, $4, 'normal', $5)`, txID, ledgerID, cashAccountID, catOutID, 3000)
	require.NoError(t, err)

	occurredAt = time.Date(2026, 2, 20, 0, 0, 0, 0, time.UTC)
	err = conn.QueryRow(ctx, `INSERT INTO transactions (ledger_id, occurred_at, description, created_by_user_id) VALUES ($1, $2, $3, $4) RETURNING id`, ledgerID, occurredAt, "Compra Cartao", "00000000-0000-0000-0000-000000000001").Scan(&txID)
	require.NoError(t, err)
	_, err = conn.Exec(ctx, `INSERT INTO entries (transaction_id, ledger_id, account_id, category_id, kind, amount_cents) VALUES ($1, $2, $3, $4, 'normal', $5)`, txID, ledgerID, cardAccountID, catOutID, 2000)
	require.NoError(t, err)

	repo := NewRepository(store)
	service := reports.NewService(repo)

	result, err := service.Balances(ctx, "00000000-0000-0000-0000-000000000001", ledgerID, time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC))
	require.NoError(t, err)

	var cashBalance int64
	var cardBalance int64
	for _, item := range result.Items {
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
