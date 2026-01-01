package httpx

import (
	"github.com/LucasSiedschlag/HausHaltsMeister/internal/domain/auth"
	"github.com/labstack/echo/v4"
)

const userContextKey = "user"

func SetUser(c echo.Context, user auth.User) {
	c.Set(userContextKey, user)
}

func GetUser(c echo.Context) (auth.User, bool) {
	value := c.Get(userContextKey)
	if value == nil {
		return auth.User{}, false
	}
	user, ok := value.(auth.User)
	return user, ok
}
