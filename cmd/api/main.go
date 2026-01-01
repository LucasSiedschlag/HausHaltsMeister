package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	httpapi "github.com/LucasSiedschlag/HausHaltsMeister/internal/adapters/http"
	"github.com/LucasSiedschlag/HausHaltsMeister/internal/adapters/postgres"
	"github.com/LucasSiedschlag/HausHaltsMeister/internal/config"
	"github.com/LucasSiedschlag/HausHaltsMeister/internal/domain/accounts"
	"github.com/LucasSiedschlag/HausHaltsMeister/internal/domain/auth"
	"github.com/LucasSiedschlag/HausHaltsMeister/internal/domain/categories"
	"github.com/LucasSiedschlag/HausHaltsMeister/internal/domain/journal"
	"github.com/LucasSiedschlag/HausHaltsMeister/internal/domain/ledger"
	"github.com/labstack/echo/v4"
	echomw "github.com/labstack/echo/v4/middleware"
)

func main() {
	cfg := config.Load()
	if cfg.DatabaseURL == "" {
		log.Fatal("DATABASE_URL required")
	}

	ctx := context.Background()
	store, err := postgres.NewStore(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("db connect: %v", err)
	}
	defer store.Close()

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

	authService := auth.NewService(store, auth.ServiceConfig{
		JWTSecret:  cfg.JWTSecret,
		AccessTTL:  cfg.AccessTokenTTL,
		RefreshTTL: cfg.RefreshTokenTTL,
		Providers:  providers,
	})
	ledgerService := ledger.NewService(store)
	accountsService := accounts.NewService(store)
	categoriesService := categories.NewService(store)
	journalService := journal.NewService(store)

	ratelimiter := httpapi.NewRateLimiter(5, 10*time.Minute)
	
	e := echo.New()
	e.HideBanner = true
	e.Use(echomw.Recover())
	e.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})

	authHandler := &httpapi.AuthHandler{
		Service:     authService,
		Config:      cfg,
		RateLimiter: ratelimiter,
	}
	group := e.Group("/auth")
	authHandler.Register(group)

	ledgerHandler := &httpapi.LedgerHandler{
		Service: ledgerService,
	}
	ledgerGroup := e.Group("/ledgers", httpapi.RequireAuth(authService))
	ledgerHandler.Register(ledgerGroup)

	accountsHandler := &httpapi.AccountsHandler{
		Service: accountsService,
	}
	accountsGroup := e.Group("/ledgers/:ledgerId/accounts", httpapi.RequireAuth(authService))
	accountsHandler.Register(accountsGroup)

	categoriesHandler := &httpapi.CategoriesHandler{
		Service: categoriesService,
	}
	categoriesGroup := e.Group("/ledgers/:ledgerId/categories", httpapi.RequireAuth(authService))
	categoriesHandler.Register(categoriesGroup)

	journalHandler := &httpapi.JournalHandler{
		Service: journalService,
	}
	journalGroup := e.Group("/ledgers/:ledgerId/transactions", httpapi.RequireAuth(authService))
	journalHandler.Register(journalGroup)

	go func() {
		if err := e.Start(cfg.HTTPAddr); err != nil && err != http.ErrServerClosed {
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
