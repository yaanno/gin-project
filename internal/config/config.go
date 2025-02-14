// pkg/config/config.go
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"math/rand/v2"

	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
)

// Config represents the application configuration
type Config struct {
	// Server Configuration
	ServerAddress string
	ServerPort    string
	Environment   string
	ReadTimeout   time.Duration
	WriteTimeout  time.Duration
	IdleTimeout   time.Duration

	// Database Configuration
	DatabaseDriver string
	DatabasePath   string
	MaxOpenConns   int
	MaxIdleConns   int

	// JWT Configuration
	JWTSecret        string
	JWTRefreshSecret string
	AccessTokenTTL   time.Duration
	RefreshTokenTTL  time.Duration

	// Rate Limit Configuration
	RateLimitLimit    int
	RateLimitBurst    int64
	RateLimitDuration time.Duration

	// Logging Configuration
	LogLevel zerolog.Level
	LogPath  string

	// Security Configuration
	AllowedOrigins   []string
	MaxLoginAttempts int
	LockoutDuration  time.Duration
}

// DefaultConfig provides sensible default configuration values
func DefaultConfig() *Config {
	return &Config{
		// Server Defaults
		ServerAddress: "localhost",
		ServerPort:    "8080",
		Environment:   "development",
		ReadTimeout:   10 * time.Second,
		WriteTimeout:  10 * time.Second,
		IdleTimeout:   120 * time.Second,

		// Database Defaults
		DatabaseDriver: "sqlite3",
		DatabasePath:   "./data/users.db",
		MaxOpenConns:   25,
		MaxIdleConns:   25,

		// JWT Defaults
		AccessTokenTTL:  15 * time.Minute,
		RefreshTokenTTL: 7 * 24 * time.Hour,

		// Logging Defaults
		LogLevel: zerolog.InfoLevel,
		LogPath:  "./logs",

		// Security Defaults
		AllowedOrigins:   []string{"*"},
		MaxLoginAttempts: 5,
		LockoutDuration:  30 * time.Minute,
	}
}

// LoadConfig loads configuration from environment variables and .env file
func LoadConfig(envFiles ...string) (*Config, error) {
	// Start with default configuration
	cfg := DefaultConfig()

	// Load .env files if provided
	for _, envFile := range envFiles {
		if err := godotenv.Load(envFile); err != nil {
			// Only return error if a specific file is requested and not found
			if envFile != "" {
				return &Config{}, fmt.Errorf("error loading .env file: %w", err)
			}
		}
	}

	// Attempt to load default .env files if no specific file provided
	if len(envFiles) == 0 {
		possibleEnvFiles := []string{
			".env",
			filepath.Join(".config", ".env"),
			filepath.Join(os.Getenv("HOME"), ".config", "user-management-api", ".env"),
		}
		for _, file := range possibleEnvFiles {
			_ = godotenv.Load(file)
		}
	}

	// Override defaults with environment variables
	cfg = overrideConfigFromEnv(cfg)

	// Validate configuration
	if err := validateConfig(cfg); err != nil {
		return &Config{}, err
	}

	return cfg, nil
}

// overrideConfigFromEnv updates config with environment variable values
func overrideConfigFromEnv(cfg *Config) *Config {
	// Server Configuration
	if port := os.Getenv("SERVER_PORT"); port != "" {
		cfg.ServerPort = port
	}
	if address := os.Getenv("SERVER_ADDRESS"); address != "" {
		cfg.ServerAddress = address
	}
	cfg.Environment = getEnvOrDefault("ENVIRONMENT", cfg.Environment)

	// Database Configuration
	cfg.DatabaseDriver = getEnvOrDefault("DB_DRIVER", cfg.DatabaseDriver)
	cfg.DatabasePath = getEnvOrDefault("DB_PATH", cfg.DatabasePath)
	cfg.MaxOpenConns = getEnvIntOrDefault("DB_MAX_OPEN_CONNS", cfg.MaxOpenConns)
	cfg.MaxIdleConns = getEnvIntOrDefault("DB_MAX_IDLE_CONNS", cfg.MaxIdleConns)

	// JWT Configuration
	cfg.JWTSecret = getEnvOrDefault("JWT_SECRET", generateDefaultSecret(32))
	cfg.JWTRefreshSecret = getEnvOrDefault("JWT_REFRESH_SECRET", generateDefaultSecret(64))
	cfg.AccessTokenTTL = getEnvDurationOrDefault("ACCESS_TOKEN_TTL", cfg.AccessTokenTTL)
	cfg.RefreshTokenTTL = getEnvDurationOrDefault("REFRESH_TOKEN_TTL", cfg.RefreshTokenTTL)

	// Logging Configuration
	cfg.LogLevel = getEnvLogLevelOrDefault("LOG_LEVEL", cfg.LogLevel)
	cfg.LogPath = getEnvOrDefault("LOG_PATH", cfg.LogPath)

	// Security Configuration
	if origins := os.Getenv("ALLOWED_ORIGINS"); origins != "" {
		cfg.AllowedOrigins = strings.Split(origins, ",")
	}
	cfg.MaxLoginAttempts = getEnvIntOrDefault("MAX_LOGIN_ATTEMPTS", cfg.MaxLoginAttempts)
	cfg.LockoutDuration = getEnvDurationOrDefault("LOCKOUT_DURATION", cfg.LockoutDuration)

	return cfg
}

// validateConfig performs configuration validation
func validateConfig(cfg *Config) error {
	// Add validation rules
	if cfg.ServerPort == "" {
		return fmt.Errorf("server port cannot be empty")
	}

	if cfg.JWTSecret == "" || cfg.JWTRefreshSecret == "" {
		return fmt.Errorf("JWT secrets cannot be empty")
	}

	if cfg.AccessTokenTTL <= 0 || cfg.RefreshTokenTTL <= 0 {
		return fmt.Errorf("token TTLs must be positive")
	}

	return nil
}

// Helper functions for environment variable parsing
func getEnvOrDefault(key string, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvIntOrDefault(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

func getEnvDurationOrDefault(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}

func getEnvLogLevelOrDefault(key string, defaultValue zerolog.Level) zerolog.Level {
	if value := os.Getenv(key); value != "" {
		switch strings.ToLower(value) {
		case "debug":
			return zerolog.DebugLevel
		case "info":
			return zerolog.InfoLevel
		case "warn":
			return zerolog.WarnLevel
		case "error":
			return zerolog.ErrorLevel
		case "fatal":
			return zerolog.FatalLevel
		case "panic":
			return zerolog.PanicLevel
		}
	}
	return defaultValue
}

// generateDefaultSecret creates a secure random secret if not provided
func generateDefaultSecret(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*()_+"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.IntN(len(charset))]
	}
	return string(b)
}
