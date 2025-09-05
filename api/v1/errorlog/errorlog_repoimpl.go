package errorlog

import (
	"context"
	"time"

	"github.com/hftamayo/gotodo/pkg/utils"
)

type ErrorLogRepositoryImpl struct {
	Redis utils.RedisClientInterface
}

func NewErrorLogRepositoryImpl(redisClient utils.RedisClientInterface) *ErrorLogRepositoryImpl {
	return &ErrorLogRepositoryImpl{Redis: redisClient}
}

func (r *ErrorLogRepositoryImpl) LogError(operation string, err error) error {
	ctx := context.Background()
	logEntry := map[string]interface{}{
		"operation": operation,
		"error":     err.Error(),
		"timestamp": time.Now().Format(time.RFC3339),
	}
	return r.Redis.HMSet(ctx, "logs", logEntry).Err()
}
