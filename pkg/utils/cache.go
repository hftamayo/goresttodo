package utils

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/redis/v9"
)

type Cache struct {
	RedisClient *redis.Client
}

func NewCache(redisClient *redis.Client) *Cache {
	return &Cache{RedisClient: redisClient}
}

func (c *Cache) Set(key string, value interface{}, expiration time.Duration) error {
	ctx := context.Background()
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return c.RedisClient.Set(ctx, key, data, expiration).Err()
}

func (c *Cache) Get(key string, dest interface{}) error {
	ctx := context.Background()
	data, err := c.RedisClient.Get(ctx, key).Result()
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(data), dest)
}
