package service_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/Go-Yadro-Group-1/Jira-Analyzer/internal/repository"
	"github.com/Go-Yadro-Group-1/Jira-Analyzer/internal/service"
	"github.com/Go-Yadro-Group-1/Jira-Analyzer/internal/service/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

var (
	errDB    = errors.New("db error")
	errCache = errors.New("cache miss")
)

func setupCacheMiss(repo *mocks.MockRepository, cache *mocks.MockCache) {
	repo.EXPECT().
		GetProjectLastUpdated(gomock.Any(), gomock.Any()).
		Return(time.Now(), nil).
		AnyTimes()
	cache.EXPECT().
		GetLastUpdated(gomock.Any(), gomock.Any()).
		Return(time.Time{}, errCache).
		AnyTimes()
	cache.EXPECT().
		Set(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return(nil).
		AnyTimes()
	cache.EXPECT().SetLastUpdated(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
}

func TestGetChart_UnknownChartType(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	svc := service.New(mocks.NewMockRepository(ctrl), mocks.NewMockCache(ctrl))

	_, err := svc.GetChart(context.Background(), 1, "nonexistent")

	require.ErrorIs(t, err, service.ErrUnknownChartType)
}

func TestGetChart_AllTypes(t *testing.T) {
	t.Parallel()

	now := time.Now()

	chartTypes := []service.ChartType{
		service.ChartTypeOpenStateHistogram,
		service.ChartTypeStateDistribution,
		service.ChartTypeDailyActivity,
		service.ChartTypeComplexityHistogram,
		service.ChartTypePriority,
	}

	for _, chartType := range chartTypes {
		t.Run(string(chartType), func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			repo := mocks.NewMockRepository(ctrl)
			cache := mocks.NewMockCache(ctrl)

			setupCacheMiss(repo, cache)

			repo.EXPECT().GetIssuesDurationByProject(gomock.Any(), 1).
				Return([]repository.IssueDuration{{IssueID: 1, Duration: hour}}, nil).
				AnyTimes()
			repo.EXPECT().GetStatusTransitionsByProject(gomock.Any(), 1).Return(
				[]repository.StatusTransition{
					{IssueID: 1, FromStatus: "Open", ToStatus: "Closed", ChangeTime: now},
					{
						IssueID:    1,
						FromStatus: "Closed",
						ToStatus:   "",
						ChangeTime: now.Add(time.Hour),
					},
				}, nil,
			).AnyTimes()
			repo.EXPECT().GetDailyActivityByProject(gomock.Any(), 1).
				Return([]repository.DailyActivity{{Date: now, Creation: 1, Completion: 0}}, nil).
				AnyTimes()
			repo.EXPECT().GetIssuesTimeSpentByProject(gomock.Any(), 1).
				Return([]repository.IssueTimeSpent{{IssueID: 1, TimeSpent: hour}}, nil).
				AnyTimes()
			repo.EXPECT().GetPriorityStatsByProject(gomock.Any(), 1).
				Return([]repository.PriorityStats{{Priority: "High", Count: 5}}, nil).
				AnyTimes()

			svc := service.New(repo, cache)

			data, err := svc.GetChart(context.Background(), 1, chartType)

			require.NoError(t, err)
			assert.NotEmpty(t, data)
		})
	}
}

func TestGetChart_CacheHit(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	repo := mocks.NewMockRepository(ctrl)
	cache := mocks.NewMockCache(ctrl)
	ctx := context.Background()

	dbTime := time.Now()

	cachedHist := service.IssuesDurationHistogram{
		Bars: []service.HistogramBar{
			{Label: "0h", Count: 0},
			{Label: "1h", Count: 1},
		},
	}

	cachedData, err := json.Marshal(cachedHist)
	require.NoError(t, err)

	gomock.InOrder(
		repo.EXPECT().GetProjectLastUpdated(gomock.Any(), 1).Return(dbTime, nil),
		cache.EXPECT().GetLastUpdated(gomock.Any(), 1).Return(time.Time{}, errCache),
		repo.EXPECT().GetIssuesDurationByProject(gomock.Any(), 1).
			Return([]repository.IssueDuration{{IssueID: 1, Duration: hour}}, nil),
		cache.EXPECT().Set(gomock.Any(), 1, gomock.Any(), gomock.Any()).Return(nil),
		cache.EXPECT().SetLastUpdated(gomock.Any(), 1, gomock.Any()).Return(nil),
		repo.EXPECT().GetProjectLastUpdated(gomock.Any(), 1).Return(dbTime, nil),
		cache.EXPECT().GetLastUpdated(gomock.Any(), 1).Return(dbTime, nil),
		cache.EXPECT().Get(gomock.Any(), 1, gomock.Any()).Return(cachedData, nil),
	)

	svc := service.New(repo, cache)

	_, err = svc.GetChart(ctx, 1, service.ChartTypeOpenStateHistogram)
	require.NoError(t, err)

	data, err := svc.GetChart(ctx, 1, service.ChartTypeOpenStateHistogram)
	require.NoError(t, err)

	var result service.IssuesDurationHistogram

	require.NoError(t, json.Unmarshal(data, &result))
	assert.Equal(t, "1h", result.Bars[1].Label)
}

func TestGetChart_StaleCache(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	repo := mocks.NewMockRepository(ctrl)
	cache := mocks.NewMockCache(ctrl)

	dbTime := time.Now()
	cacheTime := dbTime.Add(-time.Hour)

	staleData, err := json.Marshal(service.IssuesDurationHistogram{Bars: nil})
	require.NoError(t, err)

	repo.EXPECT().GetProjectLastUpdated(gomock.Any(), 1).Return(dbTime, nil)
	cache.EXPECT().GetLastUpdated(gomock.Any(), 1).Return(cacheTime, nil)
	repo.EXPECT().GetIssuesDurationByProject(gomock.Any(), 1).
		Return([]repository.IssueDuration{{IssueID: 1, Duration: 2 * hour}}, nil)
	cache.EXPECT().Set(gomock.Any(), 1, gomock.Any(), gomock.Any()).Return(nil)
	cache.EXPECT().SetLastUpdated(gomock.Any(), 1, gomock.Any()).Return(nil)

	_ = staleData

	svc := service.New(repo, cache)

	data, err := svc.GetChart(context.Background(), 1, service.ChartTypeOpenStateHistogram)

	require.NoError(t, err)

	var result service.IssuesDurationHistogram

	require.NoError(t, json.Unmarshal(data, &result))
	assert.NotEmpty(t, result.Bars)
}

func TestGetChart_RepoError(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	repo := mocks.NewMockRepository(ctrl)
	cache := mocks.NewMockCache(ctrl)

	repo.EXPECT().GetProjectLastUpdated(gomock.Any(), 1).Return(time.Now(), nil)
	cache.EXPECT().GetLastUpdated(gomock.Any(), 1).Return(time.Time{}, errCache)
	repo.EXPECT().GetIssuesDurationByProject(gomock.Any(), 1).Return(nil, errDB)

	svc := service.New(repo, cache)

	_, err := svc.GetChart(context.Background(), 1, service.ChartTypeOpenStateHistogram)

	require.Error(t, err)
}

func TestGetChart_RepoError_AllChartTypes(t *testing.T) {
	t.Parallel()

	now := time.Now()

	tests := []struct {
		chartType service.ChartType
		setupRepo func(*mocks.MockRepository)
	}{
		{
			chartType: service.ChartTypeStateDistribution,
			setupRepo: func(repo *mocks.MockRepository) {
				repo.EXPECT().GetStatusTransitionsByProject(gomock.Any(), 1).Return(nil, errDB)
			},
		},
		{
			chartType: service.ChartTypeDailyActivity,
			setupRepo: func(repo *mocks.MockRepository) {
				repo.EXPECT().GetDailyActivityByProject(gomock.Any(), 1).Return(nil, errDB)
			},
		},
		{
			chartType: service.ChartTypeComplexityHistogram,
			setupRepo: func(repo *mocks.MockRepository) {
				repo.EXPECT().GetIssuesTimeSpentByProject(gomock.Any(), 1).Return(nil, errDB)
			},
		},
		{
			chartType: service.ChartTypePriority,
			setupRepo: func(repo *mocks.MockRepository) {
				repo.EXPECT().GetPriorityStatsByProject(gomock.Any(), 1).Return(nil, errDB)
			},
		},
	}

	for _, test := range tests {
		t.Run(string(test.chartType), func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			repo := mocks.NewMockRepository(ctrl)
			cache := mocks.NewMockCache(ctrl)

			repo.EXPECT().GetProjectLastUpdated(gomock.Any(), 1).Return(now, nil)
			cache.EXPECT().GetLastUpdated(gomock.Any(), 1).Return(time.Time{}, errCache)
			test.setupRepo(repo)

			_, err := service.New(repo, cache).GetChart(context.Background(), 1, test.chartType)

			require.Error(t, err)
		})
	}
}

