//go:build integration

package postgres_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/Go-Yadro-Group-1/Jira-Analyzer/internal/metrics"
	"github.com/Go-Yadro-Group-1/Jira-Analyzer/internal/repository/postgres"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListProjects_ReturnsSeedProject(t *testing.T) {
	t.Parallel()

	repo := postgres.New(database, metrics.New())

	ctx, cancel := context.WithTimeout(t.Context(), dbTimeout)
	defer cancel()

	got, err := repo.ListProjects(ctx)
	require.NoError(t, err)

	// Migration 000003 seeds project id=1 "Test Project".
	require.NotEmpty(t, got)

	found := false

	for _, p := range got {
		if p.ID == 1 {
			assert.Equal(t, "Test Project", p.Title)

			found = true
		}
	}

	assert.True(t, found, "expected seed project id=1 in list")
}

func TestListProjects_OrderedByID(t *testing.T) {
	t.Parallel()

	repo := postgres.New(database, metrics.New())

	ctx, cancel := context.WithTimeout(t.Context(), dbTimeout)
	defer cancel()

	got, err := repo.ListProjects(ctx)
	require.NoError(t, err)

	for i := 1; i < len(got); i++ {
		assert.GreaterOrEqual(t, got[i].ID, got[i-1].ID, "projects must be ordered by id ASC")
	}
}

// TestListProjects_ModelFields is a sanity check that all fields are populated.
func TestListProjects_ModelFields(t *testing.T) {
	t.Parallel()

	repo := postgres.New(database, metrics.New())

	ctx, cancel := context.WithTimeout(t.Context(), dbTimeout)
	defer cancel()

	got, err := repo.ListProjects(ctx)
	require.NoError(t, err)

	for _, p := range got {
		assert.NotZero(t, p.ID, "project ID must be non-zero")
		assert.NotEmpty(t, p.Title, "project Title must be non-empty")
	}
}

// TestDeleteProject_NotFound verifies that deleting a non-existent project returns sql.ErrNoRows.
func TestDeleteProject_NotFound(t *testing.T) {
	t.Parallel()

	repo := postgres.New(database, metrics.New())

	ctx, cancel := context.WithTimeout(t.Context(), dbTimeout)
	defer cancel()

	err := repo.DeleteProject(ctx, 999999)
	require.Error(t, err)
	assert.ErrorIs(t, err, sql.ErrNoRows)
}

// seedDeleteTestProject inserts an isolated project with one issue and one status change
// for use in cascade-delete tests.
func seedDeleteTestProject(ctx context.Context, t *testing.T, projectID, authorID, issueID int) {
	t.Helper()

	_, err := database.ExecContext(ctx,
		`INSERT INTO raw.project (id, title) VALUES ($1, $2)`,
		projectID, "Delete Me",
	)
	require.NoError(t, err)

	_, err = database.ExecContext(ctx,
		`INSERT INTO raw.author (id, name) VALUES ($1, $2) ON CONFLICT DO NOTHING`,
		authorID, "Ghost",
	)
	require.NoError(t, err)

	_, err = database.ExecContext(ctx, `
		INSERT INTO raw.issue
			(id, project_id, author_id, key, summary, type, priority, status, created_time)
		VALUES ($1, $2, $3, 'DEL-1', 'to delete', 'Task', 'Low', 'Open', NOW())`,
		issueID, projectID, authorID,
	)
	require.NoError(t, err)

	_, err = database.ExecContext(ctx, `
		INSERT INTO raw.status_changes (issue_id, author_id, change_time, from_status, to_status)
		VALUES ($1, $2, NOW(), 'Open', 'Closed')`,
		issueID, authorID,
	)
	require.NoError(t, err)
}

// assertCascadeDeleted checks that the project, its issues and status_changes are all gone.
func assertCascadeDeleted(ctx context.Context, t *testing.T, projectID, issueID int) {
	t.Helper()

	var count int

	err := database.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM raw.project WHERE id = $1`, projectID,
	).Scan(&count)
	require.NoError(t, err)
	assert.Zero(t, count, "project row should be deleted")

	err = database.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM raw.issue WHERE id = $1`, issueID,
	).Scan(&count)
	require.NoError(t, err)
	assert.Zero(t, count, "issue row should be cascade-deleted")

	err = database.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM raw.status_changes WHERE issue_id = $1`, issueID,
	).Scan(&count)
	require.NoError(t, err)
	assert.Zero(t, count, "status_changes rows should be cascade-deleted")
}

// TestDeleteProject_CascadesAndReturnsNotFoundAfter inserts a fresh project with issues
// and status_changes, deletes it, then verifies that:
//   - all related rows are gone (cascade works)
//   - a second delete attempt returns sql.ErrNoRows
//
// Not parallel — uses INSERT/DELETE against shared DB state with deterministic IDs.
//
//nolint:paralleltest
func TestDeleteProject_CascadesAndReturnsNotFoundAfter(t *testing.T) {
	const (
		testProjectID = 9001
		testAuthorID  = 9001
		testIssueID   = 9001
	)

	ctx, cancel := context.WithTimeout(t.Context(), dbTimeout)
	defer cancel()

	seedDeleteTestProject(ctx, t, testProjectID, testAuthorID, testIssueID)

	repo := postgres.New(database, metrics.New())

	err := repo.DeleteProject(ctx, testProjectID)
	require.NoError(t, err)

	assertCascadeDeleted(ctx, t, testProjectID, testIssueID)

	// Second delete must return sql.ErrNoRows (not-found).
	err = repo.DeleteProject(ctx, testProjectID)
	require.Error(t, err)
	assert.ErrorIs(t, err, sql.ErrNoRows)
}
