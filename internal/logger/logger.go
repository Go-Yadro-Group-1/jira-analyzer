package logger

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"strings"
)

var ErrUnknownLogLevel = errors.New("unknown log level")

type ctxKey struct{}

//nolint:gochecknoglobals
var loggerKey ctxKey

func New(out io.Writer, level string) (*slog.Logger, error) {
	lvl, err := parseLevel(level)
	if err != nil {
		return nil, err
	}

	handler := slog.NewJSONHandler(out, &slog.HandlerOptions{Level: lvl}) //nolint:exhaustruct

	return slog.New(handler), nil
}

func parseLevel(level string) (slog.Level, error) {
	switch strings.ToLower(level) {
	case "debug":
		return slog.LevelDebug, nil
	case "info":
		return slog.LevelInfo, nil
	case "warn":
		return slog.LevelWarn, nil
	case "error":
		return slog.LevelError, nil
	}

	return 0, fmt.Errorf("%w: %q", ErrUnknownLogLevel, level)
}

func WithLogger(ctx context.Context, l *slog.Logger) context.Context {
	return context.WithValue(ctx, loggerKey, l)
}

func FromContext(ctx context.Context) *slog.Logger {
	if l, ok := ctx.Value(loggerKey).(*slog.Logger); ok {
		return l
	}

	return slog.Default()
}
