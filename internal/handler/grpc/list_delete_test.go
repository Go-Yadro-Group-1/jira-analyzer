package grpchandler_test

import (
	"context"
	"testing"

	analyzerv1 "github.com/Go-Yadro-Group-1/Jira-Analyzer/gen/grpc/analyzer/v1"
	"github.com/Go-Yadro-Group-1/Jira-Analyzer/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ListProjects

func TestListProjects_Success(t *testing.T) {
	t.Parallel()

	handler, svc := newHandler(t)

	svc.EXPECT().
		ListProjects(gomock.Any()).
		Return([]service.Project{
			{ID: 1, Title: "Alpha"},
			{ID: 2, Title: "Beta"},
		}, nil)

	resp, err := handler.ListProjects(context.Background(), &analyzerv1.ListProjectsRequest{})

	require.NoError(t, err)
	require.Len(t, resp.GetProjects(), 2)
	assert.Equal(t, int64(1), resp.GetProjects()[0].GetProjectId())
	assert.Equal(t, "Alpha", resp.GetProjects()[0].GetProjectKey())
	assert.Equal(t, int64(2), resp.GetProjects()[1].GetProjectId())
	assert.Equal(t, "Beta", resp.GetProjects()[1].GetProjectKey())
}

func TestListProjects_Empty(t *testing.T) {
	t.Parallel()

	handler, svc := newHandler(t)

	svc.EXPECT().ListProjects(gomock.Any()).Return(nil, nil)

	resp, err := handler.ListProjects(context.Background(), &analyzerv1.ListProjectsRequest{})

	require.NoError(t, err)
	assert.Empty(t, resp.GetProjects())
}

func TestListProjects_ServiceError(t *testing.T) {
	t.Parallel()

	handler, svc := newHandler(t)

	svc.EXPECT().ListProjects(gomock.Any()).Return(nil, errService)

	_, err := handler.ListProjects(context.Background(), &analyzerv1.ListProjectsRequest{})

	require.Error(t, err)
	assert.Equal(t, codes.Internal, status.Code(err))
}

// DeleteProject

func TestDeleteProject_Success(t *testing.T) {
	t.Parallel()

	handler, svc := newHandler(t)

	svc.EXPECT().DeleteProject(gomock.Any(), 7).Return(nil)

	resp, err := handler.DeleteProject(
		context.Background(),
		&analyzerv1.DeleteProjectRequest{ProjectId: 7},
	)

	require.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestDeleteProject_NotFound(t *testing.T) {
	t.Parallel()

	handler, svc := newHandler(t)

	svc.EXPECT().DeleteProject(gomock.Any(), 99).Return(service.ErrProjectNotFound)

	_, err := handler.DeleteProject(
		context.Background(),
		&analyzerv1.DeleteProjectRequest{ProjectId: 99},
	)

	require.Error(t, err)
	assert.Equal(t, codes.NotFound, status.Code(err))
}

func TestDeleteProject_ServiceError(t *testing.T) {
	t.Parallel()

	handler, svc := newHandler(t)

	svc.EXPECT().DeleteProject(gomock.Any(), 1).Return(errService)

	_, err := handler.DeleteProject(
		context.Background(),
		&analyzerv1.DeleteProjectRequest{ProjectId: 1},
	)

	require.Error(t, err)
	assert.Equal(t, codes.Internal, status.Code(err))
}
