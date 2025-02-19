package config

import (
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/hftamayo/gotodo/pkg/utils"
)

func SetupRateLimiter(redisClient *redis.Client, limit int, window time.Duration) *utils.RateLimiter {
	return utils.NewRateLimiter(redisClient, limit, window)
}
