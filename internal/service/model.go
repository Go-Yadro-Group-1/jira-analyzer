package service

import "time"

type DurationUnit string

const (
	UnitMinute DurationUnit = "minute"
	UnitHour   DurationUnit = "hour"
	UnitDay    DurationUnit = "day"
	UnitMonth  DurationUnit = "month"
	UnitYear   DurationUnit = "year"
)

const MaxYearBars = 8

type IssuesDurationHistogram struct {
	Unit DurationUnit
	Bars []int
}

type IssuesTimeSpentHistogram struct {
	Unit DurationUnit
	Bars []int
}

type StatusHistogram struct {
	Status string
	Unit   DurationUnit
	Bars   []int
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
