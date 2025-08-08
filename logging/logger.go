package logging

import (
	"context"
	"log/slog"
	"os"
	"strings"

	"github.com/dunamismax/MTG-Card-Bot/errors"
)

var DefaultLogger *slog.Logger

// LogLevel represents the logging level
type LogLevel string

const (
	LevelDebug LogLevel = "debug"
	LevelInfo  LogLevel = "info"
	LevelWarn  LogLevel = "warn"
	LevelError LogLevel = "error"
)

// InitializeLogger initializes the global logger with the specified level and format
func InitializeLogger(level string, jsonFormat bool) {
	var slogLevel slog.Level

	switch strings.ToLower(level) {
	case "debug":
		slogLevel = slog.LevelDebug
	case "info":
		slogLevel = slog.LevelInfo
	case "warn", "warning":
		slogLevel = slog.LevelWarn
	case "error":
		slogLevel = slog.LevelError
	default:
		slogLevel = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{
		Level: slogLevel,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			// Customize timestamp format
			if a.Key == slog.TimeKey {
				a.Value = slog.StringValue(a.Value.Time().Format("2006-01-02T15:04:05.000Z07:00"))
			}
			return a
		},
	}

	var handler slog.Handler
	if jsonFormat {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		handler = slog.NewTextHandler(os.Stdout, opts)
	}

	DefaultLogger = slog.New(handler)
	slog.SetDefault(DefaultLogger)
}

// WithContext returns a logger with context values
func WithContext(ctx context.Context) *slog.Logger {
	return DefaultLogger.With()
}

// WithComponent returns a logger with a component field
func WithComponent(component string) *slog.Logger {
	return DefaultLogger.With("component", component)
}

// WithUser returns a logger with user information
func WithUser(userID, username string) *slog.Logger {
	return DefaultLogger.With("user_id", userID, "username", username)
}

// WithCommand returns a logger with command information
func WithCommand(command string) *slog.Logger {
	return DefaultLogger.With("command", command)
}

// WithCard returns a logger with card information
func WithCard(cardName string) *slog.Logger {
	return DefaultLogger.With("card_name", cardName)
}

// LogError logs an MTGError with appropriate structured fields
func LogError(logger *slog.Logger, err error, message string) {
	if mtgErr, ok := err.(*errors.MTGError); ok {
		attrs := []slog.Attr{
			slog.String("error_type", string(mtgErr.Type)),
			slog.String("error_message", mtgErr.Message),
		}

		if mtgErr.StatusCode != 0 {
			attrs = append(attrs, slog.Int("status_code", mtgErr.StatusCode))
		}

		if mtgErr.Cause != nil {
			attrs = append(attrs, slog.String("cause", mtgErr.Cause.Error()))
		}

		for key, value := range mtgErr.Context {
			attrs = append(attrs, slog.Any(key, value))
		}

		logger.LogAttrs(context.Background(), slog.LevelError, message, attrs...)
	} else {
		logger.Error(message, "error", err)
	}
}

// Debug logs a debug message with optional attributes
func Debug(msg string, args ...any) {
	DefaultLogger.Debug(msg, args...)
}

// Info logs an info message with optional attributes
func Info(msg string, args ...any) {
	DefaultLogger.Info(msg, args...)
}

// Warn logs a warning message with optional attributes
func Warn(msg string, args ...any) {
	DefaultLogger.Warn(msg, args...)
}

// Error logs an error message with optional attributes
func Error(msg string, args ...any) {
	DefaultLogger.Error(msg, args...)
}

// DebugWithContext logs a debug message with context
func DebugWithContext(ctx context.Context, msg string, args ...any) {
	DefaultLogger.DebugContext(ctx, msg, args...)
}

// InfoWithContext logs an info message with context
func InfoWithContext(ctx context.Context, msg string, args ...any) {
	DefaultLogger.InfoContext(ctx, msg, args...)
}

// WarnWithContext logs a warning message with context
func WarnWithContext(ctx context.Context, msg string, args ...any) {
	DefaultLogger.WarnContext(ctx, msg, args...)
}

// ErrorWithContext logs an error message with context
func ErrorWithContext(ctx context.Context, msg string, args ...any) {
	DefaultLogger.ErrorContext(ctx, msg, args...)
}

// LogStartup logs application startup information
func LogStartup(botName, prefix, logLevel string, debugMode bool) {
	logger := WithComponent("startup")
	logger.Info("Starting MTG Card Bot",
		"bot_name", botName,
		"command_prefix", prefix,
		"log_level", logLevel,
		"debug_mode", debugMode,
	)
}

// LogShutdown logs application shutdown information
func LogShutdown() {
	logger := WithComponent("shutdown")
	logger.Info("Bot shutdown complete")
}

// LogAPIRequest logs API request information
func LogAPIRequest(endpoint string, duration int64) {
	logger := WithComponent("scryfall")
	logger.Debug("API request completed",
		"endpoint", endpoint,
		"duration_ms", duration,
	)
}

// LogDiscordCommand logs Discord command execution
func LogDiscordCommand(userID, username, command string, success bool) {
	logger := WithComponent("discord").With(
		"user_id", userID,
		"username", username,
		"command", command,
		"success", success,
	)

	if success {
		logger.Info("Command executed successfully")
	} else {
		logger.Warn("Command execution failed")
	}
}

// LogCacheOperation logs cache operations
func LogCacheOperation(operation, key string, hit bool, duration int64) {
	logger := WithComponent("cache")
	logger.Debug("Cache operation",
		"operation", operation,
		"key", key,
		"hit", hit,
		"duration_ns", duration,
	)
}
