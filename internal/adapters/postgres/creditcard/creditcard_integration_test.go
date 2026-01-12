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

	_, err = conn.Exec(ctx, `INSERT INTO users (id, email) VALUES ($1, $2)`, "00000000-0000-0000-0000-000000000001", "user@example.com")
	require.NoError(t, err)

	var ledgerID string
	err = conn.QueryRow(ctx, `INSERT INTO ledgers (owner_user_id, name, currency_code) VALUES ($1, $2, $3) RETURNING id`, "00000000-0000-0000-0000-000000000001", "Pessoal", "BRL").Scan(&ledgerID)
	require.NoError(t, err)

	_, err = conn.Exec(ctx, `INSERT INTO ledger_members (ledger_id, user_id, role) VALUES ($1, $2, 'owner')`, ledgerID, "00000000-0000-0000-0000-000000000001")
	require.NoError(t, err)

	var parentAccountID string
	err = conn.QueryRow(ctx, `INSERT INTO accounts (ledger_id, name, type, nature, is_active) VALUES ($1, $2, 'current', 'asset', true) RETURNING id`, ledgerID, "Conta Principal").Scan(&parentAccountID)
	require.NoError(t, err)

	var liabilityAccountID string
	err = conn.QueryRow(ctx, `INSERT INTO accounts (ledger_id, name, type, nature, is_active) VALUES ($1, $2, 'current', 'liability', true) RETURNING id`, ledgerID, "Cartao Passivo").Scan(&liabilityAccountID)
	require.NoError(t, err)

	var cardID string
	err = conn.QueryRow(ctx, `
		INSERT INTO credit_cards (ledger_id, parent_account_id, liability_account_id, label, brand, closing_day, due_day, active)
		VALUES ($1, $2, $3, $4, $5, $6, $7, true)
		RETURNING id
	`, ledgerID, parentAccountID, liabilityAccountID, "Cartao", "visa", 10, 20).Scan(&cardID)
	require.NoError(t, err)

	var categoryID string
	err = conn.QueryRow(ctx, `INSERT INTO categories (ledger_id, name, direction, is_budget_base, is_budget_relevant, is_active) VALUES ($1, $2, 'out', false, true, true) RETURNING id`, ledgerID, "Mercado").Scan(&categoryID)
	require.NoError(t, err)

	purchaseAt := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)
	firstDue := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)

	var planID string
	err = conn.QueryRow(ctx, `
		INSERT INTO installment_plans (ledger_id, credit_card_id, purchase_occurred_at, merchant, description, category_id, total_amount_cents, installments_count, installment_amount_cents, first_due_month, status, created_by_user_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, 'active', $11)
		RETURNING id
	`, ledgerID, cardID, purchaseAt, "Loja", "Compra", categoryID, 12000, 1, 12000, firstDue, "00000000-0000-0000-0000-000000000001").Scan(&planID)
	require.NoError(t, err)

	_, err = conn.Exec(ctx, `
		INSERT INTO installments (ledger_id, plan_id, installment_no, due_month, amount_cents, status)
		VALUES ($1, $2, 1, $3, $4, 'scheduled')
	`, ledgerID, planID, firstDue, 12000)
	require.NoError(t, err)

	repo := NewRepository(store)
	service := creditcard.NewService(repo)

	result, err := service.PostMonth(ctx, "00000000-0000-0000-0000-000000000001", cardID, firstDue)
	require.NoError(t, err)
	require.Equal(t, 1, result.PostedCount)

	result, err = service.PostMonth(ctx, "00000000-0000-0000-0000-000000000001", cardID, firstDue)
	require.NoError(t, err)
	require.Equal(t, 0, result.PostedCount)

	var postedCount int
	err = conn.QueryRow(ctx, `SELECT COUNT(*) FROM installments WHERE plan_id = $1 AND status = 'posted' AND posted_transaction_id IS NOT NULL`, planID).Scan(&postedCount)
	require.NoError(t, err)
	require.Equal(t, 1, postedCount)

	var entryCount int
	err = conn.QueryRow(ctx, `SELECT COUNT(*) FROM entries WHERE ledger_id = $1 AND account_id = $2`, ledgerID, liabilityAccountID).Scan(&entryCount)
	require.NoError(t, err)
	require.Equal(t, 1, entryCount)
}

