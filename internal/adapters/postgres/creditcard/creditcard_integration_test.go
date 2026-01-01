package creditcard

import (
	"context"
	"testing"
	"time"

	pgstore "github.com/LucasSiedschlag/HausHaltsMeister/internal/adapters/postgres"
	"github.com/LucasSiedschlag/HausHaltsMeister/internal/adapters/postgres/postgrestest"
	"github.com/LucasSiedschlag/HausHaltsMeister/internal/domain/creditcard"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

func TestPostMonthMarksInstallmentPosted(t *testing.T) {
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

	_, err = conn.Exec(ctx, `INSERT INTO users (id, email) VALUES ($1, $2)`, "user-1", "user@example.com")
	require.NoError(t, err)

	var ledgerID string
	err = conn.QueryRow(ctx, `INSERT INTO ledgers (owner_user_id, name, currency_code) VALUES ($1, $2, $3) RETURNING id`, "user-1", "Pessoal", "BRL").Scan(&ledgerID)
	require.NoError(t, err)

	_, err = conn.Exec(ctx, `INSERT INTO ledger_members (ledger_id, user_id, role) VALUES ($1, $2, 'owner')`, ledgerID, "user-1")
	require.NoError(t, err)

	var cardAccountID string
	err = conn.QueryRow(ctx, `INSERT INTO accounts (ledger_id, name, type, is_active) VALUES ($1, $2, 'credit_card', true) RETURNING id`, ledgerID, "Cartao").Scan(&cardAccountID)
	require.NoError(t, err)

	var categoryID string
	err = conn.QueryRow(ctx, `INSERT INTO categories (ledger_id, name, direction, is_budget_base, is_budget_relevant, is_active) VALUES ($1, $2, 'out', false, true, true) RETURNING id`, ledgerID, "Mercado").Scan(&categoryID)
	require.NoError(t, err)

	purchaseAt := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)
	firstDue := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)

	var planID string
	err = conn.QueryRow(ctx, `
		INSERT INTO installment_plans (ledger_id, card_account_id, purchase_occurred_at, merchant, description, category_id, total_amount_cents, installments_count, installment_amount_cents, first_due_month, status, created_by_user_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, 'active', $11)
		RETURNING id
	`, ledgerID, cardAccountID, purchaseAt, "Loja", "Compra", categoryID, 12000, 1, 12000, firstDue, "user-1").Scan(&planID)
	require.NoError(t, err)

	_, err = conn.Exec(ctx, `
		INSERT INTO installments (ledger_id, plan_id, installment_no, due_month, amount_cents, status)
		VALUES ($1, $2, 1, $3, $4, 'scheduled')
	`, ledgerID, planID, firstDue, 12000)
	require.NoError(t, err)

	repo := NewRepository(store)
	service := creditcard.NewService(repo)

	result, err := service.PostMonth(ctx, "user-1", ledgerID, cardAccountID, firstDue)
	require.NoError(t, err)
	require.Equal(t, 1, result.PostedCount)

	result, err = service.PostMonth(ctx, "user-1", ledgerID, cardAccountID, firstDue)
	require.NoError(t, err)
	require.Equal(t, 0, result.PostedCount)

	var postedCount int
	err = conn.QueryRow(ctx, `SELECT COUNT(*) FROM installments WHERE plan_id = $1 AND status = 'posted' AND posted_transaction_id IS NOT NULL`, planID).Scan(&postedCount)
	require.NoError(t, err)
	require.Equal(t, 1, postedCount)

	var entryCount int
	err = conn.QueryRow(ctx, `SELECT COUNT(*) FROM entries WHERE ledger_id = $1 AND account_id = $2`, ledgerID, cardAccountID).Scan(&entryCount)
	require.NoError(t, err)
	require.Equal(t, 1, entryCount)
}
