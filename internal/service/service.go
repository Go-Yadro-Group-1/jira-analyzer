package service

//go:generate mockgen -destination=mocks/mock_repository.go -package=mocks github.com/Go-Yadro-Group-1/Jira-Analyzer/internal/service Repository
//go:generate mockgen -destination=mocks/mock_cache.go -package=mocks github.com/Go-Yadro-Group-1/Jira-Analyzer/internal/service Cache

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/Go-Yadro-Group-1/Jira-Analyzer/internal/repository"
	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/singleflight"
)

type Repository interface {
	GetProjectLastUpdated(ctx context.Context, projectID int) (time.Time, error)
	GetStatsByProject(ctx context.Context, projectID int) (repository.ProjectStats, error)
	GetIssuesDurationByProject(
		ctx context.Context,
		projectID int,
	) ([]repository.IssueDuration, error)
	GetStatusTransitionsByProject(
		ctx context.Context,
		projectID int,
	) ([]repository.StatusTransition, error)
	GetDailyActivityByProject(
		ctx context.Context,
		projectID int,
	) ([]repository.DailyActivity, error)
	GetIssuesTimeSpentByProject(
		ctx context.Context,
		projectID int,
	) ([]repository.IssueTimeSpent, error)
	GetPriorityStatsByProject(
		ctx context.Context,
		projectID int,
	) ([]repository.PriorityStats, error)
}

type Cache interface {
	Get(ctx context.Context, projectID int, dataType string) ([]byte, error)
	Set(ctx context.Context, projectID int, dataType string, value []byte) error
	Invalidate(ctx context.Context, projectID int) error
	GetLastUpdated(ctx context.Context, projectID int) (time.Time, error)
	SetLastUpdated(ctx context.Context, projectID int, t time.Time) error
}

const staleCheckTTL = 30 * time.Second

type Service struct {
	repository    Repository
	cache         Cache
	sfGroup       singleflight.Group
	lastCheckedMu sync.RWMutex
	lastCheckedAt map[int]time.Time
}

func New(repository Repository, cache Cache) *Service {
	return &Service{
		repository:    repository,
		cache:         cache,
		sfGroup:       singleflight.Group{},
		lastCheckedMu: sync.RWMutex{},
		lastCheckedAt: make(map[int]time.Time),
	}
}

const dataTypeStats = "stats"

func (s *Service) GetProjectStat(
	ctx context.Context,
	projectID int,
) (ProjectStats, error) {
	return fetchWithCache(
		ctx, s, projectID, dataTypeStats,
		func(ctx context.Context, projectID int) (ProjectStats, error) {
			raw, err := s.repository.GetStatsByProject(ctx, projectID)
			if err != nil {
				return ProjectStats{}, fmt.Errorf("get stats by project: %w", err)
			}

			return toProjectStats(raw), nil
		},
	)
}

const dataTypeIssuesDuration = "issues_duration"

func (s *Service) GetIssuesDurationHistogram(
	ctx context.Context,
	projectID int,
) (IssuesDurationHistogram, error) {
	return fetchWithCache(
		ctx, s, projectID, dataTypeIssuesDuration,
		func(ctx context.Context, projectID int) (IssuesDurationHistogram, error) {
			rows, err := s.repository.GetIssuesDurationByProject(ctx, projectID)
			if err != nil {
				return IssuesDurationHistogram{}, fmt.Errorf("get issues duration: %w", err)
			}

			return buildIssuesDurationHistogram(rows)
		},
	)
}

const dataTypeStatusTransitions = "status_transitions"

func (s *Service) GetStatusHistograms(
	ctx context.Context,
	projectID int,
) ([]StatusHistogram, error) {
	return fetchWithCache(
		ctx, s, projectID, dataTypeStatusTransitions,
		func(ctx context.Context, projectID int) ([]StatusHistogram, error) {
			rows, err := s.repository.GetStatusTransitionsByProject(ctx, projectID)
			if err != nil {
				return nil, fmt.Errorf("get status transitions: %w", err)
			}

			return buildStatusHistograms(rows)
		},
	)
}

const dataTypeDailyActivity = "daily_activity"

func (s *Service) GetDailyActivityChart(
	ctx context.Context,
	projectID int,
) (DailyActivityChart, error) {
	return fetchWithCache(
		ctx, s, projectID, dataTypeDailyActivity,
		func(ctx context.Context, projectID int) (DailyActivityChart, error) {
			rows, err := s.repository.GetDailyActivityByProject(ctx, projectID)
			if err != nil {
				return DailyActivityChart{}, fmt.Errorf("get daily activity: %w", err)
			}

			return buildDailyActivityChart(rows)
		},
	)
}

const dataTypeIssuesTimeSpent = "issues_time_spent"

func (s *Service) GetIssuesTimeSpentHistogram(
	ctx context.Context,
	projectID int,
) (IssuesTimeSpentHistogram, error) {
	return fetchWithCache(
		ctx, s, projectID, dataTypeIssuesTimeSpent,
		func(ctx context.Context, projectID int) (IssuesTimeSpentHistogram, error) {
			rows, err := s.repository.GetIssuesTimeSpentByProject(ctx, projectID)
			if err != nil {
				return IssuesTimeSpentHistogram{}, fmt.Errorf("get issues time spent: %w", err)
			}

			return buildIssuesTimeSpentHistogram(rows)
		},
	)
}

