package memory_test

import (
	"context"
	"testing"
	"time"

	"github.com/Go-Yadro-Group-1/Jira-Analyzer/internal/repository/memory"

	"github.com/stretchr/testify/require"
)

func newRepo() *memory.CacheRepository[int, string] {
	return memory.NewCacheRepository[int, string]()
}

func TestGet_ProjectNotFound(t *testing.T) {
	t.Parallel()

	repo := newRepo()

	_, err := repo.Get(context.Background(), 1, "histogram")
	require.ErrorIs(t, err, memory.ErrProjectNotFound)
}

func TestGet_DataTypeNotFound(t *testing.T) {
	t.Parallel()

	repo := newRepo()

	require.NoError(t, repo.Set(context.Background(), 1, "histogram", []byte("data")))

	_, err := repo.Get(context.Background(), 1, "unknown")
	require.ErrorIs(t, err, memory.ErrDataTypeNotFound)
}

func TestSet_Get(t *testing.T) {
	t.Parallel()

	repo := newRepo()
	want := []byte("payload")

	require.NoError(t, repo.Set(context.Background(), 1, "histogram", want))

	got, err := repo.Get(context.Background(), 1, "histogram")
	require.NoError(t, err)
	require.Equal(t, want, got)
}

func TestSet_Overwrite(t *testing.T) {
	t.Parallel()

	repo := newRepo()

	require.NoError(t, repo.Set(context.Background(), 1, "histogram", []byte("old")))
	require.NoError(t, repo.Set(context.Background(), 1, "histogram", []byte("new")))

	got, err := repo.Get(context.Background(), 1, "histogram")
	require.NoError(t, err)
	require.Equal(t, []byte("new"), got)
}

func TestInvalidate(t *testing.T) {
	t.Parallel()

	repo := newRepo()

	require.NoError(t, repo.Set(context.Background(), 1, "histogram", []byte("data")))
	require.NoError(t, repo.Invalidate(context.Background(), 1))

	_, err := repo.Get(context.Background(), 1, "histogram")
	require.ErrorIs(t, err, memory.ErrProjectNotFound)
}

func TestInvalidate_ClearsUpdatedAt(t *testing.T) {
	t.Parallel()

	repo := newRepo()

	require.NoError(t, repo.SetLastUpdated(context.Background(), 1, time.Now()))
	require.NoError(t, repo.Invalidate(context.Background(), 1))

	_, err := repo.GetLastUpdated(context.Background(), 1)
	require.ErrorIs(t, err, memory.ErrProjectNotFound)
}

func TestSetLastUpdated_GetLastUpdated(t *testing.T) {
	t.Parallel()

	repo := newRepo()
	want := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)

	require.NoError(t, repo.SetLastUpdated(context.Background(), 1, want))

	got, err := repo.GetLastUpdated(context.Background(), 1)
	require.NoError(t, err)
	require.True(t, got.Equal(want))
}

func TestGetLastUpdated_ProjectNotFound(t *testing.T) {
	t.Parallel()

	repo := newRepo()

	_, err := repo.GetLastUpdated(context.Background(), 99)
	require.ErrorIs(t, err, memory.ErrProjectNotFound)
}
