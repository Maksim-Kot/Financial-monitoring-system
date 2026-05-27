package logger

import "context"

type LogLevel int

const (
	DebugLevel LogLevel = -4
	InfoLevel  LogLevel = 0
	WarnLevel  LogLevel = 4
	ErrorLevel LogLevel = 8
)

type Logger interface {
	With(kv ...any) Logger
	WithGroup(name string) Logger
	Debug(msg string, args ...any)
	DebugContext(ctx context.Context, msg string, args ...any)
	Info(msg string, args ...any)
	InfoContext(ctx context.Context, msg string, args ...any)
	Warn(msg string, args ...any)
	WarnContext(ctx context.Context, msg string, args ...any)
	Error(msg string, args ...any)
	ErrorContext(ctx context.Context, msg string, args ...any)
}

func parseLogLevel(level string) LogLevel {
	switch level {
	case "info":
		return InfoLevel
	case "warn":
		return WarnLevel
	case "error":
		return ErrorLevel
	default:
		return DebugLevel
	}
}
