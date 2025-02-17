package config

import (
	"context"

	"github.com/go-redis/redis/v8"
)

func ErrorLogConnect() (*redis.Client, error) {
	redisClient := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	ctx := context.Background()
	_, err := redisClient.Ping(ctx).Result()
	if err != nil {
		return nil, err
	}
	return redisClient, nil
}
