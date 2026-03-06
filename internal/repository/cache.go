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
	DataTypeStats               DataType = "stats"
)

type CacheRepository interface {
	Get(ctx context.Context, projectID int, dataType DataType) ([]byte, error)
	Set(ctx context.Context, projectID int, dataType DataType, value []byte) error
	Invalidate(ctx context.Context, projectID int) error
	GetLastUpdated(ctx context.Context, projectID int) (time.Time, error)
	SetLastUpdated(ctx context.Context, projectID int, t time.Time) error
}
