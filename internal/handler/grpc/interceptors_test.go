package grpchandler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"testing"

	grpchandler "github.com/Go-Yadro-Group-1/Jira-Analyzer/internal/handler/grpc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

var errBoom = errors.New("boom")

func decodeJSONLines(t *testing.T, raw []byte) []map[string]any {
	t.Helper()

	var entries []map[string]any

	dec := json.NewDecoder(bytes.NewReader(raw))
	for dec.More() {
		var entry map[string]any
		require.NoError(t, dec.Decode(&entry))

		entries = append(entries, entry)
	}

	return entries
}

func runInterceptor(
	ctx context.Context,
	t *testing.T,
	handler grpc.UnaryHandler,
) ([]map[string]any, error) {
	t.Helper()

	var buf bytes.Buffer

	base := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug}))
	interceptor := grpchandler.UnaryServerLogging(base)
	info := &grpc.UnaryServerInfo{FullMethod: "/analyzer.v1.AnalyzerService/GetStats"}

	_, err := interceptor(ctx, nil, info, handler)

	return decodeJSONLines(t, buf.Bytes()), err
}

func TestUnaryServerLogging_SuccessLogsInfo(t *testing.T) {
	t.Parallel()

	entries, err := runInterceptor(context.Background(), t,
		func(_ context.Context, _ any) (any, error) { return "ok", nil },
	)
	require.NoError(t, err)

	require.Len(t, entries, 2)
	assert.Equal(t, "DEBUG", entries[0]["level"])
	assert.Equal(t, "INFO", entries[1]["level"])
	assert.Equal(t, "rpc done", entries[1]["msg"])
	assert.Equal(t, codes.OK.String(), entries[1]["code"])
}

func TestUnaryServerLogging_InternalErrorLogsError(t *testing.T) {
	t.Parallel()

	entries, err := runInterceptor(context.Background(), t,
		func(_ context.Context, _ any) (any, error) {
			return nil, status.Error(codes.Internal, "db down")
		},
	)
	require.Error(t, err)

	last := entries[len(entries)-1]
	assert.Equal(t, "ERROR", last["level"])
	assert.Equal(t, codes.Internal.String(), last["code"])
}

func TestUnaryServerLogging_ClientErrorLogsWarn(t *testing.T) {
	t.Parallel()

	entries, err := runInterceptor(context.Background(), t,
		func(_ context.Context, _ any) (any, error) {
			return nil, status.Error(codes.InvalidArgument, "bad input")
		},
	)
	require.Error(t, err)

	last := entries[len(entries)-1]
	assert.Equal(t, "WARN", last["level"])
	assert.Equal(t, codes.InvalidArgument.String(), last["code"])
}

func TestUnaryServerLogging_UnknownErrorLogsError(t *testing.T) {
	t.Parallel()

	entries, err := runInterceptor(context.Background(), t,
		func(_ context.Context, _ any) (any, error) { return nil, errBoom },
	)
	require.Error(t, err)

	last := entries[len(entries)-1]
	assert.Equal(t, "ERROR", last["level"])
	assert.Equal(t, codes.Unknown.String(), last["code"])
}

func TestUnaryServerLogging_UsesRequestIDFromMetadata(t *testing.T) {
	t.Parallel()

	ctx := metadata.NewIncomingContext(context.Background(),
		metadata.Pairs("x-request-id", "fixed-id-123"),
	)

	entries, err := runInterceptor(ctx, t,
		func(_ context.Context, _ any) (any, error) { return "ok", nil },
	)
	require.NoError(t, err)

	for _, entry := range entries {
		assert.Equal(t, "fixed-id-123", entry["request_id"])
	}
}

func TestUnaryServerLogging_GeneratesRequestIDWhenMissing(t *testing.T) {
	t.Parallel()

	entries, err := runInterceptor(context.Background(), t,
		func(_ context.Context, _ any) (any, error) { return "ok", nil },
	)
	require.NoError(t, err)

	reqID, ok := entries[0]["request_id"].(string)
	require.True(t, ok)
	assert.NotEmpty(t, reqID)
	assert.NotEqual(t, "unknown", reqID)
}
