package cache

import (
	"time"

	gocache "github.com/patrickmn/go-cache"
)

// Cache wraps go-cache for in-memory caching
type Cache struct {
	store *gocache.Cache
}

// NewCache creates a new cache instance
// defaultExpiration: default expiration time for items
// cleanupInterval: interval for cleaning up expired items
func NewCache(defaultExpiration, cleanupInterval time.Duration) *Cache {
	return &Cache{
		store: gocache.New(defaultExpiration, cleanupInterval),
	}
}

// Set adds an item to the cache with default expiration
func (c *Cache) Set(key string, value interface{}) {
	c.store.Set(key, value, gocache.DefaultExpiration)
}

// SetWithExpiration adds an item to the cache with custom expiration
func (c *Cache) SetWithExpiration(key string, value interface{}, expiration time.Duration) {
	c.store.Set(key, value, expiration)
}

// Get retrieves an item from the cache
func (c *Cache) Get(key string) (interface{}, bool) {
	return c.store.Get(key)
}

// Delete removes an item from the cache
func (c *Cache) Delete(key string) {
	c.store.Delete(key)
}

// Flush removes all items from the cache
func (c *Cache) Flush() {
	c.store.Flush()
}

// ItemCount returns the number of items in the cache
func (c *Cache) ItemCount() int {
	return c.store.ItemCount()
}

// GetOrSet retrieves an item from cache, or sets it if not found
func (c *Cache) GetOrSet(key string, fetchFunc func() (interface{}, error)) (interface{}, error) {
	// Try to get from cache first
	if value, found := c.Get(key); found {
		return value, nil
	}

	// Not in cache, fetch it
	value, err := fetchFunc()
	if err != nil {
		return nil, err
	}

	// Store in cache
	c.Set(key, value)
	return value, nil
}
