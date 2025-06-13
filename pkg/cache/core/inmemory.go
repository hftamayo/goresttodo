package core

import (
	"encoding/json"
	"errors"
	"path/filepath"
	"sync"
	"time"
)

// InMemoryClient implements the Client interface with an in-memory map
type InMemoryClient struct {
    data  map[string]cacheItem
    mutex sync.RWMutex
}

type cacheItem struct {
    value      []byte
    expiration time.Time
}

// NewInMemoryClient creates a new in-memory cache client
func NewInMemoryClient() *InMemoryClient {
    return &InMemoryClient{
        data: make(map[string]cacheItem),
    }
}

// Get retrieves a value from cache
func (c *InMemoryClient) Get(key string, dest interface{}) error {
    c.mutex.RLock()
    item, exists := c.data[key]
    c.mutex.RUnlock()

    if !exists {
        return errors.New("key not found")
    }

    if !item.expiration.IsZero() && item.expiration.Before(time.Now()) {
        c.Delete(key)
        return errors.New("key expired")
    }

    return json.Unmarshal(item.value, dest)
}

// Set stores a value in cache
func (c *InMemoryClient) Set(key string, value interface{}, expiration time.Duration) error {
    data, err := json.Marshal(value)
    if err != nil {
        return err
    }

    var exp time.Time
    if expiration > 0 {
        exp = time.Now().Add(expiration)
    }

    c.mutex.Lock()
    c.data[key] = cacheItem{
        value:      data,
        expiration: exp,
    }
    c.mutex.Unlock()

    return nil
}

// Delete removes a key from cache
func (c *InMemoryClient) Delete(key string) error {
    c.mutex.Lock()
    delete(c.data, key)
    c.mutex.Unlock()
    return nil
}

// DeletePattern removes keys matching a pattern
func (c *InMemoryClient) DeletePattern(pattern string) error {
    c.mutex.Lock()
    defer c.mutex.Unlock()
    
    for k := range c.data {
        matched, err := filepath.Match(pattern, k)
        if err != nil {
            return err
        }
        if matched {
            delete(c.data, k)
        }
    }
    return nil
}

// Exists checks if a key exists
func (c *InMemoryClient) Exists(key string) bool {
    c.mutex.RLock()
    item, exists := c.data[key]
    c.mutex.RUnlock()

    if !exists {
        return false
    }

    if !item.expiration.IsZero() && item.expiration.Before(time.Now()) {
        c.Delete(key)
        return false
    }

    return true
}

// Clear empties the cache
func (c *InMemoryClient) Clear() error {
    c.mutex.Lock()
    c.data = make(map[string]cacheItem)
    c.mutex.Unlock()
    return nil
}