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

// HistogramBar — один столбец мультимасштабной гистограммы.
// Label содержит человекочитаемое обозначение временного диапазона:
// "0h"–"23h", "1day"–"30day", "1month"–"11month", "1year"–"7year", "8+year".
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
