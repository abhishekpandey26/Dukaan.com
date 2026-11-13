// Package config reads all settings from environment variables.
// Every value has a default so the app "just works" on your laptop.
package config

import "os"

type Config struct {
	Port          string
	DatabaseURL   string
	RedisAddr     string // leave empty to disable caching
	JWTSecret     string
	AdminEmail    string // an admin user is created with these on startup
	AdminPassword string
}

func Load() Config {
	return Config{
		Port:          getEnv("PORT", "8080"),
		DatabaseURL:   getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/gocommerce?sslmode=disable"),
		RedisAddr:     getEnv("REDIS_ADDR", "localhost:6379"),
		JWTSecret:     getEnv("JWT_SECRET", "dev-secret-change-me"),
		AdminEmail:    getEnv("ADMIN_EMAIL", "admin@gocommerce.com"),
		AdminPassword: getEnv("ADMIN_PASSWORD", "admin123"),
	}
}

// getEnv returns the env variable, or fallback if it's not set.
func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
