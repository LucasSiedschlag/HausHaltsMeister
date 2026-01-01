package postgres

import (
	"context"
	"crypto/sha256"
	"fmt"
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

func hashToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return fmt.Sprintf("%x", h[:])
}
