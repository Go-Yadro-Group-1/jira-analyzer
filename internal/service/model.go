package service

import (
	"errors"
	"time"

	"github.com/Go-Yadro-Group-1/Jira-Analyzer/internal/repository"
)

type ChartType string

const (
	ChartTypeOpenStateHistogram  ChartType = "open_state_histogram"
	ChartTypeStateDistribution   ChartType = "state_distribution"
	ChartTypeComplexityHistogram ChartType = "complexity_histogram"
	ChartTypePriority            ChartType = "priority"
	ChartTypeDailyActivity       ChartType = "daily_activity"
)

var ErrUnknownChartType = errors.New("unknown chart type")

type ProjectStats struct {
	CountTotal               int
	CountOpen                int
	CountClosed              int
	CountReopened            int
	CountResolved            int
	CountInProgress          int
	AvgCompletionTimeHours   float64
	AvgCreatedPerDayLastWeek float64
}

type HistogramBar struct {
	Label string `json:"label"`
	Count int    `json:"count"`
}

type IssuesDurationHistogram struct {
	Bars []HistogramBar `json:"bars"`
}

type IssuesTimeSpentHistogram struct {
	Bars []HistogramBar `json:"bars"`
}

type StatusHistogram struct {
	Status string         `json:"status"`
	Bars   []HistogramBar `json:"bars"`
}

type DailyActivityPoint struct {
	Date              time.Time `json:"date"`
	Created           int       `json:"created"`
	Closed            int       `json:"closed"`
	CumulativeCreated int       `json:"cumulativeCreated"`
	CumulativeClosed  int       `json:"cumulativeClosed"`
}

type DailyActivityChart struct {
	Points []DailyActivityPoint `json:"points"`
}

type PriorityBar struct {
	Priority string `json:"priority"`
	Count    int    `json:"count"`
}

type PriorityChart struct {
	Bars []PriorityBar `json:"bars"`
}

const (
	statsSecondsPerHour = 3600.0
	statsDaysInWeek     = 7.0
)

func toProjectStats(raw repository.ProjectStats) ProjectStats {
	var avgCompletionTimeHours float64
	if raw.CountClosed > 0 {
		avgCompletionTimeHours = float64(raw.TotalDurationClosed) /
			(float64(raw.CountClosed) * statsSecondsPerHour)
	}

	return ProjectStats{
		CountTotal:               raw.CountTotal,
		CountOpen:                raw.CountOpen,
		CountClosed:              raw.CountClosed,
		CountReopened:            raw.CountReopened,
		CountResolved:            raw.CountResolved,
		CountInProgress:          raw.CountInProgress,
		AvgCompletionTimeHours:   avgCompletionTimeHours,
		AvgCreatedPerDayLastWeek: float64(raw.CountCreatedLastWeek) / statsDaysInWeek,
	}
}
