package service

import (
	"errors"
	"time"
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
