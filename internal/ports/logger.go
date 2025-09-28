package ports

import "log/slog"

//go:generate mockery --name=Logger --output=../mocks --outpkg=mocks

type Logger interface {
	Debug(msg string, args ...any)
	Info(msg string, args ...any)
	Warn(msg string, args ...any)
	Error(msg string, args ...any)
	With(args ...any) Logger
}

// SlogAdapter wraps slog.Logger to implement our Logger interface
type SlogAdapter struct {
	logger *slog.Logger
}

func NewSlogAdapter(logger *slog.Logger) Logger {
	return &SlogAdapter{logger: logger}
}

func (l *SlogAdapter) Debug(msg string, args ...any) {
	l.logger.Debug(msg, args...)
}

func (l *SlogAdapter) Info(msg string, args ...any) {
	l.logger.Info(msg, args...)
}

func (l *SlogAdapter) Warn(msg string, args ...any) {
	l.logger.Warn(msg, args...)
}

func (l *SlogAdapter) Error(msg string, args ...any) {
	l.logger.Error(msg, args...)
}

func (l *SlogAdapter) With(args ...any) Logger {
	return &SlogAdapter{logger: l.logger.With(args...)}
}