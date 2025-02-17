package errorlog

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
)

type ErrorLogRepositoryImpl struct {
	Redis *redis.Client
}

func NewErrorLogRepositoryImpl(redisClient *redis.Client) *ErrorLogRepositoryImpl {
	return &ErrorLogRepositoryImpl{Redis: redisClient}
}

func (r *ErrorLogRepositoryImpl) LogError(operation string, err error) error {
	ctx := context.Background()
	logEntry := map[string]interface{}{
		"operation": operation,
		"error":     err.Error(),
		"timestamp": time.Now().Format(time.RFC3339),
	}
	return r.Redis.HSet(ctx, "logs", time.Now().UnixNano(), logEntry).Err()
}
