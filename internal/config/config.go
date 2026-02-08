package config

import (
	"errors"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config holds the application configuration.
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Auth     AuthConfig
	Features FeatureFlags
}

// ServerConfig holds server-related configuration.
type ServerConfig struct {
	Host         string
	Port         int
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

// DatabaseConfig holds database-related configuration.
type DatabaseConfig struct {
	Driver   string
	DSN      string
	MaxConns int
	MaxIdle  int
}

// AuthConfig holds authentication configuration.
type AuthConfig struct {
	Enabled     bool
	Realm       string
	TokenExpiry time.Duration
}

// FeatureFlags holds feature toggle configuration.
type FeatureFlags struct {
	EnableReadingLists bool
	EnableSearch       bool
	EnableMetrics      bool
}

// Load reads configuration from environment variables.
func Load() (*Config, error) {
	cfg := &Config{
		Server: ServerConfig{
			Host:         getEnv("SERVER_HOST", "0.0.0.0"),
			Port:         getEnvInt("SERVER_PORT", 8080),
			ReadTimeout:  getEnvDuration("SERVER_READ_TIMEOUT", 15*time.Second),
			WriteTimeout: getEnvDuration("SERVER_WRITE_TIMEOUT", 15*time.Second),
			IdleTimeout:  getEnvDuration("SERVER_IDLE_TIMEOUT", 60*time.Second),
		},
		Database: DatabaseConfig{
			Driver:   getEnv("DB_DRIVER", "sqlite"),
			DSN:      getEnv("DB_DSN", "bookshelf.db"),
			MaxConns: getEnvInt("DB_MAX_CONNS", 10),
			MaxIdle:  getEnvInt("DB_MAX_IDLE", 5),
		},
		Auth: AuthConfig{
			Enabled:     getEnvBool("AUTH_ENABLED", false),
			Realm:       getEnv("AUTH_REALM", "Bookshelf API"),
			TokenExpiry: getEnvDuration("AUTH_TOKEN_EXPIRY", 24*time.Hour),
		},
		Features: FeatureFlags{
			EnableReadingLists: getEnvBool("FEATURE_READING_LISTS", true),
			EnableSearch:       getEnvBool("FEATURE_SEARCH", false),
			EnableMetrics:      getEnvBool("FEATURE_METRICS", false),
		},
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// Validate checks if the configuration is valid.
func (c *Config) Validate() error {
	if c.Server.Port < 1 || c.Server.Port > 65535 {
		return errors.New("invalid server port")
	}
	if c.Database.MaxConns < 1 {
		return errors.New("database max connections must be at least 1")
	}
	if c.Database.MaxIdle < 0 {
		return errors.New("database max idle connections cannot be negative")
	}
	if c.Database.MaxIdle > c.Database.MaxConns {
		return errors.New("database max idle cannot exceed max connections")
	}
	return nil
}

// Address returns the server address in host:port format.
func (c *Config) Address() string {
	return c.Server.Host + ":" + strconv.Itoa(c.Server.Port)
}

// getEnv returns the value of an environment variable or a default.
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvInt returns an integer environment variable or a default.
func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

// getEnvBool returns a boolean environment variable or a default.
func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		lower := strings.ToLower(value)
		return lower == "true" || lower == "1" || lower == "yes" || lower == "on"
	}
	return defaultValue
}

// getEnvDuration returns a duration environment variable or a default.
func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}

// MustLoad loads configuration and panics on error.
func MustLoad() *Config {
	cfg, err := Load()
	if err != nil {
		panic("failed to load config: " + err.Error())
	}
	return cfg
}
