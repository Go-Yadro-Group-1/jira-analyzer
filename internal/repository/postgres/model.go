package postgres

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
