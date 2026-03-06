package repository

import (
	"context"
	"time"
)

type DataType string

const (
	DataTypeOpenTimeHistogram   DataType = "open_time_histogram"
	DataTypeStateTimeHistogram  DataType = "state_time_histogram"
	DataTypeActivityChart       DataType = "activity_chart"
	DataTypeComplexityHistogram DataType = "complexity_histogram"
	DataTypePriorityChart       DataType = "priority_chart"
)

type CacheRepository interface {
	Get(ctx context.Context, projectID string, dataType DataType) (string, error)
	Set(ctx context.Context, projectID string, dataType DataType, value string) error
	Invalidate(ctx context.Context, projectID string) error
	GetLastUpdated(ctx context.Context, projectID string) (time.Time, error)
	SetLastUpdated(ctx context.Context, projectID string, t time.Time) error
}
