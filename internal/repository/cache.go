package repository

import (
	"context"
	"time"
)

type CacheRepository interface {
	Get(ctx context.Context, projectID string, dataType string) (string, error)
	Set(ctx context.Context, projectID string, dataType string, value string) error
	Invalidate(ctx context.Context, projectID string) error
	GetLastUpdated(ctx context.Context, projectID string) (time.Time, error)
	SetLastUpdated(ctx context.Context, projectID string, t time.Time) error
}
