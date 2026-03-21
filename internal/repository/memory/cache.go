package memory

import (
	"context"
	"sync"
	"time"

	"github.com/Go-Yadro-Group-1/Jira-Analyzer/internal/repository"
)

type CacheRepository struct {
	mu        sync.RWMutex
	data      map[int]map[repository.DataType][]byte
	updatedAt map[int]time.Time
}

func NewCacheRepository() *CacheRepository {
	return &CacheRepository{
		mu:        sync.RWMutex{},
		data:      make(map[int]map[repository.DataType][]byte),
		updatedAt: make(map[int]time.Time),
	}
}

func (c *CacheRepository) Get(
	_ context.Context,
	projectID int,
	dataType repository.DataType,
) ([]byte, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if project, ok := c.data[projectID]; ok {
		if value, ok := project[dataType]; ok {
			return value, nil
		}
	}

	return nil, nil
}

func (c *CacheRepository) Set(
	_ context.Context,
	projectID int,
	dataType repository.DataType,
	value []byte,
) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, ok := c.data[projectID]; !ok {
		c.data[projectID] = make(map[repository.DataType][]byte)
	}

	c.data[projectID][dataType] = value

	return nil
}

func (c *CacheRepository) Invalidate(_ context.Context, projectID int) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.data, projectID)
	delete(c.updatedAt, projectID)

	return nil
}

func (c *CacheRepository) GetLastUpdated(_ context.Context, projectID int) (time.Time, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.updatedAt[projectID], nil
}

func (c *CacheRepository) SetLastUpdated(_ context.Context, projectID int, t time.Time) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.updatedAt[projectID] = t

	return nil
}
