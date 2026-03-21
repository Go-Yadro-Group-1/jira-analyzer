package service

import (
	"context"
	"encoding/json"
	"time"

	"github.com/Go-Yadro-Group-1/Jira-Analyzer/internal/repository"
	"golang.org/x/sync/errgroup"
)

type Repository interface {
	GetProjectLastUpdated(ctx context.Context, projectID int) (time.Time, error)
	GetStatsByProject(ctx context.Context, projectID int) (repository.ProjectStats, error)
	GetIssuesDurationByProject(ctx context.Context, projectID int) ([]repository.IssueDuration, error)
	GetStatusTransitionsByProject(ctx context.Context, projectID int) ([]repository.StatusTransition, error)
	GetDailyActivityByProject(ctx context.Context, projectID int) ([]repository.DailyActivity, error)
	GetIssuesTimeSpentByProject(ctx context.Context, projectID int) ([]repository.IssueTimeSpent, error)
	GetPriorityStatsByProject(ctx context.Context, projectID int) ([]repository.PriorityStats, error)
}

type Cache interface {
	Get(ctx context.Context, projectID int, dataType string) ([]byte, error)
	Set(ctx context.Context, projectID int, dataType string, value []byte) error
	Invalidate(ctx context.Context, projectID int) error
	GetLastUpdated(ctx context.Context, projectID int) (time.Time, error)
	SetLastUpdated(ctx context.Context, projectID int, t time.Time) error
}

type Service struct {
	repository Repository
	cache      Cache
}

func New(repository Repository, cache Cache) *Service {
	return &Service{
		repository: repository,
		cache:      cache,
	}
}

func (s *Service) isCacheStale(ctx context.Context, projectID int) (bool, error) {
	dbUpdatedAt, err := s.repository.GetProjectLastUpdated(ctx, projectID)
	if err != nil {
		return false, err
	}

	cacheUpdatedAt, err := s.cache.GetLastUpdated(ctx, projectID)
	if err != nil {
		return true, nil
	}

	return dbUpdatedAt.After(cacheUpdatedAt), nil
}

func fetchWithCache[T any](
	ctx context.Context,
	s *Service,
	projectID int,
	dataType string,
	fetch func(context.Context, int) (T, error),
) (T, error) {
	var zero T

	stale, err := s.isCacheStale(ctx, projectID)
	if err != nil {
		return zero, err
	}

	if stale {
		if err := s.cache.Invalidate(ctx, projectID); err != nil {
			return zero, err
		}
	} else {
		if cached, err := s.cache.Get(ctx, projectID, dataType); err == nil {
			var result T
			if err := json.Unmarshal(cached, &result); err == nil {
				return result, nil
			}
		}
	}

	result, err := fetch(ctx, projectID)
	if err != nil {
		return zero, err
	}

	if data, err := json.Marshal(result); err == nil {
		_ = s.cache.Set(ctx, projectID, dataType, data)
		_ = s.cache.SetLastUpdated(ctx, projectID, time.Now())
	}

	return result, nil
}

const dataTypeStats = "stats"

func (s *Service) GetProjectStat(ctx context.Context, projectID int) (repository.ProjectStats, error) {
	return fetchWithCache(ctx, s, projectID, dataTypeStats, s.repository.GetStatsByProject)
}

const dataTypeIssuesDuration = "issues_duration"

func (s *Service) GetIssuesDurationHistogram(ctx context.Context, projectID int) (IssuesDurationHistogram, error) {
	return fetchWithCache(ctx, s, projectID, dataTypeIssuesDuration, func(ctx context.Context, projectID int) (IssuesDurationHistogram, error) {
		rows, err := s.repository.GetIssuesDurationByProject(ctx, projectID)
		if err != nil {
			return IssuesDurationHistogram{}, err
		}

		return buildIssuesDurationHistogram(rows)
	})
}

const dataTypeStatusTransitions = "status_transitions"

func (s *Service) GetStatusHistograms(ctx context.Context, projectID int) ([]StatusHistogram, error) {
	return fetchWithCache(ctx, s, projectID, dataTypeStatusTransitions, func(ctx context.Context, projectID int) ([]StatusHistogram, error) {
		rows, err := s.repository.GetStatusTransitionsByProject(ctx, projectID)
		if err != nil {
			return nil, err
		}

		return buildStatusHistograms(rows)
	})
}

const dataTypeDailyActivity = "daily_activity"

func (s *Service) GetDailyActivityChart(ctx context.Context, projectID int) (DailyActivityChart, error) {
	return fetchWithCache(ctx, s, projectID, dataTypeDailyActivity, func(ctx context.Context, projectID int) (DailyActivityChart, error) {
		rows, err := s.repository.GetDailyActivityByProject(ctx, projectID)
		if err != nil {
			return DailyActivityChart{}, err
		}

		return buildDailyActivityChart(rows)
	})
}

const dataTypeIssuesTimeSpent = "issues_time_spent"

func (s *Service) GetIssuesTimeSpentHistogram(ctx context.Context, projectID int) (IssuesTimeSpentHistogram, error) {
	return fetchWithCache(ctx, s, projectID, dataTypeIssuesTimeSpent, func(ctx context.Context, projectID int) (IssuesTimeSpentHistogram, error) {
		rows, err := s.repository.GetIssuesTimeSpentByProject(ctx, projectID)
		if err != nil {
			return IssuesTimeSpentHistogram{}, err
		}

		return buildIssuesTimeSpentHistogram(rows)
	})
}

const dataTypePriorityStats = "priority_stats"

func (s *Service) GetPriorityChart(ctx context.Context, projectID int) (PriorityChart, error) {
	return fetchWithCache(ctx, s, projectID, dataTypePriorityStats, func(ctx context.Context, projectID int) (PriorityChart, error) {
		rows, err := s.repository.GetPriorityStatsByProject(ctx, projectID)
		if err != nil {
			return PriorityChart{}, err
		}

		return buildPriorityChart(rows), nil
	})
}

func (s *Service) CompareTwoProjects(ctx context.Context, lhsProjectID, rhsProjectID int) ([2]repository.ProjectStats, error) {
	var result [2]repository.ProjectStats

	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		var err error
		result[0], err = s.GetProjectStat(ctx, lhsProjectID)
		return err
	})

	g.Go(func() error {
		var err error
		result[1], err = s.GetProjectStat(ctx, rhsProjectID)
		return err
	})

	if err := g.Wait(); err != nil {
		return [2]repository.ProjectStats{}, err
	}

	return result, nil
}
