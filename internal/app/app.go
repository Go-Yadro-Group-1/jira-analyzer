package app

import (
	"database/sql"

	analyzerv1 "github.com/Go-Yadro-Group-1/Jira-Analyzer/gen/grpc/analyzer/v1"
	grpchandler "github.com/Go-Yadro-Group-1/Jira-Analyzer/internal/handler/grpc"
	"github.com/Go-Yadro-Group-1/Jira-Analyzer/internal/repository/memory"
	"github.com/Go-Yadro-Group-1/Jira-Analyzer/internal/repository/postgres"
	"github.com/Go-Yadro-Group-1/Jira-Analyzer/internal/service"
	"google.golang.org/grpc"
)

func NewGRPCServer(db *sql.DB) *grpc.Server {
	repo := postgres.New(db)
	cache := memory.NewCacheRepository[int, string]()
	svc := service.New(repo, cache)
	handler := grpchandler.New(svc)

	server := grpc.NewServer()
	analyzerv1.RegisterAnalyzerServiceServer(server, handler)

	return server
}
