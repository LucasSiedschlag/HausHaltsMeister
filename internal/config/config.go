package config

import (
	"os"
	"strconv"
	"time"
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
	return Config{
		DatabaseURL:           getEnv("DATABASE_URL", ""),
		HTTPAddr:              getEnv("HTTP_ADDR", ":8080"),
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
