package grpchandler

//go:generate mockgen -destination=mocks/mock_service.go -package=mocks github.com/Go-Yadro-Group-1/Jira-Analyzer/internal/handler/grpc Service

import (
	"context"
	"errors"
	"fmt"

	analyzerv1 "github.com/Go-Yadro-Group-1/Jira-Analyzer/gen/grpc/analyzer/v1"
	"github.com/Go-Yadro-Group-1/Jira-Analyzer/internal/repository"
	"github.com/Go-Yadro-Group-1/Jira-Analyzer/internal/service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Service interface {
	GetChart(ctx context.Context, projectID int, chartType service.ChartType) ([]byte, error)
	GetProjectStat(ctx context.Context, projectID int) (repository.ProjectStats, error)
	CompareTwoProjects(
		ctx context.Context,
		lhsProjectID, rhsProjectID int,
	) ([2]repository.ProjectStats, error)
}

type Handler struct {
	analyzerv1.UnimplementedAnalyzerServiceServer

	svc Service
}

func New(svc Service) *Handler {
	return &Handler{
		UnimplementedAnalyzerServiceServer: analyzerv1.UnimplementedAnalyzerServiceServer{},
		svc:                                svc,
	}
}

func (h *Handler) GetChart(
	ctx context.Context,
	req *analyzerv1.GetChartRequest,
) (*analyzerv1.GetChartResponse, error) {
	chartType, err := protoChartTypeToService(req.GetChartType())
	if err != nil {
		return nil, fmt.Errorf("get chart: %w", status.Error(codes.InvalidArgument, err.Error()))
	}

	data, err := h.svc.GetChart(ctx, int(req.GetProjectId()), chartType)
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &analyzerv1.GetChartResponse{Data: data}, nil
}

func (h *Handler) GetStats(
	ctx context.Context,
	req *analyzerv1.GetStatsRequest,
) (*analyzerv1.GetStatsResponse, error) {
	stats, err := h.svc.GetProjectStat(ctx, int(req.GetProjectId()))
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &analyzerv1.GetStatsResponse{Stats: statsToProto(stats)}, nil
}

func (h *Handler) CompareProjects(
	ctx context.Context,
	req *analyzerv1.CompareProjectsRequest,
) (*analyzerv1.CompareProjectsResponse, error) {
	results, err := h.svc.CompareTwoProjects(
		ctx,
		int(req.GetProjectIdA()),
		int(req.GetProjectIdB()),
	)
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &analyzerv1.CompareProjectsResponse{
		ProjectA: statsToProto(results[0]),
		ProjectB: statsToProto(results[1]),
	}, nil
}

func protoChartTypeToService(ct analyzerv1.ChartType) (service.ChartType, error) {
	switch ct {
	case analyzerv1.ChartType_CHART_TYPE_OPEN_STATE_HISTOGRAM:
		return service.ChartTypeOpenStateHistogram, nil
	case analyzerv1.ChartType_CHART_TYPE_STATE_DISTRIBUTION:
		return service.ChartTypeStateDistribution, nil
	case analyzerv1.ChartType_CHART_TYPE_COMPLEXITY_HISTOGRAM:
		return service.ChartTypeComplexityHistogram, nil
	case analyzerv1.ChartType_CHART_TYPE_PRIORITY:
		return service.ChartTypePriority, nil
	case analyzerv1.ChartType_CHART_TYPE_DAILY_ACTIVITY:
		return service.ChartTypeDailyActivity, nil
	case analyzerv1.ChartType_CHART_TYPE_UNSPECIFIED:
		return "", service.ErrUnknownChartType
	}

	return "", service.ErrUnknownChartType
}

func statsToProto(stats repository.ProjectStats) *analyzerv1.Stats {
	var avgDurationClosed float32
	if stats.CountClosed > 0 {
		avgDurationClosed = float32(
			stats.TotalDurationClosed,
		) / float32(
			stats.CountClosed,
		) //nolint:gosec
	}

	const daysInWeek = 7

	avgCreatedPerDay := float32(stats.CountCreatedLastWeek) / daysInWeek //nolint:gosec

	return &analyzerv1.Stats{
		CountTotal:        int32(stats.CountTotal),      //nolint:gosec
		CountOpen:         int32(stats.CountOpen),       //nolint:gosec
		CountClosed:       int32(stats.CountClosed),     //nolint:gosec
		CountReopened:     int32(stats.CountReopened),   //nolint:gosec
		CountResolved:     int32(stats.CountResolved),   //nolint:gosec
		CountInProgress:   int32(stats.CountInProgress), //nolint:gosec
		AvgDurationClosed: avgDurationClosed,
		AvgCreatedPerDay:  avgCreatedPerDay,
	}
}

func toGRPCError(err error) error {
	if errors.Is(err, service.ErrUnknownChartType) {
		return fmt.Errorf("grpc handler: %w", status.Error(codes.InvalidArgument, err.Error()))
	}

	return fmt.Errorf("grpc handler: %w", status.Error(codes.Internal, err.Error()))
}
