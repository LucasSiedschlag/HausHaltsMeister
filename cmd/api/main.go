package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/LucasSiedschlag/HausHaltsMeister/internal/adapters/http/handlers"
	"github.com/LucasSiedschlag/HausHaltsMeister/internal/adapters/http/middleware"
	"github.com/LucasSiedschlag/HausHaltsMeister/internal/adapters/postgres"
	pgaccounts "github.com/LucasSiedschlag/HausHaltsMeister/internal/adapters/postgres/accounts"
	pgauth "github.com/LucasSiedschlag/HausHaltsMeister/internal/adapters/postgres/auth"
	pgbudget "github.com/LucasSiedschlag/HausHaltsMeister/internal/adapters/postgres/budget"
	pgcategories "github.com/LucasSiedschlag/HausHaltsMeister/internal/adapters/postgres/categories"
	pgcreditcard "github.com/LucasSiedschlag/HausHaltsMeister/internal/adapters/postgres/creditcard"
	pginvestments "github.com/LucasSiedschlag/HausHaltsMeister/internal/adapters/postgres/investments"
	pgjournal "github.com/LucasSiedschlag/HausHaltsMeister/internal/adapters/postgres/journal"
	pgledger "github.com/LucasSiedschlag/HausHaltsMeister/internal/adapters/postgres/ledger"
	pgreports "github.com/LucasSiedschlag/HausHaltsMeister/internal/adapters/postgres/reports"
	"github.com/LucasSiedschlag/HausHaltsMeister/internal/config"
	"github.com/LucasSiedschlag/HausHaltsMeister/internal/domain/accounts"
	"github.com/LucasSiedschlag/HausHaltsMeister/internal/domain/auth"
	"github.com/LucasSiedschlag/HausHaltsMeister/internal/domain/budget"
	"github.com/LucasSiedschlag/HausHaltsMeister/internal/domain/categories"
	"github.com/LucasSiedschlag/HausHaltsMeister/internal/domain/creditcard"
	"github.com/LucasSiedschlag/HausHaltsMeister/internal/domain/investments"
	"github.com/LucasSiedschlag/HausHaltsMeister/internal/domain/journal"
	"github.com/LucasSiedschlag/HausHaltsMeister/internal/domain/ledger"
	"github.com/LucasSiedschlag/HausHaltsMeister/internal/domain/reports"
	"github.com/labstack/echo/v4"
	echomw "github.com/labstack/echo/v4/middleware"
	echoSwagger "github.com/swaggo/echo-swagger"
)

