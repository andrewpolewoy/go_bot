package logger

import (
	"log"
	"log/slog"
	"os"
	"strings"
)

// Logger is a simple logging interface
type Logger interface {
	Info(msg string, kv ...any)
	Error(msg string, kv ...any)
	Debug(msg string, kv ...any)
}

// Level represents log level
type Level string

const (
	LevelDebug Level = "debug"
	LevelInfo  Level = "info"
	LevelWarn  Level = "warn"
	LevelError Level = "error"
)

// StdLogger is a standard logger implementation
type StdLogger struct {
	logger *slog.Logger
}

// New creates a new logger with the specified level
func New(level Level) Logger {
	var slogLevel slog.Level
	switch strings.ToLower(string(level)) {
	case "debug":
		slogLevel = slog.LevelDebug
	case "info":
		slogLevel = slog.LevelInfo
	case "warn":
		slogLevel = slog.LevelWarn
	case "error":
		slogLevel = slog.LevelError
	default:
		slogLevel = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{
		Level: slogLevel,
	}

	handler := slog.NewTextHandler(os.Stdout, opts)
	logger := slog.New(handler)

	return &StdLogger{logger: logger}
}

// NewDefault creates a new logger with info level
func NewDefault() Logger {
	return New(LevelInfo)
}

// Info logs an info message
func (l *StdLogger) Info(msg string, kv ...any) {
	if len(kv) > 0 {
		l.logger.Info(msg, kv...)
	} else {
		l.logger.Info(msg)
	}
}

// Error logs an error message
func (l *StdLogger) Error(msg string, kv ...any) {
	if len(kv) > 0 {
		l.logger.Error(msg, kv...)
	} else {
		l.logger.Error(msg)
	}
}

// Debug logs a debug message
func (l *StdLogger) Debug(msg string, kv ...any) {
	if len(kv) > 0 {
		l.logger.Debug(msg, kv...)
	} else {
		l.logger.Debug(msg)
	}
}

// ToSlogLogger returns the underlying slog.Logger for compatibility
func (l *StdLogger) ToSlogLogger() *slog.Logger {
	return l.logger
}

// LegacyLogger wraps the old log.Logger for backward compatibility
type LegacyLogger struct {
	*log.Logger
}

// NewLegacy creates a logger from standard log.Logger
func NewLegacy(l *log.Logger) Logger {
	return &LegacyLogger{Logger: l}
}

// Info logs an info message
func (l *LegacyLogger) Info(msg string, kv ...any) {
	args := append([]any{"INFO:", msg}, kv...)
	l.Println(args...)
}

// Error logs an error message
func (l *LegacyLogger) Error(msg string, kv ...any) {
	args := append([]any{"ERROR:", msg}, kv...)
	l.Println(args...)
}

// Debug logs a debug message
func (l *LegacyLogger) Debug(msg string, kv ...any) {
	args := append([]any{"DEBUG:", msg}, kv...)
	l.Println(args...)
}