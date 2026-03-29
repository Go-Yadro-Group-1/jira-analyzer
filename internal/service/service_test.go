package service_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/Go-Yadro-Group-1/Jira-Analyzer/internal/repository"
	"github.com/Go-Yadro-Group-1/Jira-Analyzer/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	errNotFound = errors.New("not found")
	errDB       = errors.New("db error")
)

type mockRepository struct {
	lastUpdated          time.Time
	lastUpdatedErr       error
	stats                repository.ProjectStats
	statsErr             error
	issuesDuration       []repository.IssueDuration
	issuesDurationErr    error
	statusTransitions    []repository.StatusTransition
	statusTransitionsErr error
	dailyActivity        []repository.DailyActivity
	dailyActivityErr     error
	issuesTimeSpent      []repository.IssueTimeSpent
	issuesTimeSpentErr   error
	priorityStats        []repository.PriorityStats
	priorityStatsErr     error
}

func (r *mockRepository) GetProjectLastUpdated(_ context.Context, _ int) (time.Time, error) {
	return r.lastUpdated, r.lastUpdatedErr
}

func (r *mockRepository) GetStatsByProject(
	_ context.Context,
	_ int,
) (repository.ProjectStats, error) {
	return r.stats, r.statsErr
}

func (r *mockRepository) GetIssuesDurationByProject(
	_ context.Context,
	_ int,
) ([]repository.IssueDuration, error) {
	return r.issuesDuration, r.issuesDurationErr
}

func (r *mockRepository) GetStatusTransitionsByProject(
	_ context.Context,
	_ int,
) ([]repository.StatusTransition, error) {
	return r.statusTransitions, r.statusTransitionsErr
}

func (r *mockRepository) GetDailyActivityByProject(
	_ context.Context,
	_ int,
) ([]repository.DailyActivity, error) {
	return r.dailyActivity, r.dailyActivityErr
}

func (r *mockRepository) GetIssuesTimeSpentByProject(
	_ context.Context,
	_ int,
) ([]repository.IssueTimeSpent, error) {
	return r.issuesTimeSpent, r.issuesTimeSpentErr
}

func (r *mockRepository) GetPriorityStatsByProject(
	_ context.Context,
	_ int,
) ([]repository.PriorityStats, error) {
	return r.priorityStats, r.priorityStatsErr
}

type mockCache struct {
	data              map[string][]byte
	lastUpdated       map[int]time.Time
	getLastUpdatedErr error
}

func newMockCache() *mockCache {
	return &mockCache{
		data:        make(map[string][]byte),
		lastUpdated: make(map[int]time.Time),
	}
}

func (c *mockCache) Get(_ context.Context, projectID int, dataType string) ([]byte, error) {
	key := fmt.Sprintf("%d:%s", projectID, dataType)

	val, ok := c.data[key]
	if !ok {
		return nil, errNotFound
	}

	return val, nil
}

func (c *mockCache) Set(_ context.Context, projectID int, dataType string, value []byte) error {
	c.data[fmt.Sprintf("%d:%s", projectID, dataType)] = value

	return nil
}

func (c *mockCache) Invalidate(_ context.Context, _ int) error {
	return nil
}

func (c *mockCache) GetLastUpdated(_ context.Context, projectID int) (time.Time, error) {
	if c.getLastUpdatedErr != nil {
		return time.Time{}, c.getLastUpdatedErr
	}

	t, ok := c.lastUpdated[projectID]
	if !ok {
		return time.Time{}, errNotFound
	}

	return t, nil
}

func (c *mockCache) SetLastUpdated(_ context.Context, projectID int, t time.Time) error {
	c.lastUpdated[projectID] = t

	return nil
}

func TestGetChart_UnknownChartType(t *testing.T) {
	t.Parallel()

	svc := service.New(&mockRepository{}, newMockCache())

	_, err := svc.GetChart(context.Background(), 1, "nonexistent")

	require.ErrorIs(t, err, service.ErrUnknownChartType)
}

func TestGetChart_AllTypes(t *testing.T) {
	t.Parallel()

	now := time.Now()
	repo := &mockRepository{
		issuesDuration: []repository.IssueDuration{
			{IssueID: 1, Duration: hour},
		},
		statusTransitions: []repository.StatusTransition{
			{IssueID: 1, FromStatus: "Open", ToStatus: "Closed", ChangeTime: now},
			{IssueID: 1, FromStatus: "Closed", ToStatus: "", ChangeTime: now.Add(time.Hour)},
		},
		dailyActivity:   []repository.DailyActivity{{Date: now, Creation: 1, Completion: 0}},
		issuesTimeSpent: []repository.IssueTimeSpent{{IssueID: 1, TimeSpent: hour}},
		priorityStats:   []repository.PriorityStats{{Priority: "High", Count: 5}},
	}

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

			svc := service.New(repo, newMockCache())

			data, err := svc.GetChart(context.Background(), 1, chartType)

			require.NoError(t, err)
			assert.NotEmpty(t, data)
		})
	}
}

