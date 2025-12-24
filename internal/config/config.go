package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DBUrl string
	Port  string
}

func Load() *Config {
	// Load .env file if it exists, but don't fail if missing (environment might be set otherwise)
	_ = godotenv.Load()

	// Construct DB URL from individual vars if DATABASE_URL is not set
	dbUrl := os.Getenv("DATABASE_URL")
	if dbUrl == "" {
		host := getEnvOrDefault("POSTGRES_HOST", "localhost")
		port := getEnvOrDefault("POSTGRES_PORT", "5432")
		user := getEnvOrDefault("POSTGRES_USER", "postgres")
		pass := getEnvOrDefault("POSTGRES_PASSWORD", "postgres")
		dbname := getEnvOrDefault("POSTGRES_DB", "haushaltsmeister")
		sslmode := getEnvOrDefault("POSTGRES_SSLMODE", "disable")

		dbUrl = fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
			user, pass, host, port, dbname, sslmode,
		)
	}

	port := getEnvOrDefault("PORT", "8080")

	return &Config{
		DBUrl: dbUrl,
		Port:  port,
	}
}

func getEnvOrDefault(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
