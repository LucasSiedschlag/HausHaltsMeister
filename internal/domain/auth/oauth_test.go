package auth

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type timeoutErr struct{}

func (timeoutErr) Error() string   { return "timeout" }
func (timeoutErr) Timeout() bool   { return true }
func (timeoutErr) Temporary() bool { return true }

type fakeOAuthProvider struct {
	exchangeErr error
}

func (f fakeOAuthProvider) Name() string { return "fake" }

func (f fakeOAuthProvider) AuthorizeURL(state, codeChallenge, redirectURI string) (string, error) {
	return "https://example.com/auth", nil
}

func (f fakeOAuthProvider) Exchange(ctx context.Context, code, codeVerifier, redirectURI string) (OAuthProfile, error) {
	return OAuthProfile{}, f.exchangeErr
}

type oauthRepo struct{}

func (o *oauthRepo) CreateUserWithPassword(ctx context.Context, email, displayName, passwordHash string) (User, error) {
	return User{}, ErrNotFound
}
func (o *oauthRepo) CreateUserWithIdentity(ctx context.Context, params CreateIdentityParams) (User, AuthIdentity, error) {
	return User{}, AuthIdentity{}, ErrNotFound
}
func (o *oauthRepo) CreateAuthIdentity(ctx context.Context, params CreateIdentityParams) (AuthIdentity, error) {
	return AuthIdentity{}, ErrNotFound
}
func (o *oauthRepo) GetUserByEmail(ctx context.Context, email string) (User, error) {
	return User{}, ErrNotFound
}
func (o *oauthRepo) GetUserByID(ctx context.Context, userID string) (User, error) {
	return User{}, ErrNotFound
}
func (o *oauthRepo) GetAuthSecretHash(ctx context.Context, userID string) (string, error) {
	return "", ErrNotFound
}
func (o *oauthRepo) CreateAuthSession(ctx context.Context, params CreateSessionParams) (AuthSession, error) {
	return AuthSession{}, ErrNotFound
}
func (o *oauthRepo) RotateAuthSession(ctx context.Context, params RotateSessionParams) (AuthSession, User, error) {
	return AuthSession{}, User{}, ErrNotFound
}
func (o *oauthRepo) RevokeAuthSession(ctx context.Context, refreshTokenHash string) error {
	return ErrNotFound
}
func (o *oauthRepo) CreateOAuthState(ctx context.Context, provider, state, codeVerifier, redirectURI string) (OAuthState, error) {
	return OAuthState{}, ErrNotFound
}
func (o *oauthRepo) GetOAuthState(ctx context.Context, state string) (OAuthState, error) {
	return OAuthState{
		ID:           "state-1",
		Provider:     "fake",
		State:        state,
		CodeVerifier: "verifier",
		RedirectURI:  "https://app.example.com/auth/callback",
		ExpiresAt:    time.Now().Add(10 * time.Minute),
	}, nil
}
func (o *oauthRepo) MarkOAuthStateUsed(ctx context.Context, stateID string) error {
	return nil
}
func (o *oauthRepo) GetAuthIdentityByProvider(ctx context.Context, provider, providerUserID string) (AuthIdentity, error) {
	return AuthIdentity{}, ErrNotFound
}
func (o *oauthRepo) UpdateAuthIdentityLogin(ctx context.Context, userID, provider string) error {
	return nil
}
func (o *oauthRepo) MarkUserEmailVerified(ctx context.Context, userID string) error {
	return nil
}

func TestClassifyOAuthError(t *testing.T) {
	require.Equal(t, ErrGatewayTimeout, classifyOAuthError(context.DeadlineExceeded))
	require.Equal(t, ErrGatewayTimeout, classifyOAuthError(timeoutErr{}))
	require.Equal(t, ErrServiceUnavailable, classifyOAuthError(errors.New("boom")))
}

func TestOAuthCallbackPropagatesGatewayTimeout(t *testing.T) {
	service := NewService(&oauthRepo{}, ServiceConfig{
		JWTSecret:  "test",
		AccessTTL:  15 * time.Minute,
		RefreshTTL: 30 * 24 * time.Hour,
		Providers: map[string]OAuthProvider{
			"fake": fakeOAuthProvider{exchangeErr: ErrGatewayTimeout},
		},
	})

	_, _, err := service.HandleOAuthCallback(context.Background(), "fake", "code", "state", "agent", "127.0.0.1")
	require.Equal(t, ErrGatewayTimeout, err)
}
