package postgres

import "database/sql"

type Postgres struct {
	db *sql.DB
}

func New(db *sql.DB) *Postgres {
	return &Postgres{
		db: db,
	}
}

// Get Count of Issues in Project by ID
func GetIssueCountInProject(projectID int) int

// Get Count of Open Issues in Project by ID
func GetOpenIssueCountInProject(projectID int) int
