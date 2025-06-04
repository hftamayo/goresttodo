package utils

import (
	"context"
	"strconv"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

type OperationType string
const (
	OperationRead OperationType = "read"
	OperationWrite OperationType = "write"
	OperationPrefetch OperationType = "prefetch"
)

type RateLimiter struct {
	RedisClient *redis.Client
    operationLimits map[OperationType]int
    Window          time.Duration
    mu              sync.RWMutex

}

func NewRateLimiter(redisClient *redis.Client) *RateLimiter {
    return &RateLimiter{
        RedisClient: redisClient,
        operationLimits: map[OperationType]int{
            OperationRead:     100, 
            OperationWrite:    30,  
            OperationPrefetch: 200, 
        },
        Window: time.Minute, 
        mu:     sync.RWMutex{},
    }
}

// SetLimitForOperation sets the rate limit for a specific operation type
func (r *RateLimiter) SetLimitForOperation(op OperationType, limit int) {
    r.mu.Lock()
    defer r.mu.Unlock()
    r.operationLimits[op] = limit
}

// GetLimitForOperation gets the rate limit for a specific operation type
func (r *RateLimiter) GetLimitForOperation(op OperationType) int {
    r.mu.RLock()
    defer r.mu.RUnlock()
    if limit, exists := r.operationLimits[op]; exists {
        return limit
    }
    return r.operationLimits[OperationRead] // Default to read limit
}

// Allow checks if a request is allowed based on client ID and operation type
func (r *RateLimiter) Allow(clientID string) (bool, error) {
    allowed, _, _, err := r.AllowOperation(clientID, OperationRead)
    return allowed, err
}

// AllowOperation checks if a request is allowed for a specific operation type
// and returns whether it's allowed, the limit, retry time, and any error
func (r *RateLimiter) AllowOperation(clientID string, op OperationType) (bool, int, time.Time, error) {
    ctx := context.Background()
    
    // Get the rate limit for this operation
    limit := r.GetLimitForOperation(op)
    
    // Create a key that includes operation type
    key := "ratelimit:" + clientID + ":" + string(op)
    
    // Get the current window
    windowKey := key + ":window"
    
    // Check if we're in an existing window
    windowStartStr, err := r.RedisClient.Get(ctx, windowKey).Result()
    if err != nil && err != redis.Nil {
        return false, limit, time.Time{}, err
    }
    
    now := time.Now()
    var windowStart time.Time
    var requestCount int64
    
    if err == redis.Nil || windowStartStr == "" {
        // No window exists, create a new one
        windowStart = now
        if err := r.RedisClient.Set(ctx, windowKey, now.Unix(), r.Window).Err(); err != nil {
            return false, limit, time.Time{}, err
        }
    } else {
        // Window exists, parse the timestamp
        windowStartUnix, _ := strconv.ParseInt(windowStartStr, 10, 64)
        windowStart = time.Unix(windowStartUnix, 0)
        
        // If window has expired, create a new one
        if now.Sub(windowStart) > r.Window {
            windowStart = now
            if err := r.RedisClient.Set(ctx, windowKey, now.Unix(), r.Window).Err(); err != nil {
                return false, limit, time.Time{}, err
            }
            // Reset count for new window
            if err := r.RedisClient.Set(ctx, key, 0, r.Window).Err(); err != nil {
                return false, limit, time.Time{}, err
            }
        }
    }
    
    // Increment the counter and check if we're over limit
    requestCount, err = r.RedisClient.Incr(ctx, key).Result()
    if err != nil {
        return false, limit, time.Time{}, err
    }
    
    // Ensure the counter expires with the window
    if err := r.RedisClient.Expire(ctx, key, r.Window).Err(); err != nil {
        return false, limit, time.Time{}, err
    }
    
    // Check if we're over the limit
    if requestCount > int64(limit) {
        // Calculate when the window will expire
        timeUntilReset := r.Window - now.Sub(windowStart)
        retryTime := now.Add(timeUntilReset)
        return false, limit, retryTime, nil
    }
    
    return true, limit, time.Time{}, nil
}