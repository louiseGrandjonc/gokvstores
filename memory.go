package gokvstores

import (
	"time"

	"github.com/patrickmn/go-cache"
)

// MemoryStore is the in-memory implementation of KVStore.
type MemoryStore struct {
	cache           *cache.Cache
	expiration      time.Duration
	cleanupInterval time.Duration
}

// Get returns item from the cache.
func (c *MemoryStore) Get(key string) (interface{}, error) {
	item, _ := c.cache.Get(key)
	return item, nil
}

// Set sets value in the cache.
func (c *MemoryStore) Set(key string, value interface{}) error {
	c.cache.Set(key, value, c.expiration)
	return nil
}

// GetMap returns map for the given key.
func (c *MemoryStore) GetMap(key string) (map[string]interface{}, error) {
	if v, found := c.cache.Get(key); found {
		return v.(map[string]interface{}), nil
	}
	return nil, nil
}

// SetMap sets a map for the given key.
func (c *MemoryStore) SetMap(key string, value map[string]interface{}) error {
	c.cache.Set(key, value, c.expiration)
	return nil
}

// GetSlice returns slice for the given key.
func (c *MemoryStore) GetSlice(key string) ([]interface{}, error) {
	if v, found := c.cache.Get(key); found {
		return v.([]interface{}), nil
	}
	return nil, nil
}

// SetSlice sets slice for the given key.
func (c *MemoryStore) SetSlice(key string, value []interface{}) error {
	c.cache.Set(key, value, c.expiration)
	return nil
}

// AppendSlice appends values to the given slice.
func (c *MemoryStore) AppendSlice(key string, values ...interface{}) error {
	items, err := c.GetSlice(key)
	if err != nil {
		return err
	}

	for _, item := range values {
		items = append(items, item)
	}

	return c.cache.Replace(key, items, c.expiration)
}

// Close does nothing for this backend.
func (c *MemoryStore) Close() error {
	return nil
}

// Flush removes all items from the cache.
func (c *MemoryStore) Flush() error {
	c.cache.Flush()
	return nil
}

// Delete deletes the given key.
func (c *MemoryStore) Delete(key string) error {
	c.cache.Delete(key)
	return nil
}

// Exists checks if the given key exists.
func (c *MemoryStore) Exists(key string) (bool, error) {
	if _, exists := c.cache.Get(key); exists {
		return true, nil
	}
	return false, nil
}

// NewMemoryStore returns in-memory KVStore.
func NewMemoryStore(expiration time.Duration, cleanupInterval time.Duration) (KVStore, error) {
	return &MemoryStore{
		cache:           cache.New(expiration, cleanupInterval),
		expiration:      time.Duration(expiration) * time.Second,
		cleanupInterval: cleanupInterval,
	}, nil
}
