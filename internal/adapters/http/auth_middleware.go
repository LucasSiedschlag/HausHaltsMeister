package httpapi

import (
	"strings"

	"github.com/LucasSiedschlag/HausHaltsMeister/internal/domain/auth"
	"github.com/labstack/echo/v4"
)

const userContextKey = "user"

func RequireAuth(service AuthService) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
				return WriteError(c, 401, "AUTH_INVALID_CREDENTIALS", "Credenciais invalidas", nil)
			}

			token := strings.TrimPrefix(authHeader, "Bearer ")
			claims, err := service.ParseAccessToken(token)
			if err != nil {
				return WriteError(c, 401, "AUTH_INVALID_CREDENTIALS", "Credenciais invalidas", nil)
			}

			user, err := service.Me(c.Request().Context(), claims.Sub)
			if err != nil {
				return WriteAuthError(c, err)
			}
			c.Set(userContextKey, user)
			return next(c)
		}
	}
}

func GetUser(c echo.Context) (auth.User, bool) {
	value := c.Get(userContextKey)
	if value == nil {
		return auth.User{}, false
	}
	user, ok := value.(auth.User)
	return user, ok
}
