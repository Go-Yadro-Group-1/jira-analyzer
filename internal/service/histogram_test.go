package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/Go-Yadro-Group-1/Jira-Analyzer/internal/repository"
	"github.com/Go-Yadro-Group-1/Jira-Analyzer/internal/service"
	"github.com/Go-Yadro-Group-1/Jira-Analyzer/internal/service/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

const (
	hour  = 3600
	day   = 86400
	month = 30 * day
	year  = 365 * day
)

func TestGetIssuesDurationHistogram_Empty(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	repo := mocks.NewMockRepository(ctrl)
	cache := mocks.NewMockCache(ctrl)

	setupCacheMiss(repo, cache)
	repo.EXPECT().GetIssuesDurationByProject(gomock.Any(), 1).Return(nil, nil)

	svc := service.New(repo, cache)

	_, err := svc.GetIssuesDurationHistogram(context.Background(), 1)

	require.ErrorIs(t, err, service.ErrNoHistogramData)
}

func TestGetIssuesDurationHistogram_HourZone(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	repo := mocks.NewMockRepository(ctrl)
	cache := mocks.NewMockCache(ctrl)

	setupCacheMiss(repo, cache)
	repo.EXPECT().GetIssuesDurationByProject(gomock.Any(), 1).
		Return([]repository.IssueDuration{{IssueID: 1, Duration: 2 * hour}}, nil)

	svc := service.New(repo, cache)

	hist, err := svc.GetIssuesDurationHistogram(context.Background(), 1)

	require.NoError(t, err)
	require.Len(t, hist.Bars, 3)
	assert.Equal(t, "0h", hist.Bars[0].Label)
	assert.Equal(t, "2h", hist.Bars[2].Label)
	assert.Equal(t, 1, hist.Bars[2].Count)
}

func TestGetIssuesDurationHistogram_DayZone(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	repo := mocks.NewMockRepository(ctrl)
	cache := mocks.NewMockCache(ctrl)

	setupCacheMiss(repo, cache)
	repo.EXPECT().GetIssuesDurationByProject(gomock.Any(), 1).
		Return([]repository.IssueDuration{{IssueID: 1, Duration: 3 * day}}, nil)

	svc := service.New(repo, cache)

	hist, err := svc.GetIssuesDurationHistogram(context.Background(), 1)

	require.NoError(t, err)
	require.GreaterOrEqual(t, len(hist.Bars), 27)
	assert.Equal(t, "3day", hist.Bars[26].Label)
	assert.Equal(t, 1, hist.Bars[26].Count)
}

func TestGetIssuesDurationHistogram_MonthZone(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	repo := mocks.NewMockRepository(ctrl)
	cache := mocks.NewMockCache(ctrl)

	setupCacheMiss(repo, cache)
	repo.EXPECT().GetIssuesDurationByProject(gomock.Any(), 1).
		Return([]repository.IssueDuration{{IssueID: 1, Duration: 2 * month}}, nil)

	svc := service.New(repo, cache)

	hist, err := svc.GetIssuesDurationHistogram(context.Background(), 1)

	require.NoError(t, err)
	require.GreaterOrEqual(t, len(hist.Bars), 56)
	assert.Equal(t, "2month", hist.Bars[55].Label)
	assert.Equal(t, 1, hist.Bars[55].Count)
}

func TestGetIssuesDurationHistogram_YearZone(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	repo := mocks.NewMockRepository(ctrl)
	cache := mocks.NewMockCache(ctrl)

	setupCacheMiss(repo, cache)
	repo.EXPECT().GetIssuesDurationByProject(gomock.Any(), 1).
		Return([]repository.IssueDuration{{IssueID: 1, Duration: 3 * year}}, nil)

	svc := service.New(repo, cache)

	hist, err := svc.GetIssuesDurationHistogram(context.Background(), 1)

	require.NoError(t, err)
	require.GreaterOrEqual(t, len(hist.Bars), 68)
	assert.Equal(t, "3year", hist.Bars[67].Label)
	assert.Equal(t, 1, hist.Bars[67].Count)
}

func TestGetIssuesDurationHistogram_MaxYear(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	repo := mocks.NewMockRepository(ctrl)
	cache := mocks.NewMockCache(ctrl)

	setupCacheMiss(repo, cache)
	repo.EXPECT().GetIssuesDurationByProject(gomock.Any(), 1).
		Return([]repository.IssueDuration{{IssueID: 1, Duration: 10 * year}}, nil)

	svc := service.New(repo, cache)

	hist, err := svc.GetIssuesDurationHistogram(context.Background(), 1)

	require.NoError(t, err)
	assert.Equal(t, "8+year", hist.Bars[72].Label)
	assert.Equal(t, 1, hist.Bars[72].Count)
}

func TestGetIssuesDurationHistogram_MultipleZones(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	repo := mocks.NewMockRepository(ctrl)
	cache := mocks.NewMockCache(ctrl)

	setupCacheMiss(repo, cache)
	repo.EXPECT().GetIssuesDurationByProject(gomock.Any(), 1).Return(
		[]repository.IssueDuration{
			{IssueID: 1, Duration: 1 * hour},
			{IssueID: 2, Duration: 5 * day},
			{IssueID: 3, Duration: 5 * day},
		}, nil,
	)

	svc := service.New(repo, cache)

	hist, err := svc.GetIssuesDurationHistogram(context.Background(), 1)

	require.NoError(t, err)
	assert.Equal(t, 1, hist.Bars[1].Count)
	assert.Equal(t, 2, hist.Bars[28].Count)
}

