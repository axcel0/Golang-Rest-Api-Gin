// Package logger provides structured logging using Go's standard library slog
// with support for JSON and text formats, configurable log levels, and
// contextual logging with structured fields.
package logger

import (
	"context"
	"log/slog"
	"os"
)

var logger *slog.Logger

// Init initializes the global logger
func Init(level, format string) {
	var handler slog.Handler

	// Set log level
	var logLevel slog.Level
	switch level {
	case "debug":
		logLevel = slog.LevelDebug
	case "info":
		logLevel = slog.LevelInfo
	case "warn":
		logLevel = slog.LevelWarn
	case "error":
		logLevel = slog.LevelError
	default:
		logLevel = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{
		Level: logLevel,
	}

	// Set format (JSON or console)
	if format == "json" {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		handler = slog.NewTextHandler(os.Stdout, opts)
	}

	logger = slog.New(handler)
	slog.SetDefault(logger)
}

// Get returns the global logger instance
func Get() *slog.Logger {
	if logger == nil {
		// Fallback to default logger if not initialized
		logger = slog.Default()
	}
	return logger
}

// Debug logs a debug message with attributes
func Debug(msg string, args ...any) {
	Get().Debug(msg, args...)
}

// Info logs an info message with attributes
func Info(msg string, args ...any) {
	Get().Info(msg, args...)
}

// Warn logs a warning message with attributes
func Warn(msg string, args ...any) {
	Get().Warn(msg, args...)
}

// Error logs an error message with attributes
func Error(msg string, args ...any) {
	Get().Error(msg, args...)
}

// With returns a logger with the given attributes
func With(args ...any) *slog.Logger {
	return Get().With(args...)
}

// WithContext returns a logger with context
func WithContext(ctx context.Context) *slog.Logger {
	return Get().With(slog.Any("context", ctx))
}
