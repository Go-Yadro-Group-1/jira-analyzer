package postgres

import (
	"context"
	"database/sql"
)

type Postgres struct {
	db *sql.DB
}

func New(db *sql.DB) *Postgres {
	return &Postgres{
		db: db,
	}
}

// nolint: revive, godoclint
func GetStatsByProject(ctx context.Context, projectID int) (ProjectStats, error) {
	return ProjectStats{}, nil
}
