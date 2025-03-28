package config

import (
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/hftamayo/gotodo/pkg/utils"
)

func SetupRateLimiter(redisClient *redis.Client, limit int, window time.Duration) *utils.RateLimiter {
	return utils.NewRateLimiter(redisClient, limit, window)
}
