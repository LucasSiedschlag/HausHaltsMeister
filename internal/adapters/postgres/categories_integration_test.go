package postgres

import (
	"context"
	"testing"

	"github.com/LucasSiedschlag/HausHaltsMeister/internal/domain/categories"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

func TestCreateCategoryUniquePerLedger(t *testing.T) {
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

	service := categories.NewService(store)

	_, err = service.CreateCategory(ctx, "user-1", ledgerID, categories.CreateCategoryParams{
		Name:             "Mercado",
		Direction:        "out",
		IsBudgetBase:     false,
		IsBudgetRelevant: true,
		IsActive:         true,
	})
	require.NoError(t, err)

	_, err = service.CreateCategory(ctx, "user-1", ledgerID, categories.CreateCategoryParams{
		Name:             "Mercado",
		Direction:        "out",
		IsBudgetBase:     false,
		IsBudgetRelevant: true,
		IsActive:         true,
	})
	require.Error(t, err)
	require.Equal(t, categories.ErrDuplicateName, err)
}
