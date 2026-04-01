package config

import (
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"time"
)

// Config holds all runtime configuration for the service.
// All values are loaded from environment variables at startup.
// The service will exit if any required value is missing.
type Config struct {
	GRPCPort      int
	MetricsPort   int
	AnthropicKey  string
	APIKey        string // shared secret for bearer auth
	LogLevel      string
	OFFBaseURL    string
	ClaudeModel   string
	OFFTimeout    time.Duration
	ClaudeTimeout time.Duration
}

// Load reads configuration from environment variables and validates required fields.
// Returns an error if any required variable is missing or unparseable.
func Load() (*Config, error) {
	cfg := &Config{
		GRPCPort:      envInt("GRPC_PORT", 50051),
		MetricsPort:   envInt("METRICS_PORT", 9090),
		AnthropicKey:  os.Getenv("ANTHROPIC_API_KEY"),
		APIKey:        os.Getenv("ECOSCAN_API_KEY"),
		LogLevel:      envStr("LOG_LEVEL", "info"),
		OFFBaseURL:    envStr("OFF_BASE_URL", "https://world.openfoodfacts.org"),
		ClaudeModel:   envStr("CLAUDE_MODEL", "claude-sonnet-4-6"),
		OFFTimeout:    envDuration("OFF_TIMEOUT", 5*time.Second),
		ClaudeTimeout: envDuration("CLAUDE_TIMEOUT", 15*time.Second),
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (c *Config) validate() error {
	if c.AnthropicKey == "" {
		return fmt.Errorf("ANTHROPIC_API_KEY is required")
	}
	if c.APIKey == "" {
		return fmt.Errorf("ECOSCAN_API_KEY is required")
	}
	return nil
}

func (c *Config) SlogLevel() slog.Level {
	switch c.LogLevel {
	case "debug":
		return slog.LevelDebug
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

func envStr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func envInt(key string, fallback int) int {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return n
}

func envDuration(key string, fallback time.Duration) time.Duration {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return fallback
	}
	return d
}
