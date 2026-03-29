package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/Go-Yadro-Group-1/Jira-Analyzer/internal/repository"
	"github.com/Go-Yadro-Group-1/Jira-Analyzer/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	hour  = 3600
	day   = 86400
	month = 30 * day
	year  = 365 * day
)

func TestGetIssuesDurationHistogram_Empty(t *testing.T) {
	t.Parallel()

	svc := service.New(&mockRepository{}, newMockCache())

	_, err := svc.GetIssuesDurationHistogram(context.Background(), 1)

	require.ErrorIs(t, err, service.ErrNoClosedIssues)
}

func TestGetIssuesDurationHistogram_HourZone(t *testing.T) {
	t.Parallel()

	repo := &mockRepository{
		issuesDuration: []repository.IssueDuration{
			{IssueID: 1, Duration: 2 * hour},
		},
	}
	svc := service.New(repo, newMockCache())

	hist, err := svc.GetIssuesDurationHistogram(context.Background(), 1)

	require.NoError(t, err)
	require.Len(t, hist.Bars, 3)
	assert.Equal(t, "0h", hist.Bars[0].Label)
	assert.Equal(t, "2h", hist.Bars[2].Label)
	assert.Equal(t, 1, hist.Bars[2].Count)
}

func TestGetIssuesDurationHistogram_DayZone(t *testing.T) {
	t.Parallel()

	repo := &mockRepository{
		issuesDuration: []repository.IssueDuration{
			{IssueID: 1, Duration: 3 * day},
		},
	}
	svc := service.New(repo, newMockCache())

	hist, err := svc.GetIssuesDurationHistogram(context.Background(), 1)

	require.NoError(t, err)
	require.GreaterOrEqual(t, len(hist.Bars), 27)
	assert.Equal(t, "3day", hist.Bars[26].Label)
	assert.Equal(t, 1, hist.Bars[26].Count)
}

func TestGetIssuesDurationHistogram_MonthZone(t *testing.T) {
	t.Parallel()

	repo := &mockRepository{
		issuesDuration: []repository.IssueDuration{
			{IssueID: 1, Duration: 2 * month},
		},
	}
	svc := service.New(repo, newMockCache())

	hist, err := svc.GetIssuesDurationHistogram(context.Background(), 1)

	require.NoError(t, err)
	require.GreaterOrEqual(t, len(hist.Bars), 56)
	assert.Equal(t, "2month", hist.Bars[55].Label)
	assert.Equal(t, 1, hist.Bars[55].Count)
}

func TestGetIssuesDurationHistogram_YearZone(t *testing.T) {
	t.Parallel()

	repo := &mockRepository{
		issuesDuration: []repository.IssueDuration{
			{IssueID: 1, Duration: 3 * year},
		},
	}
	svc := service.New(repo, newMockCache())

	hist, err := svc.GetIssuesDurationHistogram(context.Background(), 1)

	require.NoError(t, err)
	require.GreaterOrEqual(t, len(hist.Bars), 68)
	assert.Equal(t, "3year", hist.Bars[67].Label)
	assert.Equal(t, 1, hist.Bars[67].Count)
}

func TestGetIssuesDurationHistogram_MaxYear(t *testing.T) {
	t.Parallel()

	repo := &mockRepository{
		issuesDuration: []repository.IssueDuration{
			{IssueID: 1, Duration: 10 * year},
		},
	}
	svc := service.New(repo, newMockCache())

	hist, err := svc.GetIssuesDurationHistogram(context.Background(), 1)

	require.NoError(t, err)
	assert.Equal(t, "8+year", hist.Bars[72].Label)
	assert.Equal(t, 1, hist.Bars[72].Count)
}

func TestGetIssuesDurationHistogram_MultipleZones(t *testing.T) {
	t.Parallel()

	repo := &mockRepository{
		issuesDuration: []repository.IssueDuration{
			{IssueID: 1, Duration: 1 * hour},
			{IssueID: 2, Duration: 5 * day},
			{IssueID: 3, Duration: 5 * day},
		},
	}
	svc := service.New(repo, newMockCache())

	hist, err := svc.GetIssuesDurationHistogram(context.Background(), 1)

	require.NoError(t, err)
	assert.Equal(t, 1, hist.Bars[1].Count)
	assert.Equal(t, 2, hist.Bars[28].Count)
}

func TestGetStatusHistograms_Empty(t *testing.T) {
	t.Parallel()

	svc := service.New(&mockRepository{}, newMockCache())

	_, err := svc.GetStatusHistograms(context.Background(), 1)

	require.ErrorIs(t, err, service.ErrNoClosedIssues)
}

func TestGetStatusHistograms_SingleIssue(t *testing.T) {
	t.Parallel()

	now := time.Now()
	repo := &mockRepository{
		statusTransitions: []repository.StatusTransition{
			{IssueID: 1, FromStatus: "Open", ToStatus: "In Progress", ChangeTime: now},
			{
				IssueID:    1,
				FromStatus: "In Progress",
				ToStatus:   "Closed",
				ChangeTime: now.Add(2 * time.Hour),
			},
			{IssueID: 1, FromStatus: "Closed", ToStatus: "", ChangeTime: now.Add(3 * time.Hour)},
		},
	}
	svc := service.New(repo, newMockCache())

	histograms, err := svc.GetStatusHistograms(context.Background(), 1)

	require.NoError(t, err)
	require.Len(t, histograms, 2)
	assert.Equal(t, "In Progress", histograms[0].Status)
	assert.Equal(t, "Open", histograms[1].Status)
}

func TestGetStatusHistograms_GroupsByIssue(t *testing.T) {
	t.Parallel()

	now := time.Now()
	repo := &mockRepository{
		statusTransitions: []repository.StatusTransition{
			{IssueID: 1, FromStatus: "Open", ToStatus: "Done", ChangeTime: now},
			{IssueID: 1, FromStatus: "Done", ToStatus: "", ChangeTime: now.Add(time.Hour)},
			{IssueID: 2, FromStatus: "Open", ToStatus: "Done", ChangeTime: now.Add(2 * time.Hour)},
			{IssueID: 2, FromStatus: "Done", ToStatus: "", ChangeTime: now.Add(5 * time.Hour)},
		},
	}
	svc := service.New(repo, newMockCache())

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

	svc := service.New(&mockRepository{}, newMockCache())

	chart, err := svc.GetDailyActivityChart(context.Background(), 1)

	require.NoError(t, err)
	assert.Empty(t, chart.Points)
}

func TestGetDailyActivityChart_SortedWithCumulative(t *testing.T) {
	t.Parallel()

	day1 := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	day2 := time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC)

	repo := &mockRepository{
		dailyActivity: []repository.DailyActivity{
			{Date: day2, Creation: 3, Completion: 1},
			{Date: day1, Creation: 2, Completion: 0},
		},
	}
	svc := service.New(repo, newMockCache())

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

	repo := &mockRepository{
		priorityStats: []repository.PriorityStats{
			{Priority: "High", Count: 5},
			{Priority: "Low", Count: 2},
		},
	}
	svc := service.New(repo, newMockCache())

	chart, err := svc.GetPriorityChart(context.Background(), 1)

	require.NoError(t, err)
	require.Len(t, chart.Bars, 2)
	assert.Equal(t, "High", chart.Bars[0].Priority)
	assert.Equal(t, 5, chart.Bars[0].Count)
}
