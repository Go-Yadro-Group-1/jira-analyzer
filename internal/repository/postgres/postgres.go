package postgres

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"time"

	"github.com/Go-Yadro-Group-1/Jira-Analyzer/internal/repository"
)

//go:embed queries
var queriesFS embed.FS

func mustQuery(name string) string {
	b, err := queriesFS.ReadFile("queries/" + name)
	if err != nil {
		panic(err)
	}

	return string(b)
}

// nolint: gochecknoglobals
var (
	getProjectLastUpdatedQuery         = mustQuery("get_project_last_updated.sql")
	getStatsByProjectQuery             = mustQuery("get_stats_by_project.sql")
	getIssuesDurationByProjectQuery    = mustQuery("get_issues_duration_by_project.sql")
	getStatusTransitionsByProjectQuery = mustQuery("get_status_transitions_by_project.sql")
	getDailyActivityByProjectQuery     = mustQuery("get_daily_activity_by_project.sql")
	getIssuesTimeSpentByProjectQuery   = mustQuery("get_issues_time_spent_by_project.sql")
	getPriorityStatsByProjectQuery     = mustQuery("get_priority_stats_by_project.sql")
)

type ProjectQueryError struct {
	ProjectID int
	Action    string
	Err       error
}

func (e *ProjectQueryError) Error() string {
	return fmt.Sprintf("failed to %s for project %d: %v", e.Action, e.ProjectID, e.Err)
}

func (e *ProjectQueryError) Unwrap() error {
	return e.Err
}

func newProjectErr(action string, projectID int, err error) error {
	return &ProjectQueryError{
		ProjectID: projectID,
		Action:    action,
		Err:       err,
	}
}

type Postgres struct {
	db *sql.DB
}

func New(db *sql.DB) *Postgres {
	return &Postgres{
		db: db,
	}
}

func (p *Postgres) GetProjectLastUpdated(
	ctx context.Context,
	projectID int,
) (time.Time, error) {
	var t time.Time

	err := p.db.QueryRowContext(ctx, getProjectLastUpdatedQuery, projectID).Scan(&t)
	if err != nil {
		return time.Time{}, newProjectErr("get last updated", projectID, err)
	}

	return t, nil
}

func (p *Postgres) GetStatsByProject(
	ctx context.Context,
	projectID int,
) (repository.ProjectStats, error) {
	var stats repository.ProjectStats

	stats.ProjectID = projectID

	err := p.db.QueryRowContext(ctx, getStatsByProjectQuery, projectID).Scan(
		&stats.CountTotal,
		&stats.CountOpen,
		&stats.CountClosed,
		&stats.CountReopened,
		&stats.CountResolved,
		&stats.CountInProgress,
		&stats.TotalDurationClosed,
		&stats.CountCreatedLastWeek,
	)
	if err != nil {
		return repository.ProjectStats{}, newProjectErr("get stats", projectID, err)
	}

	return stats, nil
}

func (p *Postgres) GetIssuesDurationByProject(
	ctx context.Context,
	projectID int,
) ([]repository.IssueDuration, error) {
	rows, err := p.db.QueryContext(ctx, getIssuesDurationByProjectQuery, projectID)
	if err != nil {
		return nil, newProjectErr("query issues duration", projectID, err)
	}
	defer rows.Close()

	var result []repository.IssueDuration

	for rows.Next() {
		var item repository.IssueDuration

		err := rows.Scan(&item.IssueID, &item.Duration)
		if err != nil {
			return nil, newProjectErr("scan issue duration", projectID, err)
		}

		result = append(result, item)
	}

	err = rows.Err()
	if err != nil {
		return nil, newProjectErr("iterate issue duration rows", projectID, err)
	}

	return result, nil
}

func (p *Postgres) GetStatusTransitionsByProject(
	ctx context.Context,
	projectID int,
) ([]repository.StatusTransition, error) {
	rows, err := p.db.QueryContext(ctx, getStatusTransitionsByProjectQuery, projectID)
	if err != nil {
		return nil, newProjectErr("query status transitions", projectID, err)
	}
	defer rows.Close()

	var result []repository.StatusTransition

	for rows.Next() {
		var item repository.StatusTransition

		err := rows.Scan(&item.ChangeTime, &item.FromStatus, &item.ToStatus)
		if err != nil {
			return nil, newProjectErr("scan status transition", projectID, err)
		}

		result = append(result, item)
	}

	err = rows.Err()
	if err != nil {
		return nil, newProjectErr("iterate status transition rows", projectID, err)
	}

	return result, nil
}

func (p *Postgres) GetDailyActivityByProject(
	ctx context.Context,
	projectID int,
) ([]repository.DailyActivity, error) {
	rows, err := p.db.QueryContext(ctx, getDailyActivityByProjectQuery, projectID)
	if err != nil {
		return nil, newProjectErr("query daily activity", projectID, err)
	}
	defer rows.Close()

	var result []repository.DailyActivity

	for rows.Next() {
		var item repository.DailyActivity

		err := rows.Scan(&item.Date, &item.Creation, &item.Completion)
		if err != nil {
			return nil, newProjectErr("scan daily activity", projectID, err)
		}

		result = append(result, item)
	}

	err = rows.Err()
	if err != nil {
		return nil, newProjectErr("iterate daily activity rows", projectID, err)
	}

	return result, nil
}

func (p *Postgres) GetIssuesTimeSpentByProject(
	ctx context.Context,
	projectID int,
) ([]repository.IssueTimeSpent, error) {
	rows, err := p.db.QueryContext(ctx, getIssuesTimeSpentByProjectQuery, projectID)
	if err != nil {
		return nil, newProjectErr("query time spent", projectID, err)
	}
	defer rows.Close()

	var result []repository.IssueTimeSpent

	for rows.Next() {
		var item repository.IssueTimeSpent

		err := rows.Scan(&item.IssueID, &item.TimeSpent)
		if err != nil {
			return nil, newProjectErr("scan time spent", projectID, err)
		}

		result = append(result, item)
	}

	err = rows.Err()
	if err != nil {
		return nil, newProjectErr("iterate time spent rows", projectID, err)
	}

	return result, nil
}

func (p *Postgres) GetPriorityStatsByProject(
	ctx context.Context,
	projectID int,
) ([]repository.PriorityStats, error) {
	rows, err := p.db.QueryContext(ctx, getPriorityStatsByProjectQuery, projectID)
	if err != nil {
		return nil, newProjectErr("query priority stats", projectID, err)
	}
	defer rows.Close()

	var result []repository.PriorityStats

	for rows.Next() {
		var item repository.PriorityStats

		err := rows.Scan(&item.Priority, &item.Count)
		if err != nil {
			return nil, newProjectErr("scan priority stats", projectID, err)
		}

		result = append(result, item)
	}

	err = rows.Err()
	if err != nil {
		return nil, newProjectErr("iterate priority stats rows", projectID, err)
	}

	return result, nil
}
