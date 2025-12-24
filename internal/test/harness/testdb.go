package harness

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/tern/v2/migrate"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

type TestDB struct {
	Pool      *pgxpool.Pool
	Container *postgres.PostgresContainer
	ConnStr   string
}

func SetupTestDB(t *testing.T) *TestDB {
	ctx := context.Background()

	dbName := "testdb"
	dbUser := "user"
	dbPassword := "password"

	postgresContainer, err := postgres.Run(ctx,
		"postgres:15-alpine",
		postgres.WithDatabase(dbName),
		postgres.WithUsername(dbUser),
		postgres.WithPassword(dbPassword),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(5*time.Second)),
	)
	if err != nil {
		t.Fatalf("failed to start postgres container: %s", err)
	}

	connStr, err := postgresContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("failed to get connection string: %s", err)
	}

	pool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		t.Fatalf("failed to create pool: %s", err)
	}

	// Run Migrations
	err = runMigrations(ctx, pool, connStr)
	if err != nil {
		t.Fatalf("failed to run migrations: %s", err)
	}

	tdb := &TestDB{
		Pool:      pool,
		Container: postgresContainer,
		ConnStr:   connStr,
	}

	// Register cleanup
	t.Cleanup(func() {
		pool.Close()
		if err := postgresContainer.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate container: %s", err)
		}
	})

	return tdb
}

func runMigrations(ctx context.Context, pool *pgxpool.Pool, connStr string) error {
	conn, err := pool.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

	// Locate migrations folder relative to this file
	_, filename, _, _ := runtime.Caller(0)
	root := filepath.Join(filepath.Dir(filename), "../../../") // internal/test/harness -> ../../../ -> root
	migrationsPath := filepath.Join(root, "migrations")

	migrator, err := migrate.NewMigrator(ctx, conn.Conn(), "public.schema_version")
	if err != nil {
		return fmt.Errorf("unable to create migrator: %v", err)
	}

	err = migrator.LoadMigrations(os.DirFS(migrationsPath))
	if err != nil {
		return fmt.Errorf("unable to load migrations: %v", err)
	}

	err = migrator.Migrate(ctx)
	if err != nil {
		return fmt.Errorf("unable to migrate: %v", err)
	}

	return nil
}
