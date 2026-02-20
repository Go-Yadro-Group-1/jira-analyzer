//go:build integration

package postgres_test

import (
	"context"
	"database/sql"
	_ "embed"
	"log"
	"os"
	"testing"
	"time"

	"github.com/Go-Yadro-Group-1/Jira-Analyzer/internal/repository/postgres"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

//go:embed testdata/seed.sql
var seedSQL string

//go:embed testdata/clean.sql
var cleanSQL string

const (
	dbTimeout = time.Second * 5
)

//nolint:gochecknoglobals
var database *sql.DB

func TestMain(m *testing.M) {
	dsn := os.Getenv("TEST_DATABASE_URL")

	database, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("connect to db: %v", err)
	}

	ctx := context.Background()

	if err = database.PingContext(ctx); err != nil {
		log.Fatalf("ping db: %v", err)
	}

	cleanDB(database)
	seedDB(database)

	code := m.Run()

	cleanDB(database)
	database.Close()
	os.Exit(code)
}

func TestGetStatsByProjectBase(t *testing.T) {
	t.Parallel()

	repo := postgres.New(database)

	ctx, cancel := context.WithTimeout(t.Context(), dbTimeout)
	defer cancel()

	want := postgres.ProjectStats{
		ProjectID:        1,
		Total:            6,
		Open:             1,
		Closed:           3,
		Resolved:         1,
		InProgress:       1,
		Reopened:         1,
		AvgDurationHours: time.Hour * 131,
	}
	have, err := repo.GetStatsByProject(ctx, want.ProjectID)
	require.NoError(t, err)

	haveAvgDailyLastWeek := have.AvgDailyLastWeek
	have.AvgDailyLastWeek = 0.0
	wantAvgDurationHours := 5.0 / 7.0

	require.Equal(t, want, have)
	require.InDelta(t, wantAvgDurationHours, haveAvgDailyLastWeek, 0.01)
}

func TestGetStatsByProjectEmpty(t *testing.T) {
	t.Parallel()

	repo := postgres.New(database)

	ctx, cancel := context.WithTimeout(t.Context(), dbTimeout)
	defer cancel()

	want := postgres.ProjectStats{}
	have, err := repo.GetStatsByProject(ctx, want.ProjectID)
	require.NoError(t, err)

	require.Equal(t, want, have)
}

func seedDB(conn *sql.DB) {
	ctx := context.Background()

	if _, err := conn.ExecContext(ctx, seedSQL); err != nil {
		log.Fatalf("seed db: %v", err)
	}
}

func cleanDB(conn *sql.DB) {
	ctx := context.Background()

	if _, err := conn.ExecContext(ctx, cleanSQL); err != nil {
		log.Fatalf("clean db: %v", err)
	}
}
