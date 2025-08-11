// Package config handles application configuration loading from environment variables.
package config

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config holds the application configuration settings.
type Config struct {
	DiscordToken    string
	CommandPrefix   string
	LogLevel        string
	JSONLogging     bool
	BotName         string
	ShutdownTimeout time.Duration
	RequestTimeout  time.Duration
	MaxRetries      int
	DebugMode       bool
	CacheTTL        time.Duration
	CacheSize       int
}

// Load loads configuration from environment variables.
func Load() (*Config, error) {
	cfg := &Config{
		CommandPrefix:   "!",    // default prefix.
		LogLevel:        "info", // default log level.
		JSONLogging:     false,  // default to text logging.
		BotName:         getEnv("BOT_NAME", "mtg-card-bot"),
		ShutdownTimeout: 30 * time.Second, // default shutdown timeout.
		RequestTimeout:  30 * time.Second, // default request timeout.
		MaxRetries:      3,                // default max retries.
		DebugMode:       false,            // default debug mode.
		CacheTTL:        1 * time.Hour,    // increased cache TTL for better performance.
		CacheSize:       1000,             // increased cache size for better hit rate.
	}

	// Discord token is required.
	cfg.DiscordToken = os.Getenv("DISCORD_TOKEN")
	if cfg.DiscordToken == "" {
		return nil, fmt.Errorf("DISCORD_TOKEN environment variable is required")
	}

	// Optional configurations.
	if prefix := os.Getenv("COMMAND_PREFIX"); prefix != "" {
		cfg.CommandPrefix = prefix
	}

	if logLevel := os.Getenv("LOG_LEVEL"); logLevel != "" {
		cfg.LogLevel = strings.ToLower(logLevel)
	}

	// Parse timeout configurations.
	if timeout := os.Getenv("SHUTDOWN_TIMEOUT"); timeout != "" {
		if parsed, err := time.ParseDuration(timeout); err == nil {
			cfg.ShutdownTimeout = parsed
		} else {
			log.Printf("Warning: invalid SHUTDOWN_TIMEOUT format '%s', using default", timeout)
		}
	}

	if timeout := os.Getenv("REQUEST_TIMEOUT"); timeout != "" {
		if parsed, err := time.ParseDuration(timeout); err == nil {
			cfg.RequestTimeout = parsed
		} else {
			log.Printf("Warning: invalid REQUEST_TIMEOUT format '%s', using default", timeout)
		}
	}

	// Parse retry configuration.
	cfg.MaxRetries = GetInt("MAX_RETRIES", cfg.MaxRetries)

	// Parse debug mode.
	cfg.DebugMode = GetBool("DEBUG", cfg.DebugMode)

	// Parse JSON logging.
	cfg.JSONLogging = GetBool("JSON_LOGGING", cfg.JSONLogging)

	// Parse cache configuration.
	if ttl := os.Getenv("CACHE_TTL"); ttl != "" {
		if parsed, err := time.ParseDuration(ttl); err == nil {
			cfg.CacheTTL = parsed
		} else {
			log.Printf("Warning: invalid CACHE_TTL format '%s', using default", ttl)
		}
	}

	cfg.CacheSize = GetInt("CACHE_SIZE", cfg.CacheSize)

	return cfg, nil
}

// Validate validates the configuration.
func (c *Config) Validate() error {
	if c.DiscordToken == "" {
		return fmt.Errorf("discord token is required")
	}

	if c.CommandPrefix == "" {
		return fmt.Errorf("command prefix cannot be empty")
	}

	validLogLevels := []string{"debug", "info", "warn", "error"}
	if !contains(validLogLevels, c.LogLevel) {
		return fmt.Errorf("invalid log level: %s (valid: %s)", c.LogLevel, strings.Join(validLogLevels, ", "))
	}

	if c.ShutdownTimeout <= 0 {
		return fmt.Errorf("shutdown timeout must be positive")
	}

	if c.RequestTimeout <= 0 {
		return fmt.Errorf("request timeout must be positive")
	}

	if c.MaxRetries < 0 {
		return fmt.Errorf("max retries cannot be negative")
	}

	if c.CacheTTL <= 0 {
		return fmt.Errorf("cache TTL must be positive")
	}

	if c.CacheSize <= 0 {
		return fmt.Errorf("cache size must be positive")
	}

	return nil
}

// GetBool returns a boolean environment variable with a default value.
func GetBool(key string, defaultValue bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}

	boolVal, err := strconv.ParseBool(value)
	if err != nil {
		return defaultValue
	}

	return boolVal
}

// GetInt returns an integer environment variable with a default value.
func GetInt(key string, defaultValue int) int {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}

	intVal, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}

	return intVal
}

// getEnv returns an environment variable with a default value.
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return defaultValue
}

// contains checks if a slice contains a string.
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}

	return false
}
