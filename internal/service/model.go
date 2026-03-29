package service

import "time"

type ChartType string

const (
	ChartTypeOpenStateHistogram  ChartType = "open_state_histogram"
	ChartTypeStateDistribution   ChartType = "state_distribution"
	ChartTypeComplexityHistogram ChartType = "complexity_histogram"
	ChartTypePriority            ChartType = "priority"
	ChartTypeDailyActivity       ChartType = "daily_activity"
)

type HistogramBar struct {
	Label string
	Count int
}

type IssuesDurationHistogram struct {
	Bars []HistogramBar
}

type IssuesTimeSpentHistogram struct {
	Bars []HistogramBar
}

type StatusHistogram struct {
	Status string
	Bars   []HistogramBar
}

type DailyActivityPoint struct {
	Date              time.Time
	Created           int
	Closed            int
	CumulativeCreated int
	CumulativeClosed  int
}

type DailyActivityChart struct {
	Points []DailyActivityPoint
}

type PriorityBar struct {
	Priority string
	Count    int
}

type PriorityChart struct {
	Bars []PriorityBar
}
