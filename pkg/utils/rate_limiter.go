package utils

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// OperationType represents the type of operation being rate limited
type OperationType string

const (
	OperationRead     OperationType = "read"
	OperationWrite    OperationType = "write"
	OperationPrefetch OperationType = "prefetch"
)

// RateLimitConfig holds the configuration for rate limiting
type RateLimitConfig struct {
	MaxRequests int64 // Maximum number of requests allowed in the window
	Window      int64 // Time window in seconds
}

type RateLimiter struct {
	RedisClient RedisClientInterface
	operationLimits map[OperationType]int
	Window          time.Duration
	mu              sync.RWMutex
}

func NewRateLimiter(redisClient RedisClientInterface) *RateLimiter {
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

// Allow checks if a request is allowed for a specific operation type
func (r *RateLimiter) Allow(clientID string, op OperationType) (bool, int64, time.Time, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Get the appropriate rate limit configuration based on operation type
	var config *RateLimitConfig
	switch op {
	case OperationRead:
		config = &RateLimitConfig{
			MaxRequests: 100, // Default read limit
			Window:      60,  // 1 minute window
		}
	case OperationWrite:
		config = &RateLimitConfig{
			MaxRequests: 50, // Default write limit
			Window:      60, // 1 minute window
		}
	case OperationPrefetch:
		config = &RateLimitConfig{
			MaxRequests: 200, // Default prefetch limit
			Window:      60,  // 1 minute window
		}
	default:
		return false, 0, time.Time{}, fmt.Errorf("unknown operation type: %s", op)
	}

	return r.AllowOperation(clientID, config)
}

// validateRateLimitConfig validates the rate limit configuration
func validateRateLimitConfig(config *RateLimitConfig) error {
	if config == nil {
		return fmt.Errorf("rate limit config is nil")
	}
	if config.MaxRequests <= 0 {
		return fmt.Errorf("max requests must be greater than 0")
	}
	if config.Window <= 0 {
		return fmt.Errorf("window must be greater than 0")
	}
	return nil
}

// getRateLimitKey generates the Redis key for rate limiting
func getRateLimitKey(identifier string, config *RateLimitConfig) string {
	return fmt.Sprintf("rate_limit:%s:%d", identifier, config.Window)
}

// checkRateLimit checks if the operation is allowed based on current count
func checkRateLimit(currentCount int64, config *RateLimitConfig) bool {
	return currentCount <= config.MaxRequests
}

// AllowOperation checks if an operation is allowed based on rate limiting rules
func (rl *RateLimiter) AllowOperation(identifier string, config *RateLimitConfig) (bool, int64, time.Time, error) {
	if err := validateRateLimitConfig(config); err != nil {
		return false, 0, time.Time{}, err
	}

	key := getRateLimitKey(identifier, config)
	ctx := context.Background()
	now := time.Now()

	// Get current count
	currentCount, err := rl.RedisClient.Incr(ctx, key).Result()
	if err != nil {
		return false, 0, time.Time{}, fmt.Errorf("failed to increment rate limit counter: %w", err)
	}

	// Set expiration if this is the first request
	if currentCount == 1 {
		if err := rl.RedisClient.Expire(ctx, key, time.Duration(config.Window)*time.Second).Err(); err != nil {
			return false, 0, time.Time{}, fmt.Errorf("failed to set rate limit expiration: %w", err)
		}
	}

	// Check if we're over the limit
	if currentCount > config.MaxRequests {
		// Calculate retry time based on window expiration
		retryTime := now.Add(time.Duration(config.Window) * time.Second)
		return false, config.MaxRequests, retryTime, nil
	}

	return true, config.MaxRequests, time.Time{}, nil
}