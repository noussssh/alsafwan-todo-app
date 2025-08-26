package cache

import (
	"encoding/json"
	"sync"
	"time"
)

// CacheItem represents a cached item with expiration
type CacheItem struct {
	Data      interface{}
	ExpiresAt time.Time
}

// IsExpired checks if the cache item has expired
func (ci *CacheItem) IsExpired() bool {
	return time.Now().After(ci.ExpiresAt)
}

// Cache is a simple in-memory cache with TTL support
type Cache struct {
	items   sync.Map
	cleanup *time.Ticker
	stop    chan bool
}

// New creates a new cache instance
func New(cleanupInterval time.Duration) *Cache {
	c := &Cache{
		cleanup: time.NewTicker(cleanupInterval),
		stop:    make(chan bool),
	}
	
	// Start cleanup goroutine
	go c.cleanupExpired()
	
	return c
}

// Set stores a value with TTL
func (c *Cache) Set(key string, value interface{}, ttl time.Duration) error {
	item := &CacheItem{
		Data:      value,
		ExpiresAt: time.Now().Add(ttl),
	}
	
	c.items.Store(key, item)
	return nil
}

// Get retrieves a value from cache
func (c *Cache) Get(key string) (interface{}, bool) {
	value, exists := c.items.Load(key)
	if !exists {
		return nil, false
	}
	
	item := value.(*CacheItem)
	if item.IsExpired() {
		c.items.Delete(key)
		return nil, false
	}
	
	return item.Data, true
}

// SetJSON stores a JSON-serializable value
func (c *Cache) SetJSON(key string, value interface{}, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	
	return c.Set(key, data, ttl)
}

// GetJSON retrieves and unmarshals a JSON value
func (c *Cache) GetJSON(key string, dest interface{}) (bool, error) {
	data, exists := c.Get(key)
	if !exists {
		return false, nil
	}
	
	jsonData, ok := data.([]byte)
	if !ok {
		return false, nil
	}
	
	err := json.Unmarshal(jsonData, dest)
	return err == nil, err
}

// Delete removes an item from cache
func (c *Cache) Delete(key string) {
	c.items.Delete(key)
}

// Clear removes all items from cache
func (c *Cache) Clear() {
	c.items.Range(func(key, value interface{}) bool {
		c.items.Delete(key)
		return true
	})
}

// cleanupExpired removes expired items
func (c *Cache) cleanupExpired() {
	for {
		select {
		case <-c.cleanup.C:
			now := time.Now()
			c.items.Range(func(key, value interface{}) bool {
				item := value.(*CacheItem)
				if now.After(item.ExpiresAt) {
					c.items.Delete(key)
				}
				return true
			})
		case <-c.stop:
			return
		}
	}
}

// Close stops the cleanup goroutine
func (c *Cache) Close() {
	c.cleanup.Stop()
	close(c.stop)
}

// Size returns the number of items in cache (for monitoring)
func (c *Cache) Size() int {
	count := 0
	c.items.Range(func(key, value interface{}) bool {
		count++
		return true
	})
	return count
}