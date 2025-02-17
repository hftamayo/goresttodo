package config

import (
	"context"

	"github.com/go-redis/redis/v8"
)

func ErrorLogConnect() *redis.Client {
	redisClient := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	ctx := context.Background()
	_, err := redisClient.Ping(ctx).Result()
	if err != nil {
		return nil
	}
	return redisClient
}
