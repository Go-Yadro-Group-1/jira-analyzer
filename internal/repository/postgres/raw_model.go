package postgres

import "time"

type Project struct {
	ID    int
	Title string
}

type Author struct {
	ID   int
	Name string
}

type Issue struct {
	ID          int
	ProjectID   int
	AuthorID    int
	AssigneeID  int
	Key         string
	Summary     string
	Description string
	Type        string
	Priority    string
	Status      string
	CreatedAt   time.Time
	ClosedAt    time.Time
	UpdatedAt   time.Time
	TimeSpent   int
}

type StatusChange struct {
	ID         int
	IssueID    int
	FromStatus string
	ToStatus   string
	ChangedAt  time.Time
}
