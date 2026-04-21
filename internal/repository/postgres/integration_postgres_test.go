//go:build integration

package postgres_test

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/Go-Yadro-Group-1/Jira-Analyzer/internal/repository"
	"github.com/Go-Yadro-Group-1/Jira-Analyzer/internal/repository/postgres"
	_ "github.com/lib/pq"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	"github.com/stretchr/testify/require"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
)

const (
	dbTimeout  = time.Second * 5
	dbImage    = "postgres:14.5-alpine"
	dbName     = "postgres"
	dbUser     = "postgres"
	dbPassword = "postgres"
)

//nolint:gochecknoglobals
var database *sql.DB

func initPostgres(ctx context.Context) (string, func(), error) {
	container, err := tcpostgres.Run(ctx,
		dbImage,
		tcpostgres.WithDatabase(dbName),
		tcpostgres.WithUsername(dbUser),
		tcpostgres.WithPassword(dbPassword),
		tcpostgres.BasicWaitStrategies(),
	)
	if err != nil {
		return "", nil, fmt.Errorf("start postgres container: %w", err)
	}

	dsn, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		_ = container.Terminate(ctx)

		return "", nil, fmt.Errorf("get connection string: %w", err)
	}

	return dsn, func() { _ = container.Terminate(ctx) }, nil
}

func TestMain(m *testing.M) {
	ctx := context.Background()

	dsn, terminate, err := initPostgres(ctx)
	if err != nil {
		log.Fatalf("init postgres: %v", err)
	}

	database, err = sql.Open("postgres", dsn)
	if err != nil {
		terminate()
		log.Fatalf("connect to db: %v", err)
	}

	runMigrations(dsn, 3)

	code := m.Run()

	database.Close()
	terminate()
	os.Exit(code)
}

func runMigrations(dsn string, targetVersion uint) {
	workDir, err := os.Getwd()
	if err != nil {
		log.Fatalf("failed to get working directory: %s", err)
	}

	migrationsDir := findMigrationsDir(workDir)
	if migrationsDir == "" {
		log.Fatalf("migrations directory not found from %s", workDir)
	}

	migr, err := migrate.New(
		"file://"+filepath.ToSlash(migrationsDir),
		dsn,
	)
	if err != nil {
		log.Fatalf("failed to init migrate: %s", err)
	}

	err = migr.Migrate(targetVersion)
	switch {
	case err == nil:
	case errors.Is(err, migrate.ErrNoChange):
	default:
		log.Fatalf("failed to migrate to version %d: %s", targetVersion, err)
	}

	migr.Close()
}

func findMigrationsDir(startPath string) string {
	dir := startPath
	for range 5 {
		if _, err := os.Stat(filepath.Join(dir, "migrations")); err == nil {
			return filepath.Join(dir, "migrations")
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}

		dir = parent
	}

	return ""
}

func TestGetStatsByProjectBase(t *testing.T) {
	t.Parallel()

	repo := postgres.New(database)

	ctx, cancel := context.WithTimeout(t.Context(), dbTimeout)
	defer cancel()

	want := repository.ProjectStats{
		ProjectID:            1,
		CountTotal:           6,
		CountOpen:            1,
		CountClosed:          3,
		CountResolved:        1,
		CountInProgress:      1,
		CountReopened:        1,
		TotalDurationClosed:  1814400,
		CountCreatedLastWeek: 3,
	}
	have, err := repo.GetStatsByProject(ctx, want.ProjectID)
	require.NoError(t, err)

	require.Equal(t, want, have)
}

func TestGetStatsByProjectEmpty(t *testing.T) {
	t.Parallel()

	repo := postgres.New(database)

	ctx, cancel := context.WithTimeout(t.Context(), dbTimeout)
	defer cancel()

	//nolint:exhaustruct
	want := repository.ProjectStats{}
	have, err := repo.GetStatsByProject(ctx, want.ProjectID)
	require.NoError(t, err)

	require.Equal(t, want, have)
}

func TestGetIssuesDurationByProject(t *testing.T) {
	t.Parallel()

	repo := postgres.New(database)

	ctx, cancel := context.WithTimeout(t.Context(), dbTimeout)
	defer cancel()

	have, err := repo.GetIssuesDurationByProject(ctx, 1)
	require.NoError(t, err)

	require.Len(t, have, 4)

	want := []repository.IssueDuration{
		{IssueID: 1, Duration: 5 * 24 * 3600},
		{IssueID: 2, Duration: 5 * 24 * 3600},
		{IssueID: 3, Duration: 6 * 24 * 3600},
		{IssueID: 6, Duration: 5 * 24 * 3600},
	}

	require.Equal(t, want, have)
}

func TestGetIssuesDurationByProjectEmpty(t *testing.T) {
	t.Parallel()

	repo := postgres.New(database)

	ctx, cancel := context.WithTimeout(t.Context(), dbTimeout)
	defer cancel()

	have, err := repo.GetIssuesDurationByProject(ctx, 999)
	require.NoError(t, err)

	require.Empty(t, have)
}

func TestGetStatusTransitionsByProject(t *testing.T) {
	t.Parallel()

	repo := postgres.New(database)

	ctx, cancel := context.WithTimeout(t.Context(), dbTimeout)
	defer cancel()

	have, err := repo.GetStatusTransitionsByProject(ctx, 1)
	require.NoError(t, err)

	baseTime := time.Now()

	want := []repository.StatusTransition{
		mkTransition(baseTime, 1, "Open", "In Progress"),
		mkTransition(baseTime, 1, "In Progress", "Closed"),
		mkTransition(baseTime, 2, "Open", "In Review"),
		mkTransition(baseTime, 2, "In Review", "Closed"),
		mkTransition(baseTime, 6, "Open", "Closed"),
		mkTransition(baseTime, 6, "Closed", "Open"),
		mkTransition(baseTime, 6, "Open", "Closed"),
		mkTransition(baseTime, 3, "Open", "Resolved"),
	}

	for ind := range have {
		have[ind].ChangeTime = baseTime
	}

	cmp := func(lhs, rhs repository.StatusTransition) int {
		res := strings.Compare(lhs.FromStatus, rhs.FromStatus)
		if res != 0 {
			return res
		}

		return strings.Compare(lhs.ToStatus, rhs.ToStatus)
	}
	slices.SortFunc(have, cmp)
	slices.SortFunc(want, cmp)

	require.Equal(t, want, have)
}

func TestGetStatusTransitionsByProjectEmpty(t *testing.T) {
	t.Parallel()

	repo := postgres.New(database)

	ctx, cancel := context.WithTimeout(t.Context(), dbTimeout)
	defer cancel()

	have, err := repo.GetStatusTransitionsByProject(ctx, 999)
	require.NoError(t, err)
	require.Empty(t, have)
}

func TestGetDailyActivityByProject(t *testing.T) {
	t.Parallel()

	repo := postgres.New(database)

	ctx, cancel := context.WithTimeout(t.Context(), dbTimeout)
	defer cancel()

	have, err := repo.GetDailyActivityByProject(ctx, 1)
	require.NoError(t, err)

	var baseTime time.Time

	want := []repository.DailyActivity{
		mkDailyActivity(baseTime, 0, 2),
		mkDailyActivity(baseTime, 1, 1),
		mkDailyActivity(baseTime, 1, 1),
		mkDailyActivity(baseTime, 1, 0),
		mkDailyActivity(baseTime, 1, 0),
		mkDailyActivity(baseTime, 1, 0),
		mkDailyActivity(baseTime, 1, 0),
	}

	for i := range have {
		have[i].Date = baseTime
	}

	cmp := func(lhs, rhs repository.DailyActivity) int {
		res := lhs.Creation - rhs.Creation
		if res != 0 {
			return res
		}

		return lhs.Completion - rhs.Completion
	}

	slices.SortFunc(have, cmp)
	slices.SortFunc(want, cmp)

	require.Equal(t, want, have)
}

