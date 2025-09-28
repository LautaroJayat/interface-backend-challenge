package testutils

import (
	"testing"

	"messaging-app/internal/ports"
)

// TestLogger is a minimal logger implementation for integration testing.
type TestLogger struct {
	T *testing.T
}

func NewTestLogger(t *testing.T) *TestLogger {
	return &TestLogger{T: t}
}

func (l *TestLogger) Debug(msg string, args ...any) {
	l.T.Logf("[DEBUG] %s %v", msg, args)
}

func (l *TestLogger) Info(msg string, args ...any) {
	l.T.Logf("[INFO] %s %v", msg, args)
}

func (l *TestLogger) Warn(msg string, args ...any) {
	l.T.Logf("[WARN] %s %v", msg, args)
}

func (l *TestLogger) Error(msg string, args ...any) {
	l.T.Logf("[ERROR] %s %v", msg, args)
}

func (l *TestLogger) With(args ...any) ports.Logger {
	// Just return itself; ignoring args is fine for tests
	return l
}
