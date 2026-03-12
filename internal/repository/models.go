package repository

import (
	"time"
)

type ProjectStats struct {
	ProjectID            int
	CountTotal           int
	CountOpen            int
	CountClosed          int
	CountReopened        int
	CountResolved        int
	CountInProgress      int
	TotalDurationClosed  int
	CountCreatedLastWeek int
}

type IssueDuration struct {
	IssueID  int
	Duration int
}

type StatusTransition struct {
	ChangeTime time.Time
	FromStatus string
	ToStatus   string
}

type DailyActivity struct {
	Date       time.Time
	Creation   int
	Completion int
}

type IssueTimeSpent struct {
	IssueID   int
	TimeSpent int
}

type PriorityStats struct {
	Priority string
	Count    int
}
