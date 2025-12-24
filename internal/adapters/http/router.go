package http

import (
	"github.com/labstack/echo/v4"
	"github.com/seuuser/cashflow/internal/domain/cashflow"
)

func NewRouter(cfService *cashflow.Service) *echo.Echo {
	e := echo.New()

	// Saúde
	e.GET("/health", func(c echo.Context) error {
		return c.JSON(200, map[string]string{"status": "ok"})
	})

	// CashFlow Handlers
	RegisterCashFlowRoutes(e, cfService)

	return e
}

