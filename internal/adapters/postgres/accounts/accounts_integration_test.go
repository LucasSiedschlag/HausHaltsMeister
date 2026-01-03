package accounts

import (
	"context"
	"testing"

	pgstore "github.com/LucasSiedschlag/HausHaltsMeister/internal/adapters/postgres"
	"github.com/LucasSiedschlag/HausHaltsMeister/internal/adapters/postgres/postgrestest"
	"github.com/LucasSiedschlag/HausHaltsMeister/internal/domain/accounts"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

func TestCreateAccountUniquePerLedger(t *testing.T) {
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

	repo := NewRepository(store)
	service := accounts.NewService(repo)

	_, err = service.CreateAccount(ctx, "00000000-0000-0000-0000-000000000001", ledgerID, "Conta", "cash", true)
	require.NoError(t, err)

	_, err = service.CreateAccount(ctx, "00000000-0000-0000-0000-000000000001", ledgerID, "Conta", "cash", true)
	require.Error(t, err)
	require.Equal(t, accounts.ErrDuplicateName, err)
}

func TestGetAccountCrossLedgerNotFound(t *testing.T) {
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

	var ledgerA string
	err = conn.QueryRow(ctx, `INSERT INTO ledgers (owner_user_id, name, currency_code) VALUES ($1, $2, $3) RETURNING id`, "00000000-0000-0000-0000-000000000001", "Ledger A", "BRL").Scan(&ledgerA)
	require.NoError(t, err)

	var ledgerB string
	err = conn.QueryRow(ctx, `INSERT INTO ledgers (owner_user_id, name, currency_code) VALUES ($1, $2, $3) RETURNING id`, "00000000-0000-0000-0000-000000000001", "Ledger B", "BRL").Scan(&ledgerB)
	require.NoError(t, err)

	_, err = conn.Exec(ctx, `INSERT INTO ledger_members (ledger_id, user_id, role) VALUES ($1, $2, 'owner')`, ledgerB, "00000000-0000-0000-0000-000000000001")
	require.NoError(t, err)

	var accountID string
	err = conn.QueryRow(ctx, `INSERT INTO accounts (ledger_id, name, type, is_active) VALUES ($1, $2, 'cash', true) RETURNING id`, ledgerA, "Conta A").Scan(&accountID)
	require.NoError(t, err)

	repo := NewRepository(store)
	service := accounts.NewService(repo)

	_, err = service.GetAccount(ctx, "00000000-0000-0000-0000-000000000001", ledgerB, accountID)
	require.Error(t, err)
	require.Equal(t, accounts.ErrAccountNotFound, err)
}
