package grpchandler_test

import (
	"context"
	"errors"
	"testing"

	analyzerv1 "github.com/Go-Yadro-Group-1/Jira-Analyzer/gen/grpc/analyzer/v1"
	grpchandler "github.com/Go-Yadro-Group-1/Jira-Analyzer/internal/handler/grpc"
	"github.com/Go-Yadro-Group-1/Jira-Analyzer/internal/handler/grpc/mocks"
	"github.com/Go-Yadro-Group-1/Jira-Analyzer/internal/repository"
	"github.com/Go-Yadro-Group-1/Jira-Analyzer/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var errService = errors.New("service error")

func newHandler(t *testing.T) (*grpchandler.Handler, *mocks.MockService) {
	t.Helper()

	ctrl := gomock.NewController(t)
	svc := mocks.NewMockService(ctrl)

	return grpchandler.New(svc), svc
}

// GetChart

func TestGetChart_Success(t *testing.T) {
	t.Parallel()

	cases := []struct {
		protoType   analyzerv1.ChartType
		serviceType service.ChartType
	}{
		{analyzerv1.ChartType_CHART_TYPE_OPEN_STATE_HISTOGRAM, service.ChartTypeOpenStateHistogram},
		{analyzerv1.ChartType_CHART_TYPE_STATE_DISTRIBUTION, service.ChartTypeStateDistribution},
		{
			analyzerv1.ChartType_CHART_TYPE_COMPLEXITY_HISTOGRAM,
			service.ChartTypeComplexityHistogram,
		},
		{analyzerv1.ChartType_CHART_TYPE_PRIORITY, service.ChartTypePriority},
		{analyzerv1.ChartType_CHART_TYPE_DAILY_ACTIVITY, service.ChartTypeDailyActivity},
	}

	for _, testcase := range cases {
		t.Run(testcase.protoType.String(), func(t *testing.T) {
			t.Parallel()

			handler, svc := newHandler(t)

			svc.EXPECT().
				GetChart(gomock.Any(), 42, testcase.serviceType).
				Return([]byte(`{"bars":[]}`), nil)

			resp, err := handler.GetChart(context.Background(), &analyzerv1.GetChartRequest{
				ProjectId: 42,
				ChartType: testcase.protoType,
			})

			require.NoError(t, err)
			assert.JSONEq(t, `{"bars":[]}`, string(resp.GetData()))
		})
	}
}

func TestGetChart_UnspecifiedChartType(t *testing.T) {
	t.Parallel()

	handler, _ := newHandler(t)

	_, err := handler.GetChart(context.Background(), &analyzerv1.GetChartRequest{
		ProjectId: 1,
		ChartType: analyzerv1.ChartType_CHART_TYPE_UNSPECIFIED,
	})

	require.Error(t, err)
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
}

func TestGetChart_ServiceError(t *testing.T) {
	t.Parallel()

	handler, svc := newHandler(t)

	svc.EXPECT().
		GetChart(gomock.Any(), 1, service.ChartTypePriority).
		Return(nil, errService)

	_, err := handler.GetChart(context.Background(), &analyzerv1.GetChartRequest{
		ProjectId: 1,
		ChartType: analyzerv1.ChartType_CHART_TYPE_PRIORITY,
	})

	require.Error(t, err)
	assert.Equal(t, codes.Internal, status.Code(err))
}

// GetStats

func TestGetStats_Success(t *testing.T) {
	t.Parallel()

	handler, svc := newHandler(t)

	stats := repository.ProjectStats{
		CountTotal:           100,
		CountOpen:            20,
		CountClosed:          60,
		CountReopened:        5,
		CountResolved:        10,
		CountInProgress:      5,
		TotalDurationClosed:  600,
		CountCreatedLastWeek: 14,
	}

	svc.EXPECT().GetProjectStat(gomock.Any(), 7).Return(stats, nil)

	resp, err := handler.GetStats(context.Background(), &analyzerv1.GetStatsRequest{ProjectId: 7})

	require.NoError(t, err)
	require.NotNil(t, resp.GetStats())

	got := resp.GetStats()
	assert.Equal(t, int32(100), got.GetCountTotal())
	assert.Equal(t, int32(20), got.GetCountOpen())
	assert.Equal(t, int32(60), got.GetCountClosed())
	assert.Equal(t, int32(5), got.GetCountReopened())
	assert.Equal(t, int32(10), got.GetCountResolved())
	assert.Equal(t, int32(5), got.GetCountInProgress())
	assert.InDelta(t, float32(10.0), got.GetAvgDurationClosed(), 0.001) // 600/60
	assert.InDelta(t, float32(2.0), got.GetAvgCreatedPerDay(), 0.001)   // 14/7
}

func TestGetStats_ZeroClosedCount(t *testing.T) {
	t.Parallel()

	handler, svc := newHandler(t)

	svc.EXPECT().
		GetProjectStat(gomock.Any(), 1).
		Return(repository.ProjectStats{CountClosed: 0, TotalDurationClosed: 999}, nil)

	resp, err := handler.GetStats(context.Background(), &analyzerv1.GetStatsRequest{ProjectId: 1})

	require.NoError(t, err)
	assert.InDelta(t, float32(0), resp.GetStats().GetAvgDurationClosed(), 0.001)
}

func TestGetStats_ServiceError(t *testing.T) {
	t.Parallel()

	handler, svc := newHandler(t)

	svc.EXPECT().GetProjectStat(gomock.Any(), 1).Return(repository.ProjectStats{}, errService)

	_, err := handler.GetStats(context.Background(), &analyzerv1.GetStatsRequest{ProjectId: 1})

	require.Error(t, err)
	assert.Equal(t, codes.Internal, status.Code(err))
}

// CompareProjects

func TestCompareProjects_Success(t *testing.T) {
	t.Parallel()

	handler, svc := newHandler(t)

	statsA := repository.ProjectStats{CountTotal: 10, CountClosed: 5, TotalDurationClosed: 50}
	statsB := repository.ProjectStats{CountTotal: 20, CountClosed: 10, TotalDurationClosed: 200}

	svc.EXPECT().
		CompareTwoProjects(gomock.Any(), 1, 2).
		Return([2]repository.ProjectStats{statsA, statsB}, nil)

	resp, err := handler.CompareProjects(context.Background(), &analyzerv1.CompareProjectsRequest{
		ProjectIdA: 1,
		ProjectIdB: 2,
	})

	require.NoError(t, err)
	assert.Equal(t, int32(10), resp.GetProjectA().GetCountTotal())
	assert.Equal(t, int32(20), resp.GetProjectB().GetCountTotal())
	assert.InDelta(t, float32(10.0), resp.GetProjectA().GetAvgDurationClosed(), 0.001) // 50/5
	assert.InDelta(t, float32(20.0), resp.GetProjectB().GetAvgDurationClosed(), 0.001) // 200/10
}

func TestCompareProjects_ServiceError(t *testing.T) {
	t.Parallel()

	handler, svc := newHandler(t)

	svc.EXPECT().
		CompareTwoProjects(gomock.Any(), 1, 2).
		Return([2]repository.ProjectStats{}, errService)

	_, err := handler.CompareProjects(context.Background(), &analyzerv1.CompareProjectsRequest{
		ProjectIdA: 1,
		ProjectIdB: 2,
	})

	require.Error(t, err)
	assert.Equal(t, codes.Internal, status.Code(err))
}
