package postgres

import "time"

// ProjectStats get raw statistics for a Jira project ОФ2.
type ProjectStats struct {
	ProjectID        int
	Total            int
	Open             int
	Closed           int
	Reopened         int
	Resolved         int
	InProgress       int
	AvgDurationHours time.Duration
	AvgDailyLastWeek float64
}

// IssueDuration information about the duration of closed issues ОФ2.1.
type IssueDuration struct {
	IssueID  int
	Duration time.Duration
}

// IssueByStatusDuration time a closed issue spent in each status ОФ2.2.
type IssueByStatusDuration struct {
	IssueID  int
	Status   string
	Duration time.Duration
}

// DailyActivity number of created and closed issues per day ОФ2.3.
type DailyActivity struct {
	Date       time.Time
	Creation   int
	Completion int
}

// IssueTimeSpent logged time spent on a closed issue ОФ2.4.
type IssueTimeSpent struct {
	IssueID   int
	TimeSpent time.Duration
}

// PriorityStats number of issues per priority level ОФ2.5.
type PriorityStats struct {
	Priority string
	Count    int
}
