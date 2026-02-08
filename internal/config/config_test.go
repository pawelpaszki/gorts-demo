package config

import (
	"os"
	"testing"
	"time"
)

func clearEnv() {
	envVars := []string{
		"SERVER_HOST", "SERVER_PORT", "SERVER_READ_TIMEOUT",
		"SERVER_WRITE_TIMEOUT", "SERVER_IDLE_TIMEOUT",
		"DB_DRIVER", "DB_DSN", "DB_MAX_CONNS", "DB_MAX_IDLE",
		"AUTH_ENABLED", "AUTH_REALM", "AUTH_TOKEN_EXPIRY",
		"FEATURE_READING_LISTS", "FEATURE_SEARCH", "FEATURE_METRICS",
	}
	for _, v := range envVars {
		os.Unsetenv(v)
	}
}

func TestLoad_Defaults(t *testing.T) {
	clearEnv()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Server defaults
	if cfg.Server.Host != "0.0.0.0" {
		t.Errorf("Server.Host = %s, want 0.0.0.0", cfg.Server.Host)
	}
	if cfg.Server.Port != 8080 {
		t.Errorf("Server.Port = %d, want 8080", cfg.Server.Port)
	}

	// Database defaults
	if cfg.Database.Driver != "sqlite" {
		t.Errorf("Database.Driver = %s, want sqlite", cfg.Database.Driver)
	}
	if cfg.Database.MaxConns != 10 {
		t.Errorf("Database.MaxConns = %d, want 10", cfg.Database.MaxConns)
	}

	// Auth defaults
	if cfg.Auth.Enabled != false {
		t.Error("Auth.Enabled should be false by default")
	}

	// Feature defaults
	if cfg.Features.EnableReadingLists != true {
		t.Error("Features.EnableReadingLists should be true by default")
	}
}

func TestLoad_FromEnv(t *testing.T) {
	clearEnv()

	os.Setenv("SERVER_HOST", "localhost")
	os.Setenv("SERVER_PORT", "3000")
	os.Setenv("DB_DRIVER", "postgres")
	os.Setenv("DB_DSN", "postgres://localhost/test")
	os.Setenv("AUTH_ENABLED", "true")
	os.Setenv("FEATURE_SEARCH", "true")

	defer clearEnv()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Server.Host != "localhost" {
		t.Errorf("Server.Host = %s, want localhost", cfg.Server.Host)
	}
	if cfg.Server.Port != 3000 {
		t.Errorf("Server.Port = %d, want 3000", cfg.Server.Port)
	}
	if cfg.Database.Driver != "postgres" {
		t.Errorf("Database.Driver = %s, want postgres", cfg.Database.Driver)
	}
	if cfg.Auth.Enabled != true {
		t.Error("Auth.Enabled should be true")
	}
	if cfg.Features.EnableSearch != true {
		t.Error("Features.EnableSearch should be true")
	}
}

func TestLoad_Duration(t *testing.T) {
	clearEnv()

	os.Setenv("SERVER_READ_TIMEOUT", "30s")
	os.Setenv("AUTH_TOKEN_EXPIRY", "48h")

	defer clearEnv()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Server.ReadTimeout != 30*time.Second {
		t.Errorf("Server.ReadTimeout = %v, want 30s", cfg.Server.ReadTimeout)
	}
	if cfg.Auth.TokenExpiry != 48*time.Hour {
		t.Errorf("Auth.TokenExpiry = %v, want 48h", cfg.Auth.TokenExpiry)
	}
}

func TestConfig_Validate_InvalidPort(t *testing.T) {
	cfg := &Config{
		Server: ServerConfig{Port: 0},
		Database: DatabaseConfig{
			MaxConns: 10,
			MaxIdle:  5,
		},
	}

	err := cfg.Validate()
	if err == nil {
		t.Error("Expected validation error for invalid port")
	}
}

func TestConfig_Validate_InvalidMaxConns(t *testing.T) {
	cfg := &Config{
		Server: ServerConfig{Port: 8080},
		Database: DatabaseConfig{
			MaxConns: 0,
			MaxIdle:  0,
		},
	}

	err := cfg.Validate()
	if err == nil {
		t.Error("Expected validation error for invalid max connections")
	}
}

func TestConfig_Validate_MaxIdleExceedsMaxConns(t *testing.T) {
	cfg := &Config{
		Server: ServerConfig{Port: 8080},
		Database: DatabaseConfig{
			MaxConns: 5,
			MaxIdle:  10,
		},
	}

	err := cfg.Validate()
	if err == nil {
		t.Error("Expected validation error when max idle exceeds max connections")
	}
}

func TestConfig_Address(t *testing.T) {
	cfg := &Config{
		Server: ServerConfig{
			Host: "localhost",
			Port: 3000,
		},
	}

	addr := cfg.Address()
	expected := "localhost:3000"
	if addr != expected {
		t.Errorf("Address() = %s, want %s", addr, expected)
	}
}

func TestGetEnvBool(t *testing.T) {
	tests := []struct {
		value    string
		expected bool
	}{
		{"true", true},
		{"TRUE", true},
		{"True", true},
		{"1", true},
		{"yes", true},
		{"YES", true},
		{"on", true},
		{"false", false},
		{"0", false},
		{"no", false},
		{"off", false},
		{"invalid", false},
	}

	for _, tt := range tests {
		t.Run(tt.value, func(t *testing.T) {
			os.Setenv("TEST_BOOL", tt.value)
			defer os.Unsetenv("TEST_BOOL")

			result := getEnvBool("TEST_BOOL", false)
			if result != tt.expected {
				t.Errorf("getEnvBool(%q) = %v, want %v", tt.value, result, tt.expected)
			}
		})
	}
}

func TestMustLoad_Panic(t *testing.T) {
	clearEnv()
	os.Setenv("SERVER_PORT", "invalid")
	defer clearEnv()

	defer func() {
		if r := recover(); r == nil {
			t.Error("MustLoad should panic on invalid config")
		}
	}()

	// This should use invalid port from env, but getEnvInt returns default on parse error
	// So let's set an actually invalid value
	os.Setenv("SERVER_PORT", "-1")
	MustLoad()
}
