package middleware

import (
	"context"
	"strings"

	"github.com/LucasSiedschlag/HausHaltsMeister/internal/adapters/http/httpx"
	"github.com/LucasSiedschlag/HausHaltsMeister/internal/domain/auth"
	"github.com/labstack/echo/v4"
)

type AuthService interface {
	ParseAccessToken(token string) (auth.AccessTokenClaims, error)
	Me(ctx context.Context, userID string) (auth.User, error)
}

func RequireAuth(service AuthService) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
				return httpx.WriteError(c, 401, "AUTH_INVALID_CREDENTIALS", "Credenciais invalidas", nil)
			}

			token := strings.TrimPrefix(authHeader, "Bearer ")
			claims, err := service.ParseAccessToken(token)
			if err != nil {
				return httpx.WriteError(c, 401, "AUTH_INVALID_CREDENTIALS", "Credenciais invalidas", nil)
			}

			user, err := service.Me(c.Request().Context(), claims.Sub)
			if err != nil {
				return httpx.WriteAuthError(c, err)
			}
			httpx.SetUser(c, user)
			return next(c)
		}
	}
}
