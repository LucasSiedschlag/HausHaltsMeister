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
	lastRevoke  string
	createdUser User
	createdHash string
	userByEmail User
	userByID    User
	secretHash  string
	createErr   error
	sessionSet  bool
}

func (f *fakeRepo) CreateUserWithPassword(ctx context.Context, email, displayName, passwordHash string) (User, error) {
	if f.createErr != nil {
		return User{}, f.createErr
	}
	f.createdHash = passwordHash
	if f.createdUser.ID != "" {
		return f.createdUser, nil
	}
	return User{ID: "user-1", Email: email, DisplayName: displayName, IsActive: true}, nil
}

func (f *fakeRepo) CreateUserWithIdentity(ctx context.Context, params CreateIdentityParams) (User, AuthIdentity, error) {
	return User{}, AuthIdentity{}, nil
}

func (f *fakeRepo) CreateAuthIdentity(ctx context.Context, params CreateIdentityParams) (AuthIdentity, error) {
	return AuthIdentity{}, nil
}

func (f *fakeRepo) UpdateAuthIdentityProfile(ctx context.Context, userID, provider, displayName, avatarURL string) error {
	return nil
}
func (f *fakeRepo) GetLatestAuthIdentityForUser(ctx context.Context, userID string) (AuthIdentity, error) {
	return AuthIdentity{}, ErrNotFound
}

func (f *fakeRepo) GetUserByEmail(ctx context.Context, email string) (User, error) {
	if f.userByEmail.ID == "" {
		return User{}, ErrNotFound
	}
	return f.userByEmail, nil
}

func (f *fakeRepo) GetUserByID(ctx context.Context, userID string) (User, error) {
	if f.userByID.ID == "" {
		return User{}, ErrNotFound
	}
	return f.userByID, nil
}

func (f *fakeRepo) GetAuthSecretHash(ctx context.Context, userID string) (string, error) {
	if f.secretHash == "" {
		return "", ErrNotFound
	}
	return f.secretHash, nil
}

func (f *fakeRepo) CreateAuthSession(ctx context.Context, params CreateSessionParams) (AuthSession, error) {
	f.sessionSet = true
	return AuthSession{ID: "session-1", UserID: params.UserID, ExpiresAt: params.ExpiresAt, IsPersistent: params.IsPersistent}, nil
}

func (f *fakeRepo) GetAuthSessionByRefreshTokenHash(ctx context.Context, refreshTokenHash string) (AuthSession, error) {
	return AuthSession{
		ID:           "session-1",
		UserID:       "user-1",
		ExpiresAt:    time.Now().Add(time.Hour),
		IsPersistent: true,
	}, nil
}

func (f *fakeRepo) RotateAuthSession(ctx context.Context, params RotateSessionParams) (AuthSession, User, error) {
	f.lastOldHash = params.OldRefreshTokenHash
	f.lastNewHash = params.NewRefreshTokenHash
	return AuthSession{ID: "session-1", UserID: "user-1", ExpiresAt: time.Now().Add(time.Hour), IsPersistent: params.IsPersistent}, User{ID: "user-1", IsActive: true}, nil
}

func (f *fakeRepo) RevokeAuthSession(ctx context.Context, refreshTokenHash string) error {
	f.lastRevoke = refreshTokenHash
	return nil
}

func (f *fakeRepo) ListAuthSessionsByUser(ctx context.Context, userID string) ([]AuthSessionDetails, error) {
	return []AuthSessionDetails{}, nil
}

func (f *fakeRepo) RevokeAuthSessionByID(ctx context.Context, userID, sessionID string) error {
	return nil
}

func (f *fakeRepo) RevokeAllAuthSessionsForUser(ctx context.Context, userID string) error {
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
		RefreshTTL:        30 * 24 * time.Hour,
		RefreshSessionTTL: 7 * 24 * time.Hour,
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

func TestSignUpCreatesSession(t *testing.T) {
	repo := &fakeRepo{createdUser: User{ID: "user-1", Email: "user@example.com", DisplayName: "User", IsActive: true}}
	service := NewService(repo, ServiceConfig{
		JWTSecret:  "test",
		AccessTTL:  15 * time.Minute,
		RefreshTTL:        30 * 24 * time.Hour,
		RefreshSessionTTL: 7 * 24 * time.Hour,
		Providers:  map[string]OAuthProvider{},
	})

	result, err := service.SignUp(context.Background(), "user@example.com", "password123", "User", "agent", "127.0.0.1")
	require.NoError(t, err)
	require.NotEmpty(t, result.Tokens.AccessToken)
	require.NotEmpty(t, result.Tokens.RefreshToken)
	require.True(t, repo.sessionSet)
}

func TestLoginCreatesSession(t *testing.T) {
	hash, err := HashPassword("password123")
	require.NoError(t, err)

	repo := &fakeRepo{
		userByEmail: User{ID: "user-1", Email: "user@example.com", DisplayName: "User", IsActive: true},
		secretHash:  hash,
	}
	service := NewService(repo, ServiceConfig{
		JWTSecret:  "test",
		AccessTTL:  15 * time.Minute,
		RefreshTTL:        30 * 24 * time.Hour,
		RefreshSessionTTL: 7 * 24 * time.Hour,
		Providers:  map[string]OAuthProvider{},
	})

	result, err := service.Login(context.Background(), "user@example.com", "password123", "agent", "127.0.0.1", true)
	require.NoError(t, err)
	require.NotEmpty(t, result.Tokens.AccessToken)
	require.True(t, repo.sessionSet)
}

func TestLogoutRevokesSession(t *testing.T) {
	repo := &fakeRepo{}
	service := NewService(repo, ServiceConfig{
		JWTSecret:  "test",
		AccessTTL:  15 * time.Minute,
		RefreshTTL:        30 * 24 * time.Hour,
		RefreshSessionTTL: 7 * 24 * time.Hour,
		Providers:  map[string]OAuthProvider{},
	})

	err := service.Logout(context.Background(), "refresh-token")
	require.NoError(t, err)
	require.Equal(t, hashToken("refresh-token"), repo.lastRevoke)
}

func TestMeReturnsUser(t *testing.T) {
	repo := &fakeRepo{userByID: User{ID: "user-1", Email: "user@example.com", DisplayName: "User", IsActive: true}}
	service := NewService(repo, ServiceConfig{
		JWTSecret:  "test",
		AccessTTL:  15 * time.Minute,
		RefreshTTL:        30 * 24 * time.Hour,
		RefreshSessionTTL: 7 * 24 * time.Hour,
		Providers:  map[string]OAuthProvider{},
	})

	user, err := service.Me(context.Background(), "user-1")
	require.NoError(t, err)
	require.Equal(t, "user-1", user.ID)
}
