package utils

import "time"

// Cache-related constants
const (
    // Default cache expiration time
    DefaultCacheTime = 10 * time.Minute
    
    // Pagination-related constants
    DefaultLimit = 10
    MaxLimit     = 100
    
    // Default order direction
    DefaultOrder = "desc"
)