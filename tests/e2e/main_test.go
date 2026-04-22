//go:build e2e

package e2e_test

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"

	analyzerv1 "github.com/Go-Yadro-Group-1/Jira-Analyzer/gen/grpc/analyzer/v1"
	"github.com/Go-Yadro-Group-1/Jira-Analyzer/internal/app"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"

	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	dbImage    = "postgres:14.5-alpine"
	dbName     = "postgres"
	dbUser     = "postgres"
	dbPassword = "postgres"

	migrationsTargetVersion = 3
)

//nolint:gochecknoglobals
var (
	client                analyzerv1.AnalyzerServiceClient
	errMigrationsNotFound = errors.New("migrations directory not found")
)

func TestMain(m *testing.M) {
	ctx := context.Background()

	dsn, terminateContainer, err := startPostgres(ctx)
	if err != nil {
		log.Fatalf("start postgres: %v", err)
	}

	database, err := sql.Open("postgres", dsn)
	if err != nil {
		terminateContainer()
		log.Fatalf("open db: %v", err)
	}

	err = runMigrations(dsn)
	if err != nil {
		database.Close()
		terminateContainer()
		log.Fatalf("migrate: %v", err)
	}

	server := app.NewGRPCServer(database)

	var lc net.ListenConfig

	listener, err := lc.Listen(ctx, "tcp", "127.0.0.1:0")
	if err != nil {
		database.Close()
		terminateContainer()
		log.Fatalf("listen: %v", err)
	}

	go func() {
		err := server.Serve(listener)
		if err != nil {
			log.Printf("grpc serve exited: %v", err)
		}
	}()

	conn, err := grpc.NewClient(
		listener.Addr().String(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		server.GracefulStop()
		database.Close()
		terminateContainer()
		log.Fatalf("grpc dial: %v", err)
	}

	client = analyzerv1.NewAnalyzerServiceClient(conn)

	code := m.Run()

	_ = conn.Close()

	server.GracefulStop()
	database.Close()
	terminateContainer()

	os.Exit(code)
}

func startPostgres(ctx context.Context) (string, func(), error) {
	container, err := tcpostgres.Run(ctx,
		dbImage,
		tcpostgres.WithDatabase(dbName),
		tcpostgres.WithUsername(dbUser),
		tcpostgres.WithPassword(dbPassword),
		tcpostgres.BasicWaitStrategies(),
	)
	if err != nil {
		return "", nil, fmt.Errorf("run container: %w", err)
	}

	dsn, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		_ = container.Terminate(ctx)

		return "", nil, fmt.Errorf("connection string: %w", err)
	}

	return dsn, func() { _ = container.Terminate(ctx) }, nil
}

func runMigrations(dsn string) error {
	migrationsDir, err := findMigrationsDir()
	if err != nil {
		return err
	}

	migr, err := migrate.New("file://"+filepath.ToSlash(migrationsDir), dsn)
	if err != nil {
		return fmt.Errorf("init migrate: %w", err)
	}
	defer migr.Close()

	err = migr.Migrate(migrationsTargetVersion)
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("migrate to %d: %w", migrationsTargetVersion, err)
	}

	return nil
}

func findMigrationsDir() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("getwd: %w", err)
	}

	for range 5 {
		candidate := filepath.Join(dir, "migrations")
		if _, err := os.Stat(candidate); err == nil {
			return candidate, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}

		dir = parent
	}

	return "", errMigrationsNotFound
}

//nolint:unparam
func callTimeout(t *testing.T) (context.Context, context.CancelFunc) {
	t.Helper()

	return context.WithTimeout(t.Context(), 5*time.Second)
}
