package postgres

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"time"

	"github.com/Go-Yadro-Group-1/Jira-Analyzer/internal/metrics"
	"github.com/Go-Yadro-Group-1/Jira-Analyzer/internal/repository"
	"github.com/prometheus/client_golang/prometheus"
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
	listProjectsQuery                  = mustQuery("list_projects.sql")
	deleteProjectQuery                 = mustQuery("delete_project.sql")
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
	db      *sql.DB
	metrics *metrics.Metrics
}

func New(db *sql.DB, metrics *metrics.Metrics) *Postgres {
	return &Postgres{
		db:      db,
		metrics: metrics,
	}
}

func (p *Postgres) GetProjectLastUpdated(
	ctx context.Context,
	projectID int,
) (time.Time, error) {
	defer p.observe("get_project_last_updated").ObserveDuration()

	var timeT time.Time

	err := p.db.QueryRowContext(ctx, getProjectLastUpdatedQuery, projectID).Scan(&timeT)
	if err != nil {
		return time.Time{}, newProjectErr("get last updated", projectID, err)
	}

	return timeT, nil
}

func (p *Postgres) GetStatsByProject(
	ctx context.Context,
	projectID int,
) (repository.ProjectStats, error) {
	defer p.observe("get_stats_by_project").ObserveDuration()

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
	defer p.observe("get_issues_duration_by_project").ObserveDuration()

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
	defer p.observe("get_status_transitions_by_project").ObserveDuration()

	rows, err := p.db.QueryContext(ctx, getStatusTransitionsByProjectQuery, projectID)
	if err != nil {
		return nil, newProjectErr("query status transitions", projectID, err)
	}
	defer rows.Close()

	var result []repository.StatusTransition

	for rows.Next() {
		var item repository.StatusTransition

		err := rows.Scan(&item.IssueID, &item.ChangeTime, &item.FromStatus, &item.ToStatus)
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
	defer p.observe("get_daily_activity_by_project").ObserveDuration()

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
	defer p.observe("get_issues_time_spent_by_project").ObserveDuration()

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
	defer p.observe("get_priority_stats_by_project").ObserveDuration()

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

func (p *Postgres) ListProjects(ctx context.Context) ([]repository.Project, error) {
	defer p.observe("list_projects").ObserveDuration()

	rows, err := p.db.QueryContext(ctx, listProjectsQuery)
	if err != nil {
		return nil, fmt.Errorf("list projects: %w", err)
	}
	defer rows.Close()

	var result []repository.Project

	for rows.Next() {
		var project repository.Project

		err := rows.Scan(&project.ID, &project.Title)
		if err != nil {
			return nil, fmt.Errorf("scan project: %w", err)
		}

		result = append(result, project)
	}

	err = rows.Err()
	if err != nil {
		return nil, fmt.Errorf("iterate project rows: %w", err)
	}

	return result, nil
}

func (p *Postgres) DeleteProject(ctx context.Context, projectID int) error {
	defer p.observe("delete_project").ObserveDuration()

	result, err := p.db.ExecContext(ctx, deleteProjectQuery, projectID)
	if err != nil {
		return newProjectErr("delete", projectID, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return newProjectErr("check delete result", projectID, err)
	}

	if rowsAffected == 0 {
		return newProjectErr("delete", projectID, sql.ErrNoRows)
	}

	return nil
}

func (p *Postgres) observe(query string) *prometheus.Timer {
	return prometheus.NewTimer(p.metrics.DBQueryDuration.WithLabelValues(query))
}
