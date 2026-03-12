package repository

import (
	"context"
)

type AnalyticsRepository interface {
	GetStatsByProject(ctx context.Context, projectID int) (ProjectStats, error)
	GetIssuesDurationByProject(ctx context.Context, projectID int) ([]IssueDuration, error)
	GetStatusTransitionsByProject(ctx context.Context, projectID int) ([]StatusTransition, error)
	GetDailyActivityByProject(ctx context.Context, projectID int) ([]DailyActivity, error)
	GetIssuesTimeSpentByProject(ctx context.Context, projectID int) ([]IssueTimeSpent, error)
	GetPriorityStatsByProject(ctx context.Context, projectID int) ([]PriorityStats, error)
}