func TestGetProjectStat(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	repo := mocks.NewMockRepository(ctrl)
	cache := mocks.NewMockCache(ctrl)

	raw := repository.ProjectStats{
		CountTotal:           10,
		CountOpen:            3,
		CountClosed:          7,
		TotalDurationClosed:  25200, // 7 задач × 3600с = 1 час на задачу
		CountCreatedLastWeek: 14,
	}

	setupCacheMiss(repo, cache)
	repo.EXPECT().GetStatsByProject(gomock.Any(), 1).Return(raw, nil)

	got, err := service.New(repo, cache).GetProjectStat(context.Background(), 1)

	require.NoError(t, err)
	assert.Equal(t, 10, got.CountTotal)
	assert.Equal(t, 3, got.CountOpen)
	assert.Equal(t, 7, got.CountClosed)
	assert.InDelta(t, 1.0, got.AvgCompletionTimeHours, 0.001)
	assert.InDelta(t, 2.0, got.AvgCreatedPerDayLastWeek, 0.001)
}

func TestGetProjectStat_ZeroClosedIssues(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	repo := mocks.NewMockRepository(ctrl)
	cache := mocks.NewMockCache(ctrl)

	raw := repository.ProjectStats{
		CountTotal:  5,
		CountOpen:   5,
		CountClosed: 0,
	}

	setupCacheMiss(repo, cache)
	repo.EXPECT().GetStatsByProject(gomock.Any(), 1).Return(raw, nil)

	got, err := service.New(repo, cache).GetProjectStat(context.Background(), 1)

	require.NoError(t, err)
	assert.InDelta(t, 0.0, got.AvgCompletionTimeHours, 0.001)
	assert.InDelta(t, 0.0, got.AvgCreatedPerDayLastWeek, 0.001)
}

