package config

import (
	"time"

	"github.com/hftamayo/gotodo/pkg/utils"
	"github.com/redis/go-redis/v9"
)

func SetupRateLimiter(redisClient *redis.Client, defaultLimit int, window time.Duration) *utils.RateLimiter {
    rateLimiter := utils.NewRateLimiter(redisClient)
    
    // Set the window
    rateLimiter.Window = window
    
    // Configure limits for different operation types
    rateLimiter.SetLimitForOperation(utils.OperationRead, 100)     // 100 read requests per minute
    rateLimiter.SetLimitForOperation(utils.OperationWrite, 30)     // 30 write requests per minute
    rateLimiter.SetLimitForOperation(utils.OperationPrefetch, 200) // 200 prefetch requests per minute
    
    return rateLimiter
}
