package budget

import (
	"context"
	"testing"
	"time"

	pgstore "github.com/LucasSiedschlag/HausHaltsMeister/internal/adapters/postgres"
	"github.com/LucasSiedschlag/HausHaltsMeister/internal/adapters/postgres/postgrestest"
	"github.com/LucasSiedschlag/HausHaltsMeister/internal/domain/budget"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

func TestMonthlyBudgetSummary(t *testing.T) {
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

	var accountID string
	err = conn.QueryRow(ctx, `INSERT INTO accounts (ledger_id, name, type, is_active) VALUES ($1, $2, $3, true) RETURNING id`, ledgerID, "Pessoal", "cash").Scan(&accountID)
	require.NoError(t, err)

	var incomeCategoryID string
	err = conn.QueryRow(ctx, `INSERT INTO categories (ledger_id, name, direction, is_budget_base, is_budget_relevant, is_active) VALUES ($1, $2, 'in', true, false, true) RETURNING id`, ledgerID, "Salario").Scan(&incomeCategoryID)
	require.NoError(t, err)

	var outCategoryID string
	err = conn.QueryRow(ctx, `INSERT INTO categories (ledger_id, name, direction, is_budget_base, is_budget_relevant, is_active) VALUES ($1, $2, 'out', false, true, true) RETURNING id`, ledgerID, "Mercado").Scan(&outCategoryID)
	require.NoError(t, err)

	month := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	var incomeTxID string
	err = conn.QueryRow(ctx, `INSERT INTO transactions (ledger_id, occurred_at, description, created_by_user_id) VALUES ($1, $2, $3, $4) RETURNING id`, ledgerID, month.AddDate(0, 0, 1), "Salario", "user-1").Scan(&incomeTxID)
	require.NoError(t, err)

	_, err = conn.Exec(ctx, `INSERT INTO entries (ledger_id, transaction_id, account_id, category_id, kind, amount_cents) VALUES ($1, $2, $3, $4, 'normal', $5)`, ledgerID, incomeTxID, accountID, incomeCategoryID, 500000)
	require.NoError(t, err)

	var spendTxID string
	err = conn.QueryRow(ctx, `INSERT INTO transactions (ledger_id, occurred_at, description, created_by_user_id) VALUES ($1, $2, $3, $4) RETURNING id`, ledgerID, month.AddDate(0, 0, 2), "Compra", "user-1").Scan(&spendTxID)
	require.NoError(t, err)

	_, err = conn.Exec(ctx, `INSERT INTO entries (ledger_id, transaction_id, account_id, category_id, kind, amount_cents) VALUES ($1, $2, $3, $4, 'normal', $5)`, ledgerID, spendTxID, accountID, outCategoryID, 100000)
	require.NoError(t, err)

	repo := NewRepository(store)
	service := budget.NewService(repo)

	_, err = service.CreateVersion(ctx, "user-1", ledgerID, month, []budget.LineInput{{CategoryID: outCategoryID, Percent: 10, IncludeChildren: false}})
	require.NoError(t, err)

	summary, err := service.MonthlySummary(ctx, "user-1", ledgerID, month)
	require.NoError(t, err)
	require.Equal(t, int64(500000), summary.IncomeBaseCents)
	require.Len(t, summary.Lines, 1)
	require.Equal(t, int64(50000), summary.Lines[0].BudgetLimitCents)
	require.Equal(t, int64(100000), summary.Lines[0].SpentActualCents)
}

func TestPeriodBudgetSummaryAggregates(t *testing.T) {
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

	var accountID string
	err = conn.QueryRow(ctx, `INSERT INTO accounts (ledger_id, name, type, is_active) VALUES ($1, $2, $3, true) RETURNING id`, ledgerID, "Pessoal", "cash").Scan(&accountID)
	require.NoError(t, err)

	var incomeCategoryID string
	err = conn.QueryRow(ctx, `INSERT INTO categories (ledger_id, name, direction, is_budget_base, is_budget_relevant, is_active) VALUES ($1, $2, 'in', true, false, true) RETURNING id`, ledgerID, "Salario").Scan(&incomeCategoryID)
	require.NoError(t, err)

	var outCategoryID string
	err = conn.QueryRow(ctx, `INSERT INTO categories (ledger_id, name, direction, is_budget_base, is_budget_relevant, is_active) VALUES ($1, $2, 'out', false, true, true) RETURNING id`, ledgerID, "Mercado").Scan(&outCategoryID)
	require.NoError(t, err)

	month1 := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	month2 := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)

	var incomeTxID string
	err = conn.QueryRow(ctx, `INSERT INTO transactions (ledger_id, occurred_at, description, created_by_user_id) VALUES ($1, $2, $3, $4) RETURNING id`, ledgerID, month1.AddDate(0, 0, 1), "Salario", "user-1").Scan(&incomeTxID)
	require.NoError(t, err)
	_, err = conn.Exec(ctx, `INSERT INTO entries (ledger_id, transaction_id, account_id, category_id, kind, amount_cents) VALUES ($1, $2, $3, $4, 'normal', $5)`, ledgerID, incomeTxID, accountID, incomeCategoryID, 500000)
	require.NoError(t, err)

	err = conn.QueryRow(ctx, `INSERT INTO transactions (ledger_id, occurred_at, description, created_by_user_id) VALUES ($1, $2, $3, $4) RETURNING id`, ledgerID, month2.AddDate(0, 0, 1), "Salario", "user-1").Scan(&incomeTxID)
	require.NoError(t, err)
	_, err = conn.Exec(ctx, `INSERT INTO entries (ledger_id, transaction_id, account_id, category_id, kind, amount_cents) VALUES ($1, $2, $3, $4, 'normal', $5)`, ledgerID, incomeTxID, accountID, incomeCategoryID, 300000)
	require.NoError(t, err)

	var spendTxID string
	err = conn.QueryRow(ctx, `INSERT INTO transactions (ledger_id, occurred_at, description, created_by_user_id) VALUES ($1, $2, $3, $4) RETURNING id`, ledgerID, month1.AddDate(0, 0, 2), "Mercado", "user-1").Scan(&spendTxID)
	require.NoError(t, err)
	_, err = conn.Exec(ctx, `INSERT INTO entries (ledger_id, transaction_id, account_id, category_id, kind, amount_cents) VALUES ($1, $2, $3, $4, 'normal', $5)`, ledgerID, spendTxID, accountID, outCategoryID, 50000)
	require.NoError(t, err)

	err = conn.QueryRow(ctx, `INSERT INTO transactions (ledger_id, occurred_at, description, created_by_user_id) VALUES ($1, $2, $3, $4) RETURNING id`, ledgerID, month2.AddDate(0, 0, 3), "Mercado", "user-1").Scan(&spendTxID)
	require.NoError(t, err)
	_, err = conn.Exec(ctx, `INSERT INTO entries (ledger_id, transaction_id, account_id, category_id, kind, amount_cents) VALUES ($1, $2, $3, $4, 'normal', $5)`, ledgerID, spendTxID, accountID, outCategoryID, 30000)
	require.NoError(t, err)

	repo := NewRepository(store)
	service := budget.NewService(repo)

	_, err = service.CreateVersion(ctx, "user-1", ledgerID, month1, []budget.LineInput{{CategoryID: outCategoryID, Percent: 10, IncludeChildren: false}})
	require.NoError(t, err)

	summary, err := service.PeriodSummary(ctx, "user-1", ledgerID, month1, month2)
	require.NoError(t, err)
	require.Len(t, summary.Months, 2)
	require.Equal(t, int64(80000), summary.TotalSpentCents)
	require.Equal(t, int64(80000), summary.TotalBudgetedCents)
}
