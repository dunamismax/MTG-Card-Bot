package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	DiscordToken string
	CommandPrefix string
	LogLevel     string
	BotName      string
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	cfg := &Config{
		CommandPrefix: "!",  // default prefix
		LogLevel:      "info", // default log level
		BotName:       getEnv("BOT_NAME", "unknown"),
	}
	
	// Discord token is required
	cfg.DiscordToken = os.Getenv("DISCORD_TOKEN")
	if cfg.DiscordToken == "" {
		return nil, fmt.Errorf("DISCORD_TOKEN environment variable is required")
	}
	
	// Optional configurations
	if prefix := os.Getenv("COMMAND_PREFIX"); prefix != "" {
		cfg.CommandPrefix = prefix
	}
	
	if logLevel := os.Getenv("LOG_LEVEL"); logLevel != "" {
		cfg.LogLevel = strings.ToLower(logLevel)
	}
	
	return cfg, nil
}

// Validate validates the configuration
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
	
	return nil
}

// GetBool returns a boolean environment variable with a default value
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

// GetInt returns an integer environment variable with a default value
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

// getEnv returns an environment variable with a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// contains checks if a slice contains a string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}