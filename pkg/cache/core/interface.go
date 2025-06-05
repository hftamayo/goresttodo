package core

import "time"

// Client defines the core caching interface
type Client interface {
    // Get retrieves a value from cache and unmarshals it into dest
    Get(key string, dest interface{}) error
    
    // Set stores a value in cache with optional expiration
    Set(key string, value interface{}, expiration time.Duration) error
    
    // Delete removes a key from cache
    Delete(key string) error
    
    // DeletePattern removes keys matching a pattern (e.g., "users_*")
    DeletePattern(pattern string) error
    
    // Exists checks if a key exists in cache
    Exists(key string) bool
    
    // Clear empties the entire cache (use with caution)
    Clear() error
}