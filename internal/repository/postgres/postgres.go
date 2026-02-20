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
func (p *Postgres) GetStatsByProject(ctx context.Context, projectID int) (ProjectStats, error) {
	return ProjectStats{}, nil
}

// nolint: revive, godoclint
func (p *Postgres) GetIssuesDurationByProject(ctx context.Context, projectID int) ([]IssueDuration, error) {
	return []IssueDuration{}, nil
}

// nolint: revive, godoclint
func (p *Postgres) GetIssuesByStatusDurationByProject(ctx context.Context, projectID int) ([]IssueByStatusDuration, error) {
	return []IssueByStatusDuration{}, nil
}

// nolint: revive, godoclint
func (p *Postgres) GetDailyActivityByProject(ctx context.Context, projectID int) ([]DailyActivity, error) {
	return []DailyActivity{}, nil
}

// nolint: revive, godoclint
func (p *Postgres) GetIssuesTimeSpentByProject(ctx context.Context, projectID int) ([]IssueTimeSpent, error) {
	return []IssueTimeSpent{}, nil
}

// nolint: revive, godoclint
func (p *Postgres) GetPriorityStatsByProject(ctx context.Context, projectID int) ([]PriorityStats, error) {
	return []PriorityStats{}, nil
}
