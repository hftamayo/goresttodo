package utils

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type RateLimiter struct {
	RedisClient *redis.Client
	Limit       int
	Window      time.Duration
}

func NewRateLimiter(redisClient *redis.Client, limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		RedisClient: redisClient,
		Limit:       limit,
		Window:      window,
	}
}

func (r *RateLimiter) Allow(key string) (bool, error) {
	ctx := context.Background()

	// Start a transaction
	pipe := r.RedisClient.TxPipeline()

	pipe.Incr(ctx, key)

	// Set the expiration time for the key if it doesn't exist
	pipe.Expire(ctx, key, r.Window)

	// Execute the transaction
	_, err := pipe.Exec(ctx)
	if err != nil {
		return false, err
	}

	// Get the current count
	count, err := r.RedisClient.Get(ctx, key).Int()
	if err != nil {
		return false, err
	}

	// Check if the count exceeds the limit
	if count > r.Limit {
		return false, nil
	}

	return true, nil
}
