package http

import (
	"net/http"

	"github.com/labstack/echo/v4"
	echoSwagger "github.com/swaggo/echo-swagger"

	_ "github.com/LucasSiedschlag/HausHaltsMeister/internal/adapters/http/docs" // Import generated docs
)

// RegisterSwaggerRoutes registers the routes for Swagger UI and OpenAPI spec.
func RegisterSwaggerRoutes(e *echo.Echo) {
	// 0. Health check
	e.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})

	// 1. Serve the Swagger UI page (redirect /swagger to /swagger/index.html)
	e.GET("/swagger/*", echoSwagger.WrapHandler)
}
