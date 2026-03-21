package memory

import (
	"context"
	"errors"
	"sync"
	"time"
)

var (
	ErrProjectNotFound  = errors.New("project not found in cache")
	ErrDataTypeNotFound = errors.New("data type not found in cache")
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

	project, ok := c.data[projectID]
	if !ok {
		return nil, ErrProjectNotFound
	}

	value, ok := project[dataType]
	if !ok {
		return nil, ErrDataTypeNotFound
	}

	return value, nil
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

	t, ok := c.updatedAt[projectID]
	if !ok {
		return time.Time{}, ErrProjectNotFound
	}

	return t, nil
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