func TestGetStatusHistograms_Empty(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	repo := mocks.NewMockRepository(ctrl)
	cache := mocks.NewMockCache(ctrl)

	setupCacheMiss(repo, cache)
	repo.EXPECT().GetStatusTransitionsByProject(gomock.Any(), 1).Return(nil, nil)

	svc := service.New(repo, cache)

	_, err := svc.GetStatusHistograms(context.Background(), 1)

	require.ErrorIs(t, err, service.ErrNoHistogramData)
}

func TestGetStatusHistograms_SingleIssue(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	repo := mocks.NewMockRepository(ctrl)
	cache := mocks.NewMockCache(ctrl)
	now := time.Now()

	setupCacheMiss(repo, cache)
	repo.EXPECT().GetStatusTransitionsByProject(gomock.Any(), 1).Return(
		[]repository.StatusTransition{
			{IssueID: 1, FromStatus: "Open", ToStatus: "In Progress", ChangeTime: now},
			{
				IssueID:    1,
				FromStatus: "In Progress",
				ToStatus:   "Closed",
				ChangeTime: now.Add(2 * time.Hour),
			},
			{IssueID: 1, FromStatus: "Closed", ToStatus: "", ChangeTime: now.Add(3 * time.Hour)},
		}, nil,
	)

	svc := service.New(repo, cache)

	histograms, err := svc.GetStatusHistograms(context.Background(), 1)

	require.NoError(t, err)
	require.Len(t, histograms, 2)
	assert.Equal(t, "In Progress", histograms[0].Status)
	assert.Equal(t, "Open", histograms[1].Status)
}

func TestGetStatusHistograms_GroupsByIssue(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	repo := mocks.NewMockRepository(ctrl)
	cache := mocks.NewMockCache(ctrl)
	now := time.Now()

	setupCacheMiss(repo, cache)
	repo.EXPECT().GetStatusTransitionsByProject(gomock.Any(), 1).Return(
		[]repository.StatusTransition{
			{IssueID: 1, FromStatus: "Open", ToStatus: "Done", ChangeTime: now},
			{IssueID: 1, FromStatus: "Done", ToStatus: "", ChangeTime: now.Add(time.Hour)},
			{IssueID: 2, FromStatus: "Open", ToStatus: "Done", ChangeTime: now.Add(2 * time.Hour)},
			{IssueID: 2, FromStatus: "Done", ToStatus: "", ChangeTime: now.Add(5 * time.Hour)},
		}, nil,
	)

	svc := service.New(repo, cache)

	histograms, err := svc.GetStatusHistograms(context.Background(), 1)

	require.NoError(t, err)

	var openTotal int

	for _, hist := range histograms {
		if hist.Status == "Open" {
			for _, bar := range hist.Bars {
				openTotal += bar.Count
			}
		}
	}

	assert.Equal(t, 2, openTotal)
}

func TestGetDailyActivityChart_Empty(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	repo := mocks.NewMockRepository(ctrl)
	cache := mocks.NewMockCache(ctrl)

	setupCacheMiss(repo, cache)
	repo.EXPECT().GetDailyActivityByProject(gomock.Any(), 1).Return(nil, nil)

	svc := service.New(repo, cache)

	chart, err := svc.GetDailyActivityChart(context.Background(), 1)

	require.NoError(t, err)
	assert.Empty(t, chart.Points)
}

func TestGetDailyActivityChart_SortedWithCumulative(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	repo := mocks.NewMockRepository(ctrl)
	cache := mocks.NewMockCache(ctrl)

	day1 := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	day2 := time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC)

	setupCacheMiss(repo, cache)
	repo.EXPECT().GetDailyActivityByProject(gomock.Any(), 1).Return(
		[]repository.DailyActivity{
			{Date: day2, Creation: 3, Completion: 1},
			{Date: day1, Creation: 2, Completion: 0},
		}, nil,
	)

	svc := service.New(repo, cache)

	chart, err := svc.GetDailyActivityChart(context.Background(), 1)

	require.NoError(t, err)
	require.Len(t, chart.Points, 2)
	assert.Equal(t, day1, chart.Points[0].Date)
	assert.Equal(t, 2, chart.Points[0].CumulativeCreated)
	assert.Equal(t, 5, chart.Points[1].CumulativeCreated)
	assert.Equal(t, 1, chart.Points[1].CumulativeClosed)
}

func TestGetPriorityChart(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	repo := mocks.NewMockRepository(ctrl)
	cache := mocks.NewMockCache(ctrl)

	setupCacheMiss(repo, cache)
	repo.EXPECT().GetPriorityStatsByProject(gomock.Any(), 1).Return(
		[]repository.PriorityStats{
			{Priority: "High", Count: 5},
			{Priority: "Low", Count: 2},
		}, nil,
	)

	svc := service.New(repo, cache)

	chart, err := svc.GetPriorityChart(context.Background(), 1)

	require.NoError(t, err)
	require.Len(t, chart.Bars, 2)
	assert.Equal(t, "High", chart.Bars[0].Priority)
	assert.Equal(t, 5, chart.Bars[0].Count)
}
