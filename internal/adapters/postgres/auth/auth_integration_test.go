package auth

import (
	"context"
	"crypto/sha256"
	"fmt"
	"testing"
	"time"

	pgstore "github.com/LucasSiedschlag/HausHaltsMeister/internal/adapters/postgres"
	"github.com/LucasSiedschlag/HausHaltsMeister/internal/adapters/postgres/postgrestest"
	authdomain "github.com/LucasSiedschlag/HausHaltsMeister/internal/domain/auth"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

func TestRefreshRotationPersists(t *testing.T) {
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
	service := authdomain.NewService(repo, authdomain.ServiceConfig{
		JWTSecret:  "test",
		AccessTTL:  15 * time.Minute,
		RefreshTTL: 30 * 24 * time.Hour,
		Providers:  map[string]authdomain.OAuthProvider{},
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
