package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type Postgres struct {
	db *sql.DB
}

func New(db *sql.DB) *Postgres {
	return &Postgres{
		db: db,
	}
}

// nolint: revive, godoclint, funlen
func (p *Postgres) GetStatsByProject(ctx context.Context, projectID int) (ProjectStats, error) {
	var stats ProjectStats
	stats.ProjectID = projectID

	query := `
		WITH issue_stats AS (
			SELECT 
				COUNT(*) AS total,
				COUNT(*) FILTER (WHERE status = 'Open') AS open,
				COUNT(*) FILTER (WHERE status = 'Closed') AS closed,
				COUNT(*) FILTER (WHERE status = 'Resolved') AS resolved,
				COUNT(*) FILTER (WHERE status = 'In Progress') AS in_progress,
				AVG(EXTRACT(EPOCH FROM (closed_time - created_time)) / 3600.0) 
					FILTER (WHERE closed_time IS NOT NULL AND created_time IS NOT NULL) AS avg_duration_hours
			FROM raw.issue
			WHERE project_id = $1
		),
		reopened_count AS (
			SELECT COUNT(DISTINCT issue_id) AS reopened
			FROM raw.status_changes
			WHERE issue_id IN (SELECT id FROM raw.issue WHERE project_id = $1)
				AND from_status IN ('Closed', 'Resolved', 'Done')
				AND to_status IN ('Open', 'Reopened', 'In Progress', 'To Do')
		),
		daily_last_week AS (
			SELECT 
				COUNT(*)::float / 7.0 AS avg_daily
			FROM raw.issue
			WHERE project_id = $1
				AND created_time >= NOW() - INTERVAL '7 days'
		)
		SELECT 
			COALESCE(i.total, 0),
			COALESCE(i.open, 0),
			COALESCE(i.closed, 0),
			COALESCE(r.reopened, 0),
			COALESCE(i.resolved, 0),
			COALESCE(i.in_progress, 0),
			COALESCE(i.avg_duration_hours, 0),
			COALESCE(d.avg_daily, 0)
		FROM issue_stats i
		CROSS JOIN reopened_count r
		CROSS JOIN daily_last_week d
	`

	var avgDurationHours float64

	err := p.db.QueryRowContext(ctx, query, projectID).Scan(
		&stats.Total,
		&stats.Open,
		&stats.Closed,
		&stats.Reopened,
		&stats.Resolved,
		&stats.InProgress,
		&avgDurationHours,
		&stats.AvgDailyLastWeek,
	)
	if err != nil {
		return ProjectStats{}, fmt.Errorf("failed to get stats for project %d: %w", projectID, err)
	}

	stats.AvgDurationHours = time.Duration(avgDurationHours * float64(time.Hour))

	return stats, nil
}

// nolint: revive, godoclint
func (p *Postgres) GetIssuesDurationByProject(
	ctx context.Context,
	projectID int,
) ([]IssueDuration, error) {
	return []IssueDuration{}, nil
}

// nolint: revive, godoclint
func (p *Postgres) GetIssuesByStatusDurationByProject(
	ctx context.Context,
	projectID int,
) ([]IssueByStatusDuration, error) {
	return []IssueByStatusDuration{}, nil
}

// nolint: revive, godoclint
func (p *Postgres) GetDailyActivityByProject(
	ctx context.Context,
	projectID int,
) ([]DailyActivity, error) {
	return []DailyActivity{}, nil
}

// nolint: revive, godoclint
func (p *Postgres) GetIssuesTimeSpentByProject(
	ctx context.Context,
	projectID int,
) ([]IssueTimeSpent, error) {
	return []IssueTimeSpent{}, nil
}

// nolint: revive, godoclint
func (p *Postgres) GetPriorityStatsByProject(
	ctx context.Context,
	projectID int,
) ([]PriorityStats, error) {
	return []PriorityStats{}, nil
}
