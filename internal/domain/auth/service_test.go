package auth

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type fakeRepo struct {
	lastOldHash string
	lastNewHash string
}

func (f *fakeRepo) CreateUserWithPassword(ctx context.Context, email, displayName, passwordHash string) (User, error) {
	return User{}, nil
}

func (f *fakeRepo) CreateUserWithIdentity(ctx context.Context, params CreateIdentityParams) (User, AuthIdentity, error) {
	return User{}, AuthIdentity{}, nil
}

func (f *fakeRepo) CreateAuthIdentity(ctx context.Context, params CreateIdentityParams) (AuthIdentity, error) {
	return AuthIdentity{}, nil
}

func (f *fakeRepo) GetUserByEmail(ctx context.Context, email string) (User, error) {
	return User{}, ErrNotFound
}

func (f *fakeRepo) GetUserByID(ctx context.Context, userID string) (User, error) {
	return User{}, ErrNotFound
}

func (f *fakeRepo) GetAuthSecretHash(ctx context.Context, userID string) (string, error) {
	return "", ErrNotFound
}

func (f *fakeRepo) CreateAuthSession(ctx context.Context, params CreateSessionParams) (AuthSession, error) {
	return AuthSession{}, nil
}

func (f *fakeRepo) RotateAuthSession(ctx context.Context, params RotateSessionParams) (AuthSession, User, error) {
	f.lastOldHash = params.OldRefreshTokenHash
	f.lastNewHash = params.NewRefreshTokenHash
	return AuthSession{ID: "session-1", UserID: "user-1", ExpiresAt: time.Now().Add(time.Hour)}, User{ID: "user-1", IsActive: true}, nil
}

func (f *fakeRepo) RevokeAuthSession(ctx context.Context, refreshTokenHash string) error {
	return nil
}

func (f *fakeRepo) CreateOAuthState(ctx context.Context, provider, state, codeVerifier, redirectURI string) (OAuthState, error) {
	return OAuthState{}, nil
}

func (f *fakeRepo) GetOAuthState(ctx context.Context, state string) (OAuthState, error) {
	return OAuthState{}, ErrNotFound
}

func (f *fakeRepo) MarkOAuthStateUsed(ctx context.Context, stateID string) error {
	return nil
}

func (f *fakeRepo) GetAuthIdentityByProvider(ctx context.Context, provider, providerUserID string) (AuthIdentity, error) {
	return AuthIdentity{}, ErrNotFound
}

func (f *fakeRepo) UpdateAuthIdentityLogin(ctx context.Context, userID, provider string) error {
	return nil
}

func (f *fakeRepo) MarkUserEmailVerified(ctx context.Context, userID string) error {
	return nil
}

func TestRefreshRotatesToken(t *testing.T) {
	repo := &fakeRepo{}
	service := NewService(repo, ServiceConfig{
		JWTSecret:  "test",
		AccessTTL:  15 * time.Minute,
		RefreshTTL: 30 * 24 * time.Hour,
		Providers:  map[string]OAuthProvider{},
	})

	oldRefresh := "old-refresh-token"
	result, err := service.Refresh(context.Background(), oldRefresh, "", "")
	require.NoError(t, err)
	require.NotEmpty(t, result.Tokens.RefreshToken)
	require.NotEqual(t, oldRefresh, result.Tokens.RefreshToken)
	require.Equal(t, hashToken(oldRefresh), repo.lastOldHash)
	require.NotEmpty(t, repo.lastNewHash)
}
