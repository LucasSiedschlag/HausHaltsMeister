package config

import (
	"os"
)

type Config struct {
	DBUrl string
	Port  string
}

func Load() *Config {
	dbUrl := os.Getenv("DATABASE_URL")
	if dbUrl == "" {
		dbUrl = "postgres://postgres:postgres@localhost:5432/haushaltsmeister?sslmode=disable"
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	return &Config{
		DBUrl: dbUrl,
		Port:  port,
	}
}
