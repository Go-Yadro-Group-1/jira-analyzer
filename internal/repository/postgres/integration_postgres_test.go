//go:build integration

package postgres_test

import (
	"database/sql"
	"log"
	"os"
	"testing"
	"time"

	"github.com/Go-Yadro-Group-1/Jira-Analyzer/internal/repository/postgres"
	"github.com/stretchr/testify/require"
)

const (
	dbTimeout = time.Second * 5
)

var repo *postgres.Postgres

func TestMain(m *testing.M) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("connect to db: %v", err)
	}
	defer db.Close()
	repo = postgres.New(db)
	os.Exit(m.Run())
}

func seedDB(t *testing.T, db *sql.DB) {
	t.Helper()

	_, err := db.Exec(`INSERT INTO raw.project (id, title) VALUES (1, 'Test Project')`)
	require.NoError(t, err)

	_, err = db.Exec(`INSERT INTO raw.author (id, name) VALUES (1, 'Alice'), (2, 'Bob')`)
	require.NoError(t, err)

	_, err = db.Exec(`
		INSERT INTO raw.issue (id, project_id, author_id, assignee_id, key, summary, type, priority, status, created_time, closed_time, updated_time, time_spent) VALUES
			(1, 1, 1, 2, 'TP-1', 'Fix bug',        'Bug',  'High',     'Closed',      NOW() - INTERVAL '10 days', NOW() - INTERVAL '5 days',   NOW() - INTERVAL '5 days',   3600),
			(2, 1, 1, 2, 'TP-2', 'Add feature',    'Task', 'High',     'Closed',      NOW() - INTERVAL '8 days',  NOW() - INTERVAL '3 days',   NOW() - INTERVAL '3 days',   7200),
			(3, 1, 1, 1, 'TP-3', 'Refactor code',  'Task', 'Critical', 'Resolved',    NOW() - INTERVAL '7 days',  NOW() - INTERVAL '1 day',    NOW() - INTERVAL '1 day',    1800),
			(4, 1, 2, NULL,'TP-4', 'Write tests',  'Task', 'Low',      'Open',        NOW() - INTERVAL '5 days',  NULL,                        NOW() - INTERVAL '5 days',   NULL),
			(5, 1, 2, 2, 'TP-5', 'Deploy',         'Task', 'Medium',   'In Progress', NOW() - INTERVAL '3 days',  NULL,                        NOW() - INTERVAL '3 days',   NULL),
			(6, 1, 1, 2, 'TP-6', 'Regression fix', 'Bug',  'Medium',   'Closed',      NOW() - INTERVAL '6 days',  NOW() - INTERVAL '12 hours', NOW() - INTERVAL '12 hours', 900)
	`)
	require.NoError(t, err)

	_, err = db.Exec(`
		INSERT INTO raw.status_changes (issue_id, author_id, change_time, from_status, to_status) VALUES
			(1, 1, NOW() - INTERVAL '9 days',   'Open',        'In Progress'),
			(1, 1, NOW() - INTERVAL '5 days',   'In Progress', 'Closed'),
			(2, 1, NOW() - INTERVAL '7 days',   'Open',        'In Review'),
			(2, 1, NOW() - INTERVAL '3 days',   'In Review',   'Closed'),
			(3, 1, NOW() - INTERVAL '6 days',   'Open',        'Resolved'),
			(6, 1, NOW() - INTERVAL '5 days',   'Open',        'Closed'),
			(6, 1, NOW() - INTERVAL '2 days',   'Closed',      'Open'),
			(6, 1, NOW() - INTERVAL '12 hours', 'Open',        'Closed')
	`)
	require.NoError(t, err)
}

func cleanDB(t *testing.T, db *sql.DB) {
	t.Helper()
	_, err := db.Exec(`
		TRUNCATE raw.status_changes, raw.issue, raw.author, raw.project RESTART IDENTITY;
	`)
	require.NoError(t, err)
}
