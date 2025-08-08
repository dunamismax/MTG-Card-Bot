// Package errors provides custom error types and utilities for the MTG Card Bot.
package errors

import (
	"errors"
	"fmt"
	"net/http"
)

// ErrorType represents the category of error that occurred.
type ErrorType string

const (
	// ErrorTypeAPI represents an API-related error.
	ErrorTypeAPI ErrorType = "api_error"
	// ErrorTypeConfig represents a configuration error.
	ErrorTypeConfig ErrorType = "config_error"
	// ErrorTypeDiscord represents a Discord-related error.
	ErrorTypeDiscord ErrorType = "discord_error"
	// ErrorTypeValidation represents a validation error.
	ErrorTypeValidation ErrorType = "validation_error"
	// ErrorTypeNotFound represents a not found error.
	ErrorTypeNotFound ErrorType = "not_found_error"
	// ErrorTypeRateLimit represents a rate limit error.
	ErrorTypeRateLimit ErrorType = "rate_limit_error"
	// ErrorTypeNetwork represents a network error.
	ErrorTypeNetwork ErrorType = "network_error"
	// ErrorTypeInternal represents an internal error.
	ErrorTypeInternal ErrorType = "internal_error"
	// ErrorTypeCache represents a cache error.
	ErrorTypeCache ErrorType = "cache_error"
)

// MTGError represents a categorized error with additional context.
type MTGError struct {
	Type       ErrorType
	Message    string
	Cause      error
	StatusCode int
	Context    map[string]interface{}
}

func (e *MTGError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s (caused by: %v)", e.Type, e.Message, e.Cause)
	}

	return fmt.Sprintf("%s: %s", e.Type, e.Message)
}

func (e *MTGError) Unwrap() error {
	return e.Cause
}

// NewAPIError creates a new API-related error.
func NewAPIError(message string, cause error) *MTGError {
	return &MTGError{
		Type:    ErrorTypeAPI,
		Message: message,
		Cause:   cause,
	}
}

// NewConfigError creates a new configuration error.
func NewConfigError(message string, cause error) *MTGError {
	return &MTGError{
		Type:    ErrorTypeConfig,
		Message: message,
		Cause:   cause,
	}
}

// NewDiscordError creates a new Discord-related error.
func NewDiscordError(message string, cause error) *MTGError {
	return &MTGError{
		Type:    ErrorTypeDiscord,
		Message: message,
		Cause:   cause,
	}
}

// NewValidationError creates a new validation error.
func NewValidationError(message string) *MTGError {
	return &MTGError{
		Type:    ErrorTypeValidation,
		Message: message,
	}
}

// NewNotFoundError creates a new not found error.
func NewNotFoundError(message string) *MTGError {
	return &MTGError{
		Type:    ErrorTypeNotFound,
		Message: message,
	}
}

// NewRateLimitError creates a new rate limit error.
func NewRateLimitError(message string, retryAfter int) *MTGError {
	return &MTGError{
		Type:    ErrorTypeRateLimit,
		Message: message,
		Context: map[string]interface{}{
			"retry_after": retryAfter,
		},
	}
}

// NewNetworkError creates a new network error.
func NewNetworkError(message string, cause error) *MTGError {
	return &MTGError{
		Type:    ErrorTypeNetwork,
		Message: message,
		Cause:   cause,
	}
}

// NewInternalError creates a new internal error.
func NewInternalError(message string, cause error) *MTGError {
	return &MTGError{
		Type:    ErrorTypeInternal,
		Message: message,
		Cause:   cause,
	}
}

// NewCacheError creates a new cache-related error.
func NewCacheError(message string, cause error) *MTGError {
	return &MTGError{
		Type:    ErrorTypeCache,
		Message: message,
		Cause:   cause,
	}
}

// IsErrorType checks if an error is of a specific type.
func IsErrorType(err error, errorType ErrorType) bool {
	var mtgErr *MTGError
	if errors.As(err, &mtgErr) {
		return mtgErr.Type == errorType
	}

	return false
}

// FromHTTPStatus creates an appropriate error based on HTTP status code.
func FromHTTPStatus(statusCode int, message string) *MTGError {
	switch {
	case statusCode == http.StatusNotFound:
		return NewNotFoundError(message)
	case statusCode == http.StatusTooManyRequests:
		return NewRateLimitError(message, 0)
	case statusCode >= 400 && statusCode < 500:
		return NewValidationError(message)
	case statusCode >= 500:
		return NewAPIError(message, nil)
	default:
		return NewInternalError(message, nil)
	}
}
