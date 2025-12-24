package main

import (
	"context"
	"log"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	httpAdapter "github.com/seuuser/cashflow/internal/adapters/http"
	"github.com/seuuser/cashflow/internal/adapters/postgres"
	"github.com/seuuser/cashflow/internal/config"
	"github.com/seuuser/cashflow/internal/db"
	"github.com/seuuser/cashflow/internal/domain/category"
)

func main() {
	// 1. Load config
	cfg := config.Load()

	// 2. Setup DB
	ctx := context.Background()
	pool, err := db.NewPool(ctx, cfg.DBUrl)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v", err)
	}
	defer pool.Close()

	// 3. Setup repositories
	catRepo := postgres.NewCategoryRepository(pool)

	// 4. Setup services
	catService := category.NewService(catRepo)

	// 5. Setup handlers
	catHandler := httpAdapter.NewCategoryHandler(catService)

	// 6. Setup Echo
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// 7. Register routes
	httpAdapter.RegisterCategoryRoutes(e, catHandler)

	// 8. Start server
	log.Printf("Starting server on port %s", cfg.Port)
	if err := e.Start(":" + cfg.Port); err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
