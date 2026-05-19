package metrics_test

import (
	"testing"

	"github.com/Go-Yadro-Group-1/Jira-Analyzer/internal/metrics"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew_RegistersCustomMetrics(t *testing.T) {
	t.Parallel()

	mtr := metrics.New()

	mtr.CacheHits.WithLabelValues("stats").Inc()
	mtr.CacheMisses.WithLabelValues("stats").Inc()

	assert.InDelta(t, 1.0, testutil.ToFloat64(mtr.CacheHits.WithLabelValues("stats")), 0.001)
	assert.InDelta(t, 1.0, testutil.ToFloat64(mtr.CacheMisses.WithLabelValues("stats")), 0.001)
}

func TestNew_DBQueryDurationObservesValues(t *testing.T) {
	t.Parallel()

	mtr := metrics.New()

	timer := prometheus.NewTimer(mtr.DBQueryDuration.WithLabelValues("test_query"))
	timer.ObserveDuration()

	got, err := mtr.Registry.Gather()
	require.NoError(t, err)

	found := false

	for _, mf := range got {
		if mf.GetName() == "analyzer_db_query_duration_seconds" {
			found = true

			break
		}
	}

	assert.True(t, found, "db_query_duration_seconds should be registered")
}

func TestNew_FreshRegistryPerCall(t *testing.T) {
	t.Parallel()

	mtr1 := metrics.New()
	mtr2 := metrics.New()

	assert.NotSame(t, mtr1.Registry, mtr2.Registry)
}
