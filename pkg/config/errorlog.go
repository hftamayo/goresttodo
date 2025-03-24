package config

import (
	"context"
	"os"

	"github.com/redis/go-redis/v9"
)

func ErrorLogConnect() (*redis.Client, error) {
	redisClient := redis.NewClient(&redis.Options{
		Addr: os.Getenv("REDIS_HOST") + ":" + os.Getenv("REDIS_PORT"),
	})

	ctx := context.Background()
	_, err := redisClient.Ping(ctx).Result()
	if err != nil {
		return nil, err
	}
	return redisClient, nil
}
