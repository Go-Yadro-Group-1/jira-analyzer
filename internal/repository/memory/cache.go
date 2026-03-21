package memory

import (
	"context"
	"sync"
	"time"

	"github.com/Go-Yadro-Group-1/Jira-Analyzer/internal/repository"
)

type cacheRepository struct {
	mu        sync.RWMutex
	data      map[int]map[repository.DataType][]byte
	updatedAt map[int]time.Time
}

func NewCacheRepository() repository.CacheRepository {
	return &cacheRepository{
		data:      make(map[int]map[repository.DataType][]byte),
		updatedAt: make(map[int]time.Time),
	}
}

func (c *cacheRepository) Get(_ context.Context, projectID int, dataType repository.DataType) ([]byte, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if project, ok := c.data[projectID]; ok {
		if value, ok := project[dataType]; ok {
			return value, nil
		}
	}
	return nil, nil
}

func (c *cacheRepository) Set(_ context.Context, projectID int, dataType repository.DataType, value []byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, ok := c.data[projectID]; !ok {
		c.data[projectID] = make(map[repository.DataType][]byte)
	}
	c.data[projectID][dataType] = value
	return nil
}

func (c *cacheRepository) Invalidate(_ context.Context, projectID int) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.data, projectID)
	delete(c.updatedAt, projectID)
	return nil
}

func (c *cacheRepository) GetLastUpdated(_ context.Context, projectID int) (time.Time, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.updatedAt[projectID], nil
}

func (c *cacheRepository) SetLastUpdated(_ context.Context, projectID int, t time.Time) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.updatedAt[projectID] = t
	return nil
}