func TestGetChart_CacheHit(t *testing.T) {
	t.Parallel()

	repo := &mockRepository{
		issuesDuration: []repository.IssueDuration{
			{IssueID: 1, Duration: hour},
		},
	}
	cache := newMockCache()
	svc := service.New(repo, cache)
	ctx := context.Background()

	_, err := svc.GetChart(ctx, 1, service.ChartTypeOpenStateHistogram)
	require.NoError(t, err)

	repo.issuesDuration = []repository.IssueDuration{
		{IssueID: 1, Duration: 5 * year},
	}

	data, err := svc.GetChart(ctx, 1, service.ChartTypeOpenStateHistogram)
	require.NoError(t, err)

	var result service.IssuesDurationHistogram

	require.NoError(t, json.Unmarshal(data, &result))
	assert.Equal(t, "1h", result.Bars[1].Label)
}

func TestGetChart_StaleCache(t *testing.T) {
	t.Parallel()

	dbTime := time.Now()
	cacheTime := dbTime.Add(-time.Hour)

	repo := &mockRepository{
		lastUpdated:    dbTime,
		issuesDuration: []repository.IssueDuration{{IssueID: 1, Duration: 2 * hour}},
	}

	cache := newMockCache()
	cache.lastUpdated[1] = cacheTime

	staleData, err := json.Marshal(service.IssuesDurationHistogram{Bars: nil})
	require.NoError(t, err)

	cache.data["1:issues_duration"] = staleData

	svc := service.New(repo, cache)

	data, err := svc.GetChart(context.Background(), 1, service.ChartTypeOpenStateHistogram)

	require.NoError(t, err)

	var result service.IssuesDurationHistogram

	require.NoError(t, json.Unmarshal(data, &result))
	assert.NotEmpty(t, result.Bars)
}

func TestGetChart_RepoError(t *testing.T) {
	t.Parallel()

	repo := &mockRepository{issuesDurationErr: errDB}
	svc := service.New(repo, newMockCache())

	_, err := svc.GetChart(context.Background(), 1, service.ChartTypeOpenStateHistogram)

	require.Error(t, err)
}

func TestGetProjectStat(t *testing.T) {
	t.Parallel()

	expected := repository.ProjectStats{
		CountTotal:  10,
		CountOpen:   3,
		CountClosed: 7,
	}
	svc := service.New(&mockRepository{stats: expected}, newMockCache())

	got, err := svc.GetProjectStat(context.Background(), 1)

	require.NoError(t, err)
	assert.Equal(t, expected, got)
}

func TestCompareTwoProjects(t *testing.T) {
	t.Parallel()

	repo := &mockRepository{stats: repository.ProjectStats{CountTotal: 5}}

	result, err := service.New(repo, newMockCache()).CompareTwoProjects(context.Background(), 1, 2)

	require.NoError(t, err)
	assert.Equal(t, 5, result[0].CountTotal)
	assert.Equal(t, 5, result[1].CountTotal)
}

func TestCompareTwoProjects_RepoError(t *testing.T) {
	t.Parallel()

	repo := &mockRepository{statsErr: errDB}

	_, err := service.New(repo, newMockCache()).CompareTwoProjects(context.Background(), 1, 2)

	require.Error(t, err)
}

func TestCompareProjectsCharts(t *testing.T) {
	t.Parallel()

	repo := &mockRepository{
		priorityStats: []repository.PriorityStats{{Priority: "High", Count: 3}},
	}

	result, err := service.New(repo, newMockCache()).CompareProjectsCharts(
		context.Background(), 1, 2, service.ChartTypePriority,
	)

	require.NoError(t, err)
	assert.NotEmpty(t, result[0])
	assert.NotEmpty(t, result[1])
}

func TestCompareProjectsCharts_UnknownType(t *testing.T) {
	t.Parallel()

	_, err := service.New(&mockRepository{}, newMockCache()).CompareProjectsCharts(
		context.Background(), 1, 2, "bad_type",
	)

	require.Error(t, err)
}