func TestGetPlanAccessDenied(t *testing.T) {
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
	_, err = conn.Exec(ctx, `INSERT INTO users (id, email) VALUES ($1, $2)`, "00000000-0000-0000-0000-000000000002", "owner@example.com")
	require.NoError(t, err)

	var ledgerA string
	err = conn.QueryRow(ctx, `INSERT INTO ledgers (owner_user_id, name, currency_code) VALUES ($1, $2, $3) RETURNING id`, "00000000-0000-0000-0000-000000000002", "Ledger A", "BRL").Scan(&ledgerA)
	require.NoError(t, err)

	var ledgerB string
	err = conn.QueryRow(ctx, `INSERT INTO ledgers (owner_user_id, name, currency_code) VALUES ($1, $2, $3) RETURNING id`, "00000000-0000-0000-0000-000000000001", "Ledger B", "BRL").Scan(&ledgerB)
	require.NoError(t, err)

	_, err = conn.Exec(ctx, `INSERT INTO ledger_members (ledger_id, user_id, role) VALUES ($1, $2, 'owner')`, ledgerB, "00000000-0000-0000-0000-000000000001")
	require.NoError(t, err)

	var parentAccountID string
	err = conn.QueryRow(ctx, `INSERT INTO accounts (ledger_id, name, type, nature, is_active) VALUES ($1, $2, 'current', 'asset', true) RETURNING id`, ledgerA, "Conta Principal").Scan(&parentAccountID)
	require.NoError(t, err)

	var liabilityAccountID string
	err = conn.QueryRow(ctx, `INSERT INTO accounts (ledger_id, name, type, nature, is_active) VALUES ($1, $2, 'current', 'liability', true) RETURNING id`, ledgerA, "Cartao Passivo").Scan(&liabilityAccountID)
	require.NoError(t, err)

	var cardID string
	err = conn.QueryRow(ctx, `
		INSERT INTO credit_cards (ledger_id, parent_account_id, liability_account_id, label, brand, closing_day, due_day, active)
		VALUES ($1, $2, $3, $4, $5, $6, $7, true)
		RETURNING id
	`, ledgerA, parentAccountID, liabilityAccountID, "Cartao", "visa", 10, 20).Scan(&cardID)
	require.NoError(t, err)

	var categoryID string
	err = conn.QueryRow(ctx, `INSERT INTO categories (ledger_id, name, direction, is_budget_base, is_budget_relevant, is_active) VALUES ($1, $2, 'out', false, true, true) RETURNING id`, ledgerA, "Mercado").Scan(&categoryID)
	require.NoError(t, err)

	purchaseAt := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)
	firstDue := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)

	var planID string
	err = conn.QueryRow(ctx, `
		INSERT INTO installment_plans (ledger_id, credit_card_id, purchase_occurred_at, merchant, description, category_id, total_amount_cents, installments_count, installment_amount_cents, first_due_month, status, created_by_user_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, 'active', $11)
		RETURNING id
	`, ledgerA, cardID, purchaseAt, "Loja", "Compra", categoryID, 12000, 1, 12000, firstDue, "00000000-0000-0000-0000-000000000002").Scan(&planID)
	require.NoError(t, err)

	repo := NewRepository(store)
	service := creditcard.NewService(repo)

	_, err = service.GetPlan(ctx, "00000000-0000-0000-0000-000000000001", cardID, planID)
	require.Error(t, err)
	require.Equal(t, creditcard.ErrAccessDenied, err)
}