func TestGetDailyActivityByProjectEmpty(t *testing.T) {
	t.Parallel()

	repo := postgres.New(database)

	ctx, cancel := context.WithTimeout(t.Context(), dbTimeout)
	defer cancel()

	have, err := repo.GetDailyActivityByProject(ctx, 999)
	require.NoError(t, err)
	require.Empty(t, have)
}

func TestGetIssuesTimeSpentByProject(t *testing.T) {
	t.Parallel()

	repo := postgres.New(database)

	ctx, cancel := context.WithTimeout(t.Context(), dbTimeout)
	defer cancel()

	have, err := repo.GetIssuesTimeSpentByProject(ctx, 1)
	require.NoError(t, err)

	want := []repository.IssueTimeSpent{
		{IssueID: 1, TimeSpent: 3600},
		{IssueID: 2, TimeSpent: 7200},
		{IssueID: 3, TimeSpent: 1800},
		{IssueID: 6, TimeSpent: 900},
	}

	require.Equal(t, want, have)
}

func TestGetIssuesTimeSpentByProjectEmpty(t *testing.T) {
	t.Parallel()

	repo := postgres.New(database)

	ctx, cancel := context.WithTimeout(t.Context(), dbTimeout)
	defer cancel()

	have, err := repo.GetIssuesTimeSpentByProject(ctx, 999)
	require.NoError(t, err)
	require.Empty(t, have)
}

func TestGetPriorityStatsByProject(t *testing.T) {
	t.Parallel()

	repo := postgres.New(database)

	ctx, cancel := context.WithTimeout(t.Context(), dbTimeout)
	defer cancel()

	have, err := repo.GetPriorityStatsByProject(ctx, 1)
	require.NoError(t, err)

	require.Len(t, have, 4)

	for i := 1; i < len(have); i++ {
		require.GreaterOrEqual(t, have[i].Priority, have[i-1].Priority,
			"priorities must be sorted alphabetically")
	}

	want := []repository.PriorityStats{
		{Priority: "Critical", Count: 1},
		{Priority: "High", Count: 2},
		{Priority: "Low", Count: 1},
		{Priority: "Medium", Count: 2},
	}

	require.Equal(t, want, have)
}

func TestGetPriorityStatsByProjectEmpty(t *testing.T) {
	t.Parallel()

	repo := postgres.New(database)

	ctx, cancel := context.WithTimeout(t.Context(), dbTimeout)
	defer cancel()

	have, err := repo.GetPriorityStatsByProject(ctx, 999)
	require.NoError(t, err)
	require.Empty(t, have)
}

func mkTransition(changeTime time.Time, issueID int, from, to string) repository.StatusTransition {
	return repository.StatusTransition{
		IssueID:    issueID,
		ChangeTime: changeTime,
		FromStatus: from,
		ToStatus:   to,
	}
}

// nolint: unparam
func mkDailyActivity(date time.Time, creation, completion int) repository.DailyActivity {
	return repository.DailyActivity{
		Date:       date,
		Creation:   creation,
		Completion: completion,
	}
}

func TestGetProjectLastUpdated(t *testing.T) {
	t.Parallel()

	repo := postgres.New(database)

	ctx, cancel := context.WithTimeout(t.Context(), dbTimeout)
	defer cancel()

	got, err := repo.GetProjectLastUpdated(ctx, 1)
	require.NoError(t, err)

	// MAX(updated_time) для проекта 1 — NOW() - INTERVAL '1 day' (issue 3 и 6)
	expected := time.Now().Add(-24 * time.Hour)
	require.WithinDuration(t, expected, got, 10*time.Second)
}

func TestGetProjectLastUpdatedEmpty(t *testing.T) {
	t.Parallel()

	repo := postgres.New(database)

	ctx, cancel := context.WithTimeout(t.Context(), dbTimeout)
	defer cancel()

	// Для несуществующего проекта COALESCE вернёт '1970-01-01'
	got, err := repo.GetProjectLastUpdated(ctx, 999)
	require.NoError(t, err)

	require.Equal(t, time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC), got.UTC())
}
