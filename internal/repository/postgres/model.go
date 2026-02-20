package postgres

import "time"

// get raw statistics for a Jira project ОФ2
type ProjectStats struct {
	ProjectID        int
	Total            int
	Open             int
	Closed           int
	Reopened         int
	Resolved         int
	InProgress       int
	AvgDurationHours float64
	AvgDailyLastWeek float64
}

// information about the duration of closed issues ОФ2.1
type IssueDuration struct {
	IssueID  int
	Duration time.Duration
}

// time a closed issue spent in each status ОФ2.2
type IssueByStatusDuration struct {
	IssueID  int
	Status   string
	Duration time.Duration
}

// number of created and closed issues per day ОФ2.3
type DailyActivity struct {
	Date       time.Time
	Creation   int
	Completion int
}

// logged time spent on a closed issue ОФ2.4
type IssueTimeSpent struct {
	IssueID   int
	TimeSpent time.Duration
}

// number of issues per priority level ОФ2.5
type Priority struct {
	Priority string
	Count    int
}
