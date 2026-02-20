//go:build integration

package postgres_test

import (
	"database/sql"
	"log"
	"os"
	"testing"

	"github.com/Go-Yadro-Group-1/Jira-Analyzer/internal/repository/postgres"
)

var repo *postgres.Postgres

func TestMain(m *testing.M) {
	dsn := os.Getenv("TEST_DATABASE_URL")

	database, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("connect to db: %v", err)
	}
	defer database.Close()

	repo = postgres.New(database)

	os.Exit(m.Run())
}
