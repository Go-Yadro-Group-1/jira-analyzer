package app

import (
	"database/sql"
	"log/slog"

	analyzerv1 "github.com/Go-Yadro-Group-1/Jira-Analyzer/gen/grpc/analyzer/v1"
	grpchandler "github.com/Go-Yadro-Group-1/Jira-Analyzer/internal/handler/grpc"
	"github.com/Go-Yadro-Group-1/Jira-Analyzer/internal/metrics"
	"github.com/Go-Yadro-Group-1/Jira-Analyzer/internal/repository/memory"
	"github.com/Go-Yadro-Group-1/Jira-Analyzer/internal/repository/postgres"
	"github.com/Go-Yadro-Group-1/Jira-Analyzer/internal/service"
	"google.golang.org/grpc"
)

func NewGRPCServer(db *sql.DB, log *slog.Logger, mtr *metrics.Metrics) *grpc.Server {
	repo := postgres.New(db, mtr)
	cache := memory.NewCacheRepository[int, string]()
	svc := service.New(repo, cache, mtr)
	handler := grpchandler.New(svc)

	server := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			grpchandler.UnaryServerLogging(log),
			mtr.GRPCServer.UnaryServerInterceptor(),
		),
	)
	analyzerv1.RegisterAnalyzerServiceServer(server, handler)
	mtr.GRPCServer.InitializeMetrics(server)

	return server
}
