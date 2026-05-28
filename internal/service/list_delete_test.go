package service_test

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	"github.com/Go-Yadro-Group-1/Jira-Analyzer/internal/metrics"
	"github.com/Go-Yadro-Group-1/Jira-Analyzer/internal/repository"
	"github.com/Go-Yadro-Group-1/Jira-Analyzer/internal/service"
	"github.com/Go-Yadro-Group-1/Jira-Analyzer/internal/service/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

// TestListProjects_ReturnsAllProjects verifies that ListProjects delegates to the repository
// and maps the result to service domain models unchanged.
func TestListProjects_ReturnsAllProjects(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	repo := mocks.NewMockRepository(ctrl)
	cache := mocks.NewMockCache(ctrl)

	repoProjects := []repository.Project{
		{ID: 1, Title: "Alpha"},
		{ID: 2, Title: "Beta"},
	}
	repo.EXPECT().ListProjects(gomock.Any()).Return(repoProjects, nil)

	svc := service.New(repo, cache, metrics.New())

	got, err := svc.ListProjects(context.Background())

	require.NoError(t, err)
	require.Len(t, got, 2)
	assert.Equal(t, 1, got[0].ID)
	assert.Equal(t, "Alpha", got[0].Title)
	assert.Equal(t, 2, got[1].ID)
	assert.Equal(t, "Beta", got[1].Title)
}

// TestListProjects_EmptyResult verifies that an empty project list is returned as an empty slice, not an error.
func TestListProjects_EmptyResult(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	repo := mocks.NewMockRepository(ctrl)
	cache := mocks.NewMockCache(ctrl)

	repo.EXPECT().ListProjects(gomock.Any()).Return(nil, nil)

	svc := service.New(repo, cache, metrics.New())

	got, err := svc.ListProjects(context.Background())

	require.NoError(t, err)
	assert.Empty(t, got)
}

// TestListProjects_RepoError verifies that repository errors are propagated.
func TestListProjects_RepoError(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	repo := mocks.NewMockRepository(ctrl)
	cache := mocks.NewMockCache(ctrl)

	repo.EXPECT().ListProjects(gomock.Any()).Return(nil, errDB)

	svc := service.New(repo, cache, metrics.New())

	_, err := svc.ListProjects(context.Background())

	require.Error(t, err)
}

// TestDeleteProject_InvalidatesCache verifies that after a successful delete,
// the cache entry for the project is invalidated.
func TestDeleteProject_InvalidatesCache(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	repo := mocks.NewMockRepository(ctrl)
	cache := mocks.NewMockCache(ctrl)

	const projectID = 42

	repo.EXPECT().DeleteProject(gomock.Any(), projectID).Return(nil)
	cache.EXPECT().Invalidate(gomock.Any(), projectID).Return(nil)

	svc := service.New(repo, cache, metrics.New())

	err := svc.DeleteProject(context.Background(), projectID)

	require.NoError(t, err)
}

// TestDeleteProject_NotFound verifies that sql.ErrNoRows from the repository
// is translated to service.ErrProjectNotFound.
func TestDeleteProject_NotFound(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	repo := mocks.NewMockRepository(ctrl)
	cache := mocks.NewMockCache(ctrl)

	const projectID = 99

	// Simulate the ProjectQueryError wrapping sql.ErrNoRows, as returned by the postgres repo.
	notFoundErr := fmt.Errorf("delete project: %w", sql.ErrNoRows)
	repo.EXPECT().DeleteProject(gomock.Any(), projectID).Return(notFoundErr)

	svc := service.New(repo, cache, metrics.New())

	err := svc.DeleteProject(context.Background(), projectID)

	require.ErrorIs(t, err, service.ErrProjectNotFound)
}

// TestDeleteProject_RepoError verifies that non-NotFound repository errors are propagated as-is.
func TestDeleteProject_RepoError(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	repo := mocks.NewMockRepository(ctrl)
	cache := mocks.NewMockCache(ctrl)

	repo.EXPECT().DeleteProject(gomock.Any(), 1).Return(errDB)

	svc := service.New(repo, cache, metrics.New())

	err := svc.DeleteProject(context.Background(), 1)

	require.Error(t, err)
	require.NotErrorIs(t, err, service.ErrProjectNotFound)
}
