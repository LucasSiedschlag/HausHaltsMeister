package postgres

import (
	"context"
	"crypto/sha256"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/LucasSiedschlag/HausHaltsMeister/internal/domain/auth"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

func TestRefreshRotationPersists(t *testing.T) {
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

	service := auth.NewService(store, auth.ServiceConfig{
		JWTSecret:  "test",
		AccessTTL:  15 * time.Minute,
		RefreshTTL: 30 * 24 * time.Hour,
		Providers:  map[string]auth.OAuthProvider{},
	})

	signupResult, err := service.SignUp(ctx, "user@example.com", "password123", "User", "", "")
	require.NoError(t, err)
	oldRefresh := signupResult.Tokens.RefreshToken

	_, err = service.Refresh(ctx, oldRefresh, "", "")
	require.NoError(t, err)

	conn, err := pgx.Connect(ctx, connStr)
	require.NoError(t, err)
	defer conn.Close(ctx)

	var revokedAt *time.Time
	err = conn.QueryRow(ctx, "SELECT revoked_at FROM auth_sessions WHERE refresh_token_hash = $1", hashToken(oldRefresh)).Scan(&revokedAt)
	require.NoError(t, err)
	require.NotNil(t, revokedAt)
}

func applyMigrations(ctx context.Context, connStr string) error {
	conn, err := pgx.Connect(ctx, connStr)
	if err != nil {
		return err
	}
	defer conn.Close(ctx)

	migrationPath := filepath.Join("..", "..", "..", "migrations", "001_init.sql")
	sqlBytes, err := os.ReadFile(migrationPath)
	if err != nil {
		return err
	}

	statements := strings.Split(string(sqlBytes), ";")
	for _, stmt := range statements {
		trimmed := strings.TrimSpace(stmt)
		if trimmed == "" {
			continue
		}
		if _, err := conn.Exec(ctx, trimmed); err != nil {
			return fmt.Errorf("migration failed: %w", err)
		}
	}
	return nil
}

func hashToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return fmt.Sprintf("%x", h[:])
}

func waitForPostgres(ctx context.Context, connStr string) error {
	deadline := time.Now().Add(30 * time.Second)
	for {
		conn, err := pgx.Connect(ctx, connStr)
		if err == nil {
			conn.Close(ctx)
			return nil
		}
		if time.Now().After(deadline) {
			return err
		}
		time.Sleep(500 * time.Millisecond)
	}
}

func skipIfDockerUnavailable(t *testing.T) {
	t.Helper()

	if os.Getenv("SKIP_DOCKER_TESTS") != "" {
		t.Skip("SKIP_DOCKER_TESTS set")
	}

	host := os.Getenv("DOCKER_HOST")
	if strings.HasPrefix(host, "tcp://") {
		target := strings.TrimPrefix(host, "tcp://")
		conn, err := net.DialTimeout("tcp", target, 2*time.Second)
		if err != nil {
			t.Skip("docker host not reachable")
		}
		conn.Close()
		return
	}

	if runtime.GOOS == "windows" {
		return
	}

	conn, err := net.DialTimeout("unix", "/var/run/docker.sock", 2*time.Second)
	if err != nil {
		t.Skip("docker socket not available")
	}
	conn.Close()
}
