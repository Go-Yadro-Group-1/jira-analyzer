package metrics

import (
	grpcprom "github.com/grpc-ecosystem/go-grpc-middleware/providers/prometheus"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
)

const namespace = "analyzer"

type Metrics struct {
	Registry        *prometheus.Registry
	CacheHits       *prometheus.CounterVec
	CacheMisses     *prometheus.CounterVec
	DBQueryDuration *prometheus.HistogramVec
	GRPCServer      *grpcprom.ServerMetrics
}

func New() *Metrics {
	cacheHits := prometheus.NewCounterVec(
		prometheus.CounterOpts{ //nolint:exhaustruct
			Namespace: namespace,
			Name:      "cache_hits_total",
			Help:      "Number of cache hits in the service layer, by data type.",
		},
		[]string{"data_type"},
	)

	cacheMisses := prometheus.NewCounterVec(
		prometheus.CounterOpts{ //nolint:exhaustruct
			Namespace: namespace,
			Name:      "cache_misses_total",
			Help:      "Number of cache misses in the service layer, by data type.",
		},
		[]string{"data_type"},
	)

	dbDuration := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{ //nolint:exhaustruct
			Namespace: namespace,
			Name:      "db_query_duration_seconds",
			Help:      "Duration of database queries by query name.",
			Buckets:   prometheus.DefBuckets,
		},
		[]string{"query"},
	)

	grpcServer := grpcprom.NewServerMetrics(
		grpcprom.WithServerHandlingTimeHistogram(),
	)

	reg := prometheus.NewRegistry()
	reg.MustRegister(cacheHits, cacheMisses, dbDuration, grpcServer)

	return &Metrics{
		Registry:        reg,
		CacheHits:       cacheHits,
		CacheMisses:     cacheMisses,
		DBQueryDuration: dbDuration,
		GRPCServer:      grpcServer,
	}
}

func (m *Metrics) RegisterRuntimeCollectors() {
	m.Registry.MustRegister(
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}), //nolint:exhaustruct
	)
}