const dataTypePriorityStats = "priority_stats"

func (s *Service) GetPriorityChart(ctx context.Context, projectID int) (PriorityChart, error) {
	return fetchWithCache(
		ctx, s, projectID, dataTypePriorityStats,
		func(ctx context.Context, projectID int) (PriorityChart, error) {
			rows, err := s.repository.GetPriorityStatsByProject(ctx, projectID)
			if err != nil {
				return PriorityChart{}, fmt.Errorf("get priority stats: %w", err)
			}

			return buildPriorityChart(rows), nil
		},
	)
}

func (s *Service) GetChart(
	ctx context.Context,
	projectID int,
	chartType ChartType,
) ([]byte, error) {
	switch chartType {
	case ChartTypeOpenStateHistogram:
		return marshalResult(s.GetIssuesDurationHistogram(ctx, projectID))
	case ChartTypeStateDistribution:
		return marshalResult(s.GetStatusHistograms(ctx, projectID))
	case ChartTypeComplexityHistogram:
		return marshalResult(s.GetIssuesTimeSpentHistogram(ctx, projectID))
	case ChartTypePriority:
		return marshalResult(s.GetPriorityChart(ctx, projectID))
	case ChartTypeDailyActivity:
		return marshalResult(s.GetDailyActivityChart(ctx, projectID))
	default:
		return nil, ErrUnknownChartType
	}
}

func (s *Service) CompareTwoProjects(
	ctx context.Context,
	lhsProjectID, rhsProjectID int,
) ([2]ProjectStats, error) {
	var result [2]ProjectStats

	group, ctx := errgroup.WithContext(ctx)

	group.Go(func() error {
		var err error

		result[0], err = s.GetProjectStat(ctx, lhsProjectID)

		return err
	})

	group.Go(func() error {
		var err error

		result[1], err = s.GetProjectStat(ctx, rhsProjectID)

		return err
	})

	err := group.Wait()
	if err != nil {
		return [2]ProjectStats{}, fmt.Errorf("compare projects: %w", err)
	}

	return result, nil
}

func (s *Service) CompareProjectsCharts(
	ctx context.Context,
	lhsProjectID, rhsProjectID int,
	chartType ChartType,
) ([2][]byte, error) {
	var result [2][]byte

	group, ctx := errgroup.WithContext(ctx)

	group.Go(func() error {
		var err error

		result[0], err = s.GetChart(ctx, lhsProjectID, chartType)

		return err
	})

	group.Go(func() error {
		var err error

		result[1], err = s.GetChart(ctx, rhsProjectID, chartType)

		return err
	})

	err := group.Wait()
	if err != nil {
		return [2][]byte{}, fmt.Errorf("compare projects charts: %w", err)
	}

	return result, nil
}

func (s *Service) isCacheStale(ctx context.Context, projectID int) (bool, error) {
	s.lastCheckedMu.RLock()

	if checkedAt, ok := s.lastCheckedAt[projectID]; ok && time.Since(checkedAt) < staleCheckTTL {
		s.lastCheckedMu.RUnlock()

		return false, nil
	}

	s.lastCheckedMu.RUnlock()

	dbUpdatedAt, err := s.repository.GetProjectLastUpdated(ctx, projectID)
	if err != nil {
		return false, fmt.Errorf("get project last updated: %w", err)
	}

	cacheUpdatedAt, err := s.cache.GetLastUpdated(ctx, projectID)
	if err != nil {
		return true, nil //nolint:nilerr
	}

	stale := dbUpdatedAt.After(cacheUpdatedAt)
	if !stale {
		s.lastCheckedMu.Lock()
		s.lastCheckedAt[projectID] = time.Now()
		s.lastCheckedMu.Unlock()
	}

	return stale, nil
}

func fetchWithCache[T any]( //nolint:ireturn
	ctx context.Context,
	svc *Service,
	projectID int,
	dataType string,
	fetch func(context.Context, int) (T, error),
) (T, error) {
	var zero T

	stale, err := svc.isCacheStale(ctx, projectID)
	if err != nil {
		return zero, err
	}

	if !stale {
		cached, cacheErr := svc.cache.Get(ctx, projectID, dataType)
		if cacheErr == nil {
			var result T
			if json.Unmarshal(cached, &result) == nil {
				return result, nil
			}
		}
	}

	key := fmt.Sprintf("%d:%s", projectID, dataType)

	sfResult, sfErr, _ := svc.sfGroup.Do(key, func() (any, error) {
		result, fetchErr := fetch(ctx, projectID)
		if fetchErr != nil {
			return nil, fetchErr
		}

		data, marshalErr := json.Marshal(result)
		if marshalErr == nil {
			_ = svc.cache.Set(ctx, projectID, dataType, data)
			_ = svc.cache.SetLastUpdated(ctx, projectID, time.Now())
		}

		return result, nil
	})
	if sfErr != nil {
		return zero, fmt.Errorf("fetch %s: %w", dataType, sfErr)
	}

	return sfResult.(T), nil //nolint:forcetypeassert
}

func marshalResult[T any](val T, err error) ([]byte, error) { //nolint:ireturn
	if err != nil {
		return nil, err
	}

	data, marshalErr := json.Marshal(val)
	if marshalErr != nil {
		return nil, fmt.Errorf("marshal result: %w", marshalErr)
	}

	return data, nil
}
