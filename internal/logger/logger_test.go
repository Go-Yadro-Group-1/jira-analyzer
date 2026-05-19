package logger_test

import (
	"bytes"
	"context"
	"log/slog"
	"testing"

	"github.com/Go-Yadro-Group-1/Jira-Analyzer/internal/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew_ValidLevels(t *testing.T) {
	t.Parallel()

	for _, level := range []string{"debug", "info", "warn", "error", "DEBUG"} {
		t.Run(level, func(t *testing.T) {
			t.Parallel()

			got, err := logger.New(&bytes.Buffer{}, level)
			require.NoError(t, err)
			assert.NotNil(t, got)
		})
	}
}

func TestNew_UnknownLevelReturnsError(t *testing.T) {
	t.Parallel()

	_, err := logger.New(&bytes.Buffer{}, "trace")

	require.ErrorIs(t, err, logger.ErrUnknownLogLevel)
}

func TestNew_RespectsLevel(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer

	log, err := logger.New(&buf, "warn")
	require.NoError(t, err)

	log.Info("info message")
	log.Warn("warn message")

	out := buf.String()
	assert.NotContains(t, out, "info message")
	assert.Contains(t, out, "warn message")
}

func TestFromContext_NoLoggerReturnsDefault(t *testing.T) {
	t.Parallel()

	got := logger.FromContext(context.Background())

	assert.Same(t, slog.Default(), got)
}

func TestFromContext_ReturnsStoredLogger(t *testing.T) {
	t.Parallel()

	want, err := logger.New(&bytes.Buffer{}, "info")
	require.NoError(t, err)

	ctx := logger.WithLogger(context.Background(), want)

	assert.Same(t, want, logger.FromContext(ctx))
}
