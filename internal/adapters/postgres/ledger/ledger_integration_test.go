package ledger

import (
	"context"
	"testing"
	"time"

	pgstore "github.com/LucasSiedschlag/HausHaltsMeister/internal/adapters/postgres"
	"github.com/LucasSiedschlag/HausHaltsMeister/internal/adapters/postgres/postgrestest"
	"github.com/LucasSiedschlag/HausHaltsMeister/internal/domain/ledger"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

func TestCreateLedgerCreatesOwnerMember(t *testing.T) {
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

	repo := NewRepository(store)
	service := ledger.NewService(repo)

	conn, err := pgx.Connect(ctx, connStr)
	require.NoError(t, err)
	defer conn.Close(ctx)

	_, err = conn.Exec(ctx, `INSERT INTO users (id, email) VALUES ($1, $2)`, "00000000-0000-0000-0000-000000000001", "user@example.com")
	require.NoError(t, err)

	created, err := service.CreateLedger(ctx, "00000000-0000-0000-0000-000000000001", "Pessoal", "BRL")
	require.NoError(t, err)
	require.NotEmpty(t, created.ID)

	ledgers, err := service.ListLedgers(ctx, "00000000-0000-0000-0000-000000000001")
	require.NoError(t, err)
	require.Len(t, ledgers, 1)
	require.Equal(t, created.ID, ledgers[0].Ledger.ID)
	require.Equal(t, "owner", ledgers[0].Role)
	require.WithinDuration(t, time.Now().UTC(), ledgers[0].Ledger.CreatedAt, time.Second*5)
}
