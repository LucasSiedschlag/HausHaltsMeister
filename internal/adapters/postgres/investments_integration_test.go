package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/LucasSiedschlag/HausHaltsMeister/internal/domain/investments"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

func TestContributionCreatesTransfer(t *testing.T) {
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

	var cashAccountID string
	err = conn.QueryRow(ctx, `INSERT INTO accounts (ledger_id, name, type, is_active) VALUES ($1, $2, 'cash', true) RETURNING id`, ledgerID, "Pessoal").Scan(&cashAccountID)
	require.NoError(t, err)

	var investAccountID string
	err = conn.QueryRow(ctx, `INSERT INTO accounts (ledger_id, name, type, is_active) VALUES ($1, $2, 'investment', true) RETURNING id`, ledgerID, "Investimentos").Scan(&investAccountID)
	require.NoError(t, err)

	categories := []struct {
		Name           string
		Direction      string
		BudgetBase     bool
		BudgetRelevant bool
	}{
		{"Aportes Investimentos", "out", false, true},
		{"Entrada Investimentos (Aporte)", "in", false, false},
		{"Resgate Investimentos", "out", false, false},
		{"Entrada Resgate (Investimentos)", "in", false, false},
		{"Rendimentos", "in", false, false},
	}

	for _, item := range categories {
		_, err = conn.Exec(ctx, `INSERT INTO categories (ledger_id, name, direction, is_budget_base, is_budget_relevant, is_active) VALUES ($1, $2, $3, $4, $5, true)`, ledgerID, item.Name, item.Direction, item.BudgetBase, item.BudgetRelevant)
		require.NoError(t, err)
	}

	service := investments.NewService(store)

	_, err = service.Contribution(ctx, "user-1", ledgerID, 10000, time.Now().UTC(), nil)
	require.NoError(t, err)

	var entriesCount int
	err = conn.QueryRow(ctx, `SELECT COUNT(*) FROM entries WHERE ledger_id = $1`, ledgerID).Scan(&entriesCount)
	require.NoError(t, err)
	require.Equal(t, 2, entriesCount)
}
