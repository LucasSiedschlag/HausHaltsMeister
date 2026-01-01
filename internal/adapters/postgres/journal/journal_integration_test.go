package journal

import (
	"context"
	"testing"
	"time"

	pgstore "github.com/LucasSiedschlag/HausHaltsMeister/internal/adapters/postgres"
	"github.com/LucasSiedschlag/HausHaltsMeister/internal/adapters/postgres/postgrestest"
	"github.com/LucasSiedschlag/HausHaltsMeister/internal/domain/journal"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

func TestCreateAndListTransactions(t *testing.T) {
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

	var categoryID string
	err = conn.QueryRow(ctx, `INSERT INTO categories (ledger_id, name, direction, is_budget_base, is_budget_relevant, is_active) VALUES ($1, $2, 'out', false, true, true) RETURNING id`, ledgerID, "Mercado").Scan(&categoryID)
	require.NoError(t, err)

	repo := NewRepository(store)
	service := journal.NewService(repo)

	created, err := service.CreateTransaction(ctx, "user-1", ledgerID, journal.CreateTransactionParams{
		OccurredAt:  time.Now().UTC(),
		Description: "Compra",
		Entries: []journal.EntryInput{
			{AccountID: accountID, CategoryID: &categoryID, Kind: "normal", AmountCents: 1000},
		},
	})
	require.NoError(t, err)
	require.Len(t, created.Entries, 1)

	list, err := service.ListTransactions(ctx, "user-1", ledgerID, journal.ListTransactionsParams{Limit: 10})
	require.NoError(t, err)
	require.Len(t, list.Items, 1)
	require.Equal(t, created.ID, list.Items[0].ID)
}
