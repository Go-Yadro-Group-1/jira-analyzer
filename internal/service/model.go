package service

import (
	"fmt"
	"time"
)

type DurationLabel string

const (
	LabelOverflowYear DurationLabel = "8+year"
)

func HourLabel(h int64) DurationLabel    { return DurationLabel(fmt.Sprintf("%dh", h)) }
func DayLabel(d int64) DurationLabel     { return DurationLabel(fmt.Sprintf("%dday", d)) }
func MonthLabel(m int64) DurationLabel   { return DurationLabel(fmt.Sprintf("%dmonth", m)) }
func YearLabel(y int64) DurationLabel    { return DurationLabel(fmt.Sprintf("%dyear", y)) }

type Bar struct {
	Label  DurationLabel
	Height int
}

type IssuesDurationHistogram struct {
	Bars []Bar
}

type StatusHistogram struct {
	Status string
	Bars   []Bar
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

type IssuesTimeSpentHistogram struct {
	Bars []Bar
}

type PriorityChart struct {
	Bars []Bar
}