func main() {
	cfg := config.Load()

	providers := map[string]auth.OAuthProvider{}
	if cfg.GoogleClientID != "" && cfg.GoogleClientSecret != "" && cfg.GoogleRedirectURL != "" {
		providers["google"] = auth.NewGoogleProvider(auth.OAuthProviderConfig{
			ClientID:     cfg.GoogleClientID,
			ClientSecret: cfg.GoogleClientSecret,
			RedirectURL:  cfg.GoogleRedirectURL,
		})
	}
	if cfg.GitHubClientID != "" && cfg.GitHubClientSecret != "" && cfg.GitHubRedirectURL != "" {
		providers["github"] = auth.NewGitHubProvider(auth.OAuthProviderConfig{
			ClientID:     cfg.GitHubClientID,
			ClientSecret: cfg.GitHubClientSecret,
			RedirectURL:  cfg.GitHubRedirectURL,
		})
	}

	ratelimiter := middleware.NewRateLimiter(5, 10*time.Minute)
	metrics := middleware.NewMetrics()

	e := echo.New()
	e.HideBanner = true
	e.Use(echomw.RequestID())
	e.Use(echomw.Recover())
	e.Use(middleware.RequestLogger())
	e.Use(metrics.Middleware())
	e.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})
	e.GET("/metrics", metrics.Handler)
	e.GET("/docs/openapi.yaml", func(c echo.Context) error {
		c.Response().Header().Set("Cache-Control", "no-store")
		c.Response().Header().Set(echo.HeaderContentType, "application/yaml")
		return c.File("docs/api/openapi.yaml")
	})
	e.GET("/docs", func(c echo.Context) error {
		return c.Redirect(http.StatusMovedPermanently, "/docs/index.html")
	})
	e.GET("/docs/", func(c echo.Context) error {
		return c.Redirect(http.StatusMovedPermanently, "/docs/index.html")
	})
	e.GET("/docs/*", echoSwagger.EchoWrapHandler(func(cfg *echoSwagger.Config) {
		cfg.URLs = []string{"/docs/openapi.yaml"}
	}))

	if cfg.DatabaseURL == "" {
		log.Fatal("DATABASE_URL required")
	}

	ctx := context.Background()
	store, err := postgres.NewStore(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("db connect: %v", err)
	}
	defer store.Close()

	authRepo := pgauth.NewRepository(store)
	ledgerRepo := pgledger.NewRepository(store)
	accountsRepo := pgaccounts.NewRepository(store)
	categoriesRepo := pgcategories.NewRepository(store)
	journalRepo := pgjournal.NewRepository(store)
	budgetRepo := pgbudget.NewRepository(store)
	investmentsRepo := pginvestments.NewRepository(store)
	creditCardRepo := pgcreditcard.NewRepository(store)
	reportsRepo := pgreports.NewRepository(store)

	authService := auth.NewService(authRepo, auth.ServiceConfig{
		JWTSecret:  cfg.JWTSecret,
		AccessTTL:  cfg.AccessTokenTTL,
		RefreshTTL: cfg.RefreshTokenTTL,
		Providers:  providers,
	})
	ledgerService := ledger.NewService(ledgerRepo)
	accountsService := accounts.NewService(accountsRepo)
	categoriesService := categories.NewService(categoriesRepo)
	journalService := journal.NewService(journalRepo)
	budgetService := budget.NewService(budgetRepo)
	investmentsService := investments.NewService(investmentsRepo)
	creditCardService := creditcard.NewService(creditCardRepo)
	reportsService := reports.NewService(reportsRepo)

	authHandler := &handlers.AuthHandler{
		Service:     authService,
		Config:      cfg,
		RateLimiter: ratelimiter,
	}
	group := e.Group("/auth")
	authHandler.Register(group)

	ledgerHandler := &handlers.LedgerHandler{
		Service: ledgerService,
	}
	ledgerGroup := e.Group("/ledgers", middleware.RequireAuth(authService))
	ledgerHandler.Register(ledgerGroup)

	accountsHandler := &handlers.AccountsHandler{
		Service: accountsService,
	}
	accountsGroup := e.Group("/ledgers/:ledgerId/accounts", middleware.RequireAuth(authService))
	accountsHandler.Register(accountsGroup)

	categoriesHandler := &handlers.CategoriesHandler{
		Service: categoriesService,
	}
	categoriesGroup := e.Group("/ledgers/:ledgerId/categories", middleware.RequireAuth(authService))
	categoriesHandler.Register(categoriesGroup)

	journalHandler := &handlers.JournalHandler{
		Service: journalService,
	}
	journalGroup := e.Group("/ledgers/:ledgerId/transactions", middleware.RequireAuth(authService))
	journalHandler.Register(journalGroup)

	budgetHandler := &handlers.BudgetHandler{
		Service: budgetService,
	}
	budgetGroup := e.Group("/ledgers/:ledgerId/budget", middleware.RequireAuth(authService))
	budgetHandler.Register(budgetGroup)

	investmentsHandler := &handlers.InvestmentsHandler{
		Service: investmentsService,
	}
	investmentsGroup := e.Group("/ledgers/:ledgerId/investments", middleware.RequireAuth(authService))
	investmentsHandler.Register(investmentsGroup)

	creditCardHandler := &handlers.CreditCardHandler{
		Service: creditCardService,
	}
	cardNetworksGroup := e.Group("/card-networks", middleware.RequireAuth(authService))
	creditCardHandler.RegisterNetworks(cardNetworksGroup)
	creditCardsGroup := e.Group("/ledgers/:ledgerId/credit-cards", middleware.RequireAuth(authService))
	creditCardHandler.Register(creditCardsGroup)

	reportsHandler := &handlers.ReportsHandler{
		Service: reportsService,
	}
	reportsGroup := e.Group("/ledgers/:ledgerId/reports", middleware.RequireAuth(authService))
	reportsHandler.Register(reportsGroup)

	startServer(e, cfg.HTTPAddr)
}

func startServer(e *echo.Echo, addr string) {
	go func() {
		if err := e.Start(addr); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)
	<-shutdown

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := e.Shutdown(ctx); err != nil {
		log.Printf("shutdown error: %v", err)
	}
}
