package grpchandler

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"time"

	"github.com/Go-Yadro-Group-1/Jira-Analyzer/internal/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const requestIDMetaKey = "x-request-id"

func UnaryServerLogging(base *slog.Logger) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		start := time.Now()
		reqID := extractOrGenerateRequestID(ctx)

		rpcLog := base.With(
			slog.String("request_id", reqID),
			slog.String("method", info.FullMethod),
		)

		rpcLog.DebugContext(ctx, "rpc start")

		ctx = logger.WithLogger(ctx, rpcLog)
		resp, err := handler(ctx, req)

		duration := time.Since(start)
		code := status.Code(err)

		attrs := []any{
			slog.Duration("duration", duration),
			slog.String("code", code.String()),
		}
		if err != nil {
			attrs = append(attrs, slog.String("error", err.Error()))
		}

		switch {
		case err == nil:
			rpcLog.InfoContext(ctx, "rpc done", attrs...)
		case isServerError(code):
			rpcLog.ErrorContext(ctx, "rpc failed", attrs...)
		default:
			rpcLog.WarnContext(ctx, "rpc failed", attrs...)
		}

		return resp, err
	}
}

//nolint:exhaustive
func isServerError(code codes.Code) bool {
	switch code {
	case codes.Internal, codes.Unknown, codes.DataLoss, codes.Unavailable:
		return true
	default:
		return false
	}
}

func extractOrGenerateRequestID(ctx context.Context) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		vals := md.Get(requestIDMetaKey)
		if len(vals) > 0 && vals[0] != "" {
			return vals[0]
		}
	}

	var buf [8]byte

	_, err := rand.Read(buf[:])
	if err != nil {
		return "unknown"
	}

	return hex.EncodeToString(buf[:])
}
