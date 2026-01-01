package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/LucasSiedschlag/HausHaltsMeister/internal/config"
	"github.com/LucasSiedschlag/HausHaltsMeister/internal/domain/auth"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
)

type fakeAuthService struct {
	loginErr error
}

func (f fakeAuthService) SignUp(ctx context.Context, email, password, displayName, userAgent, ip string) (auth.AuthResult, error) {
	return auth.AuthResult{}, nil
}

func (f fakeAuthService) Login(ctx context.Context, email, password, userAgent, ip string) (auth.AuthResult, error) {
	return auth.AuthResult{}, f.loginErr
}

func (f fakeAuthService) Refresh(ctx context.Context, refreshToken, userAgent, ip string) (auth.AuthResult, error) {
	return auth.AuthResult{}, nil
}

func (f fakeAuthService) Logout(ctx context.Context, refreshToken string) error {
	return nil
}

func (f fakeAuthService) Me(ctx context.Context, userID string) (auth.User, error) {
	return auth.User{}, nil
}

func (f fakeAuthService) StartOAuth(ctx context.Context, provider, redirectURI string) (string, error) {
	return "", nil
}

func (f fakeAuthService) HandleOAuthCallback(ctx context.Context, provider, code, state, userAgent, ip string) (auth.AuthResult, string, error) {
	return auth.AuthResult{}, "", nil
}

func (f fakeAuthService) ParseAccessToken(token string) (auth.AccessTokenClaims, error) {
	return auth.AccessTokenClaims{}, nil
}

func TestLoginReturnsInvalidCredentials(t *testing.T) {
	e := echo.New()
	payload := map[string]string{"email": "user@example.com", "password": "wrong"}
	body, err := json.Marshal(payload)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := AuthHandler{
		Service:     fakeAuthService{loginErr: auth.ErrInvalidCredentials},
		Config:      config.Config{RefreshCookieName: "hhm_refresh"},
		RateLimiter: nil,
	}

	err = handler.Login(c)
	require.NoError(t, err)
	require.Equal(t, http.StatusUnauthorized, rec.Code)

	var resp ErrorResponse
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&resp))
	require.Equal(t, "AUTH_INVALID_CREDENTIALS", resp.Code)
}
