package service

import "time"

type Bar struct {
	Label  string
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
