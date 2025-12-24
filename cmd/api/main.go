package main

import (
	"context"
	"log"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	httpAdapter "github.com/LucasSiedschlag/HausHaltsMeister/internal/adapters/http"
	"github.com/LucasSiedschlag/HausHaltsMeister/internal/adapters/postgres"
	"github.com/LucasSiedschlag/HausHaltsMeister/internal/config"
	"github.com/LucasSiedschlag/HausHaltsMeister/internal/db"
	"github.com/LucasSiedschlag/HausHaltsMeister/internal/domain/budget"
	"github.com/LucasSiedschlag/HausHaltsMeister/internal/domain/cashflow"
	"github.com/LucasSiedschlag/HausHaltsMeister/internal/domain/category"
	"github.com/LucasSiedschlag/HausHaltsMeister/internal/domain/installment"
	"github.com/LucasSiedschlag/HausHaltsMeister/internal/domain/payment"
	"github.com/LucasSiedschlag/HausHaltsMeister/internal/domain/picuinha"
)

// @title           HausHaltsMeister API
// @version         1.0
// @description     API for HausHaltsMeister (Personal Finance Management).
// @termsOfService  http://swagger.io/terms/

// @contact.name   Lucas Siedschlag
// @contact.url    http://www.swagger.io/support
// @contact.email  support@swagger.io

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost:8080
// @BasePath  /
// @schemes   http

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name X-App-Token
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
	cfRepo := postgres.NewCashFlowRepository(pool)
	bgRepo := postgres.NewBudgetRepository(pool)
	picRepo := postgres.NewPicuinhaRepository(pool)
	payRepo := postgres.NewPaymentRepository(pool)
	instRepo := postgres.NewInstallmentRepository(pool)

	// 4. Setup services
	catService := category.NewService(catRepo)
	cfService := cashflow.NewService(cfRepo, catRepo)
	bgService := budget.NewService(bgRepo, catRepo, cfRepo)
	picService := picuinha.NewService(picRepo, cfService, catRepo)
	payService := payment.NewService(payRepo)
	instService := installment.NewService(instRepo, cfService, payRepo)

	// 5. Setup handlers
	catHandler := httpAdapter.NewCategoryHandler(catService)
	cfHandler := httpAdapter.NewCashFlowHandler(cfService)
	bgHandler := httpAdapter.NewBudgetHandler(bgService)
	picHandler := httpAdapter.NewPicuinhaHandler(picService)
	payHandler := httpAdapter.NewPaymentHandler(payService)
	instHandler := httpAdapter.NewInstallmentHandler(instService)

	// 6. Setup Echo
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// 7. Register routes
	httpAdapter.RegisterCategoryRoutes(e, catHandler)
	httpAdapter.RegisterCashFlowRoutes(e, cfHandler)
	httpAdapter.RegisterBudgetRoutes(e, bgHandler)
	httpAdapter.RegisterPicuinhaRoutes(e, picHandler)
	httpAdapter.RegisterPaymentRoutes(e, payHandler)
	httpAdapter.RegisterInstallmentRoutes(e, instHandler)
	httpAdapter.RegisterSwaggerRoutes(e)

	// 8. Start server
	log.Printf("Starting server on port %s", cfg.Port)
	if err := e.Start(":" + cfg.Port); err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
