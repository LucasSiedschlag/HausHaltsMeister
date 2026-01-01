package config

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL          string
	HTTPAddr             string
	JWTSecret            string
	AccessTokenTTL       time.Duration
	RefreshTokenTTL      time.Duration
	RefreshCookieName    string
	RefreshCookieDomain  string
	RefreshCookieSecure  bool
	RefreshCookieSameSite string

	GoogleClientID     string
	GoogleClientSecret string
	GoogleRedirectURL  string

	GitHubClientID     string
	GitHubClientSecret string
	GitHubRedirectURL  string
}

func Load() Config {
	_ = godotenv.Load()

	databaseURL := getEnv("DATABASE_URL", "")
	if databaseURL == "" {
		databaseURL = buildDatabaseURL()
	}

	httpAddr := getEnv("HTTP_ADDR", "")
	if httpAddr == "" {
		httpAddr = ":" + getEnv("PORT", "8080")
	}

	return Config{
		DatabaseURL:           databaseURL,
		HTTPAddr:              httpAddr,
		JWTSecret:             getEnv("JWT_SECRET", "change-me"),
		AccessTokenTTL:        time.Duration(getEnvInt("ACCESS_TOKEN_TTL_SECONDS", 900)) * time.Second,
		RefreshTokenTTL:       time.Duration(getEnvInt("REFRESH_TOKEN_TTL_DAYS", 30)) * 24 * time.Hour,
		RefreshCookieName:     getEnv("REFRESH_COOKIE_NAME", "hhm_refresh"),
		RefreshCookieDomain:   getEnv("REFRESH_COOKIE_DOMAIN", ""),
		RefreshCookieSecure:   getEnvBool("REFRESH_COOKIE_SECURE", true),
		RefreshCookieSameSite: getEnv("REFRESH_COOKIE_SAMESITE", "Lax"),

		GoogleClientID:     getEnv("GOOGLE_CLIENT_ID", ""),
		GoogleClientSecret: getEnv("GOOGLE_CLIENT_SECRET", ""),
		GoogleRedirectURL:  getEnv("GOOGLE_REDIRECT_URL", ""),

		GitHubClientID:     getEnv("GITHUB_CLIENT_ID", ""),
		GitHubClientSecret: getEnv("GITHUB_CLIENT_SECRET", ""),
		GitHubRedirectURL:  getEnv("GITHUB_REDIRECT_URL", ""),
	}
}

func buildDatabaseURL() string {
	host := getEnv("POSTGRES_HOST", "")
	user := getEnv("POSTGRES_USER", "")
	pass := getEnv("POSTGRES_PASSWORD", "")
	name := getEnv("POSTGRES_DB", "")
	port := getEnv("POSTGRES_PORT", "")
	if port == "" {
		port = getEnv("POSTGRES_HOST_PORT", "")
	}
	if port == "" {
		port = "5432"
	}
	if host == "" || user == "" || name == "" {
		return ""
	}
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		url.QueryEscape(user),
		url.QueryEscape(pass),
		host,
		port,
		name,
	)
}

func getEnv(key, def string) string {
	value := os.Getenv(key)
	if value == "" {
		return def
	}
	return value
}

func getEnvInt(key string, def int) int {
	value := os.Getenv(key)
	if value == "" {
		return def
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return def
	}
	return parsed
}

func getEnvBool(key string, def bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return def
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return def
	}
	return parsed
}
