//go:build e2e

package e2e_test

import (
	"context"
	"testing"

	analyzerv1 "github.com/Go-Yadro-Group-1/Jira-Analyzer/gen/grpc/analyzer/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestE2E_ListProjects_ContainsSeedProject(t *testing.T) {
	t.Parallel()

	ctx, cancel := callTimeout(t)
	defer cancel()

	resp, err := client.ListProjects(ctx, &analyzerv1.ListProjectsRequest{})
	require.NoError(t, err)
	require.NotNil(t, resp)

	found := false

	for _, p := range resp.GetProjects() {
		if p.GetProjectId() == seedProjectID {
			assert.Equal(t, "Test Project", p.GetProjectKey())

			found = true
		}
	}

	assert.True(t, found, "seed project (id=%d) must appear in ListProjects", seedProjectID)
}

func TestE2E_DeleteProject_NotFound(t *testing.T) {
	t.Parallel()

	ctx, cancel := callTimeout(t)
	defer cancel()

	_, err := client.DeleteProject(ctx, &analyzerv1.DeleteProjectRequest{ProjectId: 999999})
	require.Error(t, err)
	assert.Equal(t, codes.NotFound, status.Code(err))
}

// TestE2E_DeleteProject_RemovesProjectFromList inserts an isolated project, verifies it appears
// in ListProjects, deletes it via the gRPC API, and verifies it is gone.
// A second delete must return NotFound.
//
// Not parallel — writes to the shared DB outside the gRPC API (direct SQL insert).
//
//nolint:paralleltest
func TestE2E_DeleteProject_RemovesProjectFromList(t *testing.T) {
	const (
		ephemeralProjectID = 8001
		ephemeralAuthorID  = 8001
	)

	ctx, cancel := callTimeout(t)
	defer cancel()

	// Insert a throwaway project directly via DB so we don't touch the seed data.
	_, err := sharedDB.ExecContext(ctx,
		`INSERT INTO raw.project (id, title) VALUES ($1, $2) ON CONFLICT DO NOTHING`,
		ephemeralProjectID, "Ephemeral E2E Project",
	)
	require.NoError(t, err)

	t.Cleanup(func() {
		// Best-effort cleanup: project may already be deleted by the test.
		_, _ = sharedDB.ExecContext(
			context.Background(),
			`DELETE FROM raw.project WHERE id = $1`,
			ephemeralProjectID,
		)
	})

	// Verify the project appears in ListProjects.
	listResp, err := client.ListProjects(ctx, &analyzerv1.ListProjectsRequest{})
	require.NoError(t, err)

	found := false

	for _, p := range listResp.GetProjects() {
		if p.GetProjectId() == ephemeralProjectID {
			found = true
		}
	}

	require.True(t, found, "ephemeral project must appear in ListProjects before deletion")

	// Delete the project via gRPC.
	_, err = client.DeleteProject(
		ctx,
		&analyzerv1.DeleteProjectRequest{ProjectId: ephemeralProjectID},
	)
	require.NoError(t, err)

	// After deletion: project must not appear in list.
	listResp2, err := client.ListProjects(ctx, &analyzerv1.ListProjectsRequest{})
	require.NoError(t, err)

	for _, p := range listResp2.GetProjects() {
		assert.NotEqual(t, int64(ephemeralProjectID), p.GetProjectId(),
			"deleted project must not appear in list")
	}

	// Idempotency: second delete must return NotFound.
	_, err = client.DeleteProject(
		ctx,
		&analyzerv1.DeleteProjectRequest{ProjectId: ephemeralProjectID},
	)
	require.Error(t, err)
	assert.Equal(t, codes.NotFound, status.Code(err))
}