func TestCompareTwoProjects(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	repo := mocks.NewMockRepository(ctrl)
	cache := mocks.NewMockCache(ctrl)

	stats := repository.ProjectStats{CountTotal: 5}

	setupCacheMiss(repo, cache)
	repo.EXPECT().GetStatsByProject(gomock.Any(), gomock.Any()).Return(stats, nil).Times(2)

	result, err := service.New(repo, cache).CompareTwoProjects(context.Background(), 1, 2)

	require.NoError(t, err)
	assert.Equal(t, 5, result[0].CountTotal)
	assert.Equal(t, 5, result[1].CountTotal)
}

func TestCompareTwoProjects_RepoError(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	repo := mocks.NewMockRepository(ctrl)
	cache := mocks.NewMockCache(ctrl)

	setupCacheMiss(repo, cache)
	repo.EXPECT().
		GetStatsByProject(gomock.Any(), gomock.Any()).
		Return(repository.ProjectStats{}, errDB).
		AnyTimes()

	_, err := service.New(repo, cache).CompareTwoProjects(context.Background(), 1, 2)

	require.Error(t, err)
}

func TestCompareProjectsCharts(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	repo := mocks.NewMockRepository(ctrl)
	cache := mocks.NewMockCache(ctrl)

	setupCacheMiss(repo, cache)
	repo.EXPECT().GetPriorityStatsByProject(gomock.Any(), gomock.Any()).
		Return([]repository.PriorityStats{{Priority: "High", Count: 3}}, nil).
		Times(2)

	result, err := service.New(repo, cache).CompareProjectsCharts(
		context.Background(), 1, 2, service.ChartTypePriority,
	)

	require.NoError(t, err)
	assert.NotEmpty(t, result[0])
	assert.NotEmpty(t, result[1])
}

func TestCompareProjectsCharts_UnknownType(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)

	_, err := service.New(mocks.NewMockRepository(ctrl), mocks.NewMockCache(ctrl)).
		CompareProjectsCharts(context.Background(), 1, 2, "bad_type")

	require.Error(t, err)
}
