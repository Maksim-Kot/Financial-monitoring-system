package logger

import (
	"context"
	"io"
	"log/slog"
	"os"
)

type slogLogger struct {
	log *slog.Logger
}

func newSlog(l *slog.Logger) Logger {
	if l == nil {
		l = slog.Default()
	}
	return &slogLogger{log: l}
}

func newJSON(w io.Writer, level LogLevel) Logger {
	h := slog.NewJSONHandler(w, &slog.HandlerOptions{
		Level: slog.Level(level),
	})
	return newSlog(slog.New(h))
}

func newText(w io.Writer, level LogLevel) Logger {
	h := slog.NewTextHandler(w, &slog.HandlerOptions{
		Level: slog.Level(level),
	})
	return newSlog(slog.New(h))
}

func NewStdoutJSON(level string) Logger {
	return newJSON(os.Stdout, parseLogLevel(level))
}

func NewStdoutText(level string) Logger {
	return newText(os.Stdout, parseLogLevel(level))
}

func (s *slogLogger) With(kv ...any) Logger {
	return &slogLogger{log: s.log.With(kv...)}
}

func (s *slogLogger) WithGroup(name string) Logger {
	return &slogLogger{log: s.log.WithGroup(name)}
}

func (s *slogLogger) Debug(msg string, args ...any) {
	s.log.Debug(msg, args...)
}

func (s *slogLogger) DebugContext(ctx context.Context, msg string, args ...any) {
	s.log.DebugContext(ctx, msg, args...)
}

func (s *slogLogger) Info(msg string, args ...any) {
	s.log.Info(msg, args...)
}

func (s *slogLogger) InfoContext(ctx context.Context, msg string, args ...any) {
	s.log.InfoContext(ctx, msg, args...)
}

func (s *slogLogger) Warn(msg string, args ...any) {
	s.log.Warn(msg, args...)
}

func (s *slogLogger) WarnContext(ctx context.Context, msg string, args ...any) {
	s.log.WarnContext(ctx, msg, args...)
}

func (s *slogLogger) Error(msg string, args ...any) {
	s.log.Error(msg, args...)
}

func (s *slogLogger) ErrorContext(ctx context.Context, msg string, args ...any) {
	s.log.ErrorContext(ctx, msg, args...)
}
