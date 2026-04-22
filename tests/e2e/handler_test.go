//go:build e2e

package e2e_test

import (
	"encoding/json"
	"testing"

	analyzerv1 "github.com/Go-Yadro-Group-1/Jira-Analyzer/gen/grpc/analyzer/v1"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	seedProjectID  = 1
	emptyProjectID = 999
)

func TestE2E_GetStats_SeededProject(t *testing.T) {
	t.Parallel()

	ctx, cancel := callTimeout(t)
	defer cancel()

	resp, err := client.GetStats(ctx, &analyzerv1.GetStatsRequest{ProjectId: seedProjectID})
	require.NoError(t, err)
	require.NotNil(t, resp.GetStats())

	got := resp.GetStats()
	assert.Equal(t, int32(6), got.GetCountTotal())
	assert.Equal(t, int32(1), got.GetCountOpen())
	assert.Equal(t, int32(3), got.GetCountClosed())
	assert.Equal(t, int32(1), got.GetCountResolved())
	assert.Equal(t, int32(1), got.GetCountInProgress())
	assert.Equal(t, int32(1), got.GetCountReopened())
	assert.Greater(t, got.GetAvgDurationClosed(), float32(0))
	assert.Greater(t, got.GetAvgCreatedPerDayLastWeek(), float32(0))
}

func TestE2E_GetStats_EmptyProject(t *testing.T) {
	t.Parallel()

	ctx, cancel := callTimeout(t)
	defer cancel()

	resp, err := client.GetStats(ctx, &analyzerv1.GetStatsRequest{ProjectId: emptyProjectID})
	require.NoError(t, err)
	require.NotNil(t, resp.GetStats())

	got := resp.GetStats()
	assert.Equal(t, int32(0), got.GetCountTotal())
	assert.InDelta(t, float32(0), got.GetAvgDurationClosed(), 0.001)
	assert.InDelta(t, float32(0), got.GetAvgCreatedPerDayLastWeek(), 0.001)
}

func TestE2E_GetChart_AllTypes(t *testing.T) {
	t.Parallel()

	cases := []analyzerv1.ChartType{
		analyzerv1.ChartType_CHART_TYPE_OPEN_STATE_HISTOGRAM,
		analyzerv1.ChartType_CHART_TYPE_STATE_DISTRIBUTION,
		analyzerv1.ChartType_CHART_TYPE_COMPLEXITY_HISTOGRAM,
		analyzerv1.ChartType_CHART_TYPE_PRIORITY,
		analyzerv1.ChartType_CHART_TYPE_DAILY_ACTIVITY,
	}

	for _, chartType := range cases {
		t.Run(chartType.String(), func(t *testing.T) {
			t.Parallel()

			ctx, cancel := callTimeout(t)
			defer cancel()

			resp, err := client.GetChart(ctx, &analyzerv1.GetChartRequest{
				ProjectId: seedProjectID,
				ChartType: chartType,
			})
			require.NoError(t, err)
			require.NotEmpty(t, resp.GetData())

			var parsed any
			require.NoError(t, json.Unmarshal(resp.GetData(), &parsed))
		})
	}
}

func TestE2E_GetChart_UnspecifiedReturnsInvalidArgument(t *testing.T) {
	t.Parallel()

	ctx, cancel := callTimeout(t)
	defer cancel()

	_, err := client.GetChart(ctx, &analyzerv1.GetChartRequest{
		ProjectId: seedProjectID,
		ChartType: analyzerv1.ChartType_CHART_TYPE_UNSPECIFIED,
	})
	require.Error(t, err)
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
}

func TestE2E_CompareProjects(t *testing.T) {
	t.Parallel()

	ctx, cancel := callTimeout(t)
	defer cancel()

	resp, err := client.CompareProjects(ctx, &analyzerv1.CompareProjectsRequest{
		ProjectIdA: seedProjectID,
		ProjectIdB: emptyProjectID,
	})
	require.NoError(t, err)

	assert.Equal(t, int32(6), resp.GetProjectA().GetCountTotal())
	assert.Equal(t, int32(0), resp.GetProjectB().GetCountTotal())
}
