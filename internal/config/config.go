// Package config loads application configuration from environment variables.
package config

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// Config holds all application configuration values.
type Config struct {
	Port               string
	MongoURI           string
	MongoDB            string
	JWTSecret          string
	AccessTokenTTL     time.Duration
	RefreshTokenTTL    time.Duration
	CORSAllowedOrigins []string
}

// Load reads configuration from .env (if present) and environment variables.
// In production, .env may not exist — environment variables are used directly.
func Load() (*Config, error) {
	// Best-effort: load .env in dev; ignore error if file is missing.
	_ = godotenv.Load()

	accessTTL, err := time.ParseDuration(getEnv("ACCESS_TOKEN_TTL", "15m"))
	if err != nil {
		return nil, fmt.Errorf("config: invalid ACCESS_TOKEN_TTL: %w", err)
	}
	refreshTTL, err := time.ParseDuration(getEnv("REFRESH_TOKEN_TTL", "720h"))
	if err != nil {
		return nil, fmt.Errorf("config: invalid REFRESH_TOKEN_TTL: %w", err)
	}

	jwtSecret := getEnv("JWT_SECRET", "")
	if jwtSecret == "" {
		return nil, fmt.Errorf("config: JWT_SECRET is required")
	}

	origins := getEnv("CORS_ALLOWED_ORIGINS", "*")

	return &Config{
		Port:               getEnv("PORT", "8080"),
		MongoURI:           getEnv("MONGODB_URI", "mongodb://localhost:27017"),
		MongoDB:            getEnv("MONGODB_DB", "foodsvc"),
		JWTSecret:          jwtSecret,
		AccessTokenTTL:     accessTTL,
		RefreshTokenTTL:    refreshTTL,
		CORSAllowedOrigins: splitOrigins(origins),
	}, nil
}

// getEnv returns the value of an environment variable or a fallback default.
func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// splitOrigins splits a comma-separated origins string into a slice.
func splitOrigins(raw string) []string {
	parts := strings.Split(raw, ",")
	origins := make([]string, 0, len(parts))
	for _, p := range parts {
		if t := strings.TrimSpace(p); t != "" {
			origins = append(origins, t)
		}
	}
	return origins
}
