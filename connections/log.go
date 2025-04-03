package connections

import (
	"fmt"
	"log"
)

// Logger is a simple interface for logging
type Logger interface {
	Info(format string, args ...interface{})
	Error(format string, args ...interface{})
}

// DefaultLogger is a basic logger implementation using the standard log package
type DefaultLogger struct{}

// Info logs an informational message
func (l *DefaultLogger) Info(format string, args ...interface{}) {
	log.Printf("[INFO] %s", fmt.Sprintf(format, args...))
}

// Error logs an error message
func (l *DefaultLogger) Error(format string, args ...interface{}) {
	log.Printf("[ERROR] %s", fmt.Sprintf(format, args...))
}

// NewDefaultLogger creates a new default logger
func NewDefaultLogger() Logger {
	return &DefaultLogger{}
}

// NoopLogger is a logger that discards all messages
type NoopLogger struct{}

// Info is a no-op for NoopLogger
func (l *NoopLogger) Info(format string, args ...interface{}) {
	// Do nothing
}

// Error is a no-op for NoopLogger
func (l *NoopLogger) Error(format string, args ...interface{}) {
	// Do nothing
}

// NewNoopLogger creates a new noop logger
func NewNoopLogger() Logger {
	return &NoopLogger{}
}
