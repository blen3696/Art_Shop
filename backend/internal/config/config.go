package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

// Config holds all application configuration loaded from environment variables.
type Config struct {
	// Server
	Port           string
	Env            string // "development", "production", "staging"
	AllowedOrigins []string

	// Database (Supabase PostgreSQL)
	DatabaseURL    string
	DBMaxOpenConns int
	DBMaxIdleConns int

	// JWT Authentication
	JWTSecret             string
	JWTExpiryHours        int
	JWTRefreshExpiryHours int

	// Supabase
	SupabaseURL        string
	SupabaseAnonKey    string
	SupabaseServiceKey string
	StorageBucket      string

	// AI (Google Gemini). Empty GeminiAPIKey disables generative features —
	// the /api/ai/recommendations endpoint still works (SQL-based) but the
	// description/tags generators return 503.
	GeminiAPIKey string
	AIModel      string

	// Email (Brevo). EmailFromAddress must be a sender verified in Brevo
	// (Senders & IP → Senders), or a verified domain.
	BrevoAPIKey      string
	EmailFromName    string
	EmailFromAddress string

	// AppURL is the public frontend URL — used to build links in transactional
	// emails (password reset, etc.). No trailing slash.
	AppURL string

	// Rate Limiting
	RateLimitRPS   float64
	RateLimitBurst int
}

// Load reads configuration from the .env file (if present) and environment
// variables. It returns a fully populated Config or an error if any required
// value is missing.
func Load() (*Config, error) {
	// Attempt to load .env; ignore error if the file doesn't exist (e.g. in
	// production where env vars are injected directly).
	_ = godotenv.Load()

	cfg := &Config{
		Port:           getEnv("PORT", "8080"),
		Env:            getEnv("ENV", "development"),
		AllowedOrigins: parseCSV(getEnv("ALLOWED_ORIGINS", "http://localhost:3000,http://localhost:5173")),

		DatabaseURL:    getEnv("DATABASE_URL", ""),
		DBMaxOpenConns: getEnvInt("DB_MAX_OPEN_CONNS", 25),
		DBMaxIdleConns: getEnvInt("DB_MAX_IDLE_CONNS", 10),

		JWTSecret:             getEnv("JWT_SECRET", ""),
		JWTExpiryHours:        getEnvInt("JWT_EXPIRY_HOURS", 24),
		JWTRefreshExpiryHours: getEnvInt("JWT_REFRESH_EXPIRY_HOURS", 168), // 7 days

		SupabaseURL:        getEnv("SUPABASE_URL", ""),
		SupabaseAnonKey:    getEnv("SUPABASE_ANON_KEY", ""),
		SupabaseServiceKey: getEnv("SUPABASE_SERVICE_KEY", ""),
		StorageBucket:      getEnv("STORAGE_BUCKET", "artshop-images"),

		GeminiAPIKey: getEnv("GEMINI_API_KEY", ""),
		AIModel:      getEnv("AI_MODEL", "gemini-2.0-flash"),

		BrevoAPIKey:      getEnv("BREVO_API_KEY", ""),
		EmailFromName:    getEnv("EMAIL_FROM_NAME", "ArtShop"),
		EmailFromAddress: getEnv("EMAIL_FROM_ADDRESS", ""),

		AppURL: getEnv("APP_URL", "http://localhost:5173"),

		RateLimitRPS:   getEnvFloat("RATE_LIMIT_RPS", 10),
		RateLimitBurst: getEnvInt("RATE_LIMIT_BURST", 30),
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// IsDevelopment returns true when running in development mode.
func (c *Config) IsDevelopment() bool {
	return c.Env == "development"
}

// IsProduction returns true when running in production mode.
func (c *Config) IsProduction() bool {
	return c.Env == "production"
}

// validate ensures that critical configuration values are set.
func (c *Config) validate() error {
	if c.DatabaseURL == "" {
		return fmt.Errorf("config: DATABASE_URL is required")
	}
	if c.JWTSecret == "" {
		return fmt.Errorf("config: JWT_SECRET is required")
	}
	return nil
}

// getEnv reads an environment variable or returns a fallback default.
func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

// getEnvInt reads an environment variable as an integer or returns a fallback.
func getEnvInt(key string, fallback int) int {
	raw := os.Getenv(key)
	if raw == "" {
		return fallback
	}
	val, err := strconv.Atoi(raw)
	if err != nil {
		return fallback
	}
	return val
}

// getEnvFloat reads an environment variable as a float64 or returns a fallback.
func getEnvFloat(key string, fallback float64) float64 {
	raw := os.Getenv(key)
	if raw == "" {
		return fallback
	}
	val, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		return fallback
	}
	return val
}

// parseCSV splits a comma-separated string into a trimmed slice.
func parseCSV(s string) []string {
	parts := strings.Split(s, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		trimmed := strings.TrimSpace(p)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}
