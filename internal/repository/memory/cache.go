package memory

import (
	"context"
	"sync"
	"time"
)

type CacheRepository[ProjectKey comparable, HistogrammKey comparable] struct {
	mu        sync.RWMutex
	data      map[ProjectKey]map[HistogrammKey][]byte
	updatedAt map[ProjectKey]time.Time
}

func NewCacheRepository[ProjectKey comparable, HistogrammKey comparable]() *CacheRepository[ProjectKey, HistogrammKey] {
	return &CacheRepository[ProjectKey, HistogrammKey]{
		mu:        sync.RWMutex{},
		data:      make(map[ProjectKey]map[HistogrammKey][]byte),
		updatedAt: make(map[ProjectKey]time.Time),
	}
}

func (c *CacheRepository[ProjectKey, HistogrammKey]) Get(
	_ context.Context,
	projectID ProjectKey,
	dataType HistogrammKey,
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

func (c *CacheRepository[ProjectKey, HistogrammKey]) Set(
	_ context.Context,
	projectID ProjectKey,
	dataType HistogrammKey,
	value []byte,
) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, ok := c.data[projectID]; !ok {
		c.data[projectID] = make(map[HistogrammKey][]byte)
	}

	c.data[projectID][dataType] = value

	return nil
}

func (c *CacheRepository[ProjectKey, HistogrammKey]) Invalidate(
	_ context.Context,
	projectID ProjectKey,
) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.data, projectID)
	delete(c.updatedAt, projectID)

	return nil
}

func (c *CacheRepository[ProjectKey, HistogrammKey]) GetLastUpdated(
	_ context.Context,
	projectID ProjectKey,
) (time.Time, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.updatedAt[projectID], nil
}

func (c *CacheRepository[ProjectKey, HistogrammKey]) SetLastUpdated(
	_ context.Context,
	projectID ProjectKey,
	t time.Time,
) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.updatedAt[projectID] = t

	return nil
}
