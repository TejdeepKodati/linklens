package config

import (
	"os"
)

type Config struct {
	Env         string
	Port        string
	DatabaseURL string
	RedisURL    string
	JWTSecret   string
	JWTExpiry   string // e.g. "24h"
	BaseURL     string // e.g. https://linklens.io (used when generating short URLs)
}

func Load() *Config {
	return &Config{
		Env:         getEnv("ENV", "development"),
		Port:        getEnv("PORT", "8080"),
		DatabaseURL: getEnv("DATABASE_URL", "postgres://postgres:password@localhost:5432/linklens?sslmode=disable"),
		RedisURL:    getEnv("REDIS_URL", "redis://localhost:6379"),
		JWTSecret:   getEnv("JWT_SECRET", "supersecret-change-me-in-production"),
		JWTExpiry:   getEnv("JWT_EXPIRY", "24h"),
		BaseURL:     getEnv("BASE_URL", "http://localhost:8080"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
