package logger

import (
	"context"
	"log/slog"
	"os"
	"strings"
)

const (
	LevelTrace = slog.Level(-8)
	LevelFatal = slog.Level(12)
)

type Logger struct {
	*slog.Logger
}

func New(level string) *Logger {
	lvl := slog.LevelInfo
	switch strings.ToLower(strings.TrimSpace(level)) {
	case "trace":
		lvl = LevelTrace
	case "debug":
		lvl = slog.LevelDebug
	case "info":
		lvl = slog.LevelInfo
	case "warn", "warning":
		lvl = slog.LevelWarn
	case "error":
		lvl = slog.LevelError
	case "fatal":
		lvl = LevelFatal
	}

	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{AddSource: false, Level: lvl})
	log := slog.New(handler)

	return &Logger{Logger: log}
}

func attrsToAny(attrs []slog.Attr) []any {
	out := make([]any, 0, len(attrs))
	for _, a := range attrs {
		out = append(out, a)
	}
	return out
}

func (l *Logger) Trace(msg string, attrs ...slog.Attr) {
	l.Log(context.Background(), LevelTrace, msg, attrsToAny(attrs)...)
}

func (l *Logger) Fatal(msg string, attrs ...slog.Attr) {
	l.Log(context.Background(), LevelFatal, msg, attrsToAny(attrs)...)
	os.Exit(1)
}
