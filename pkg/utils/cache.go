package utils

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisClientInterface defines the interface for Redis operations
// This allows us to mock Redis operations in tests
type RedisClientInterface interface {
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd
	Get(ctx context.Context, key string) *redis.StringCmd
	Del(ctx context.Context, keys ...string) *redis.IntCmd
	Scan(ctx context.Context, cursor uint64, match string, count int64) *redis.ScanCmd
	SAdd(ctx context.Context, key string, members ...interface{}) *redis.IntCmd
	SRem(ctx context.Context, key string, members ...interface{}) *redis.IntCmd
	SMembers(ctx context.Context, key string) *redis.StringSliceCmd
	Expire(ctx context.Context, key string, expiration time.Duration) *redis.BoolCmd
	Incr(ctx context.Context, key string) *redis.IntCmd
	HSet(ctx context.Context, key string, values ...interface{}) *redis.IntCmd
	HMSet(ctx context.Context, key string, values ...interface{}) *redis.BoolCmd
	Close() error
}

type Cache struct {
	RedisClient RedisClientInterface
}

func NewCache(redisClient RedisClientInterface) *Cache {
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

func (c *Cache) Delete(key string) error {
	ctx := context.Background()
	return c.RedisClient.Del(ctx, key).Err()
}

// DeletePattern deletes all keys matching the given pattern
func (c *Cache) DeletePattern(pattern string) error {
	ctx := context.Background()
	iter := c.RedisClient.Scan(ctx, 0, pattern, 0).Iterator()
	for iter.Next(ctx) {
		if err := c.RedisClient.Del(ctx, iter.Val()).Err(); err != nil {
			return err
		}
	}
	return iter.Err()
}

func (c *Cache) DeleteByPrefix(prefix string) error {
    return c.DeletePattern(prefix + "*")
}

// SetWithTags sets a value with associated tags
func (c *Cache) SetWithTags(key string, value interface{}, expiration time.Duration, tags ...string) error {
	ctx := context.Background()
	
	// Set the main value
	if err := c.Set(key, value, expiration); err != nil {
		return err
	}

	// Add tags
	for _, tag := range tags {
		tagKey := "tag:" + tag
		if err := c.RedisClient.SAdd(ctx, tagKey, key).Err(); err != nil {
			return err
		}
		// Set expiration on tag set
		if err := c.RedisClient.Expire(ctx, tagKey, expiration).Err(); err != nil {
			return err
		}
	}

	return nil
}

// InvalidateByTags deletes all keys associated with the given tags
func (c *Cache) InvalidateByTags(tags ...string) error {
	ctx := context.Background()
	
	for _, tag := range tags {
		tagKey := "tag:" + tag
		// Get all keys for this tag
		keys, err := c.RedisClient.SMembers(ctx, tagKey).Result()
		if err != nil {
			return err
		}

		// Delete all keys
		for _, key := range keys {
			if err := c.RedisClient.Del(ctx, key).Err(); err != nil {
				return err
			}
		}

		// Delete the tag set
		if err := c.RedisClient.Del(ctx, tagKey).Err(); err != nil {
			return err
		}
	}

	return nil
}

// AddTag adds a tag to an existing key
func (c *Cache) AddTag(key string, tag string) error {
	ctx := context.Background()
	tagKey := "tag:" + tag
	return c.RedisClient.SAdd(ctx, tagKey, key).Err()
}

// RemoveTag removes a tag from a key
func (c *Cache) RemoveTag(key string, tag string) error {
	ctx := context.Background()
	tagKey := "tag:" + tag
	return c.RedisClient.SRem(ctx, tagKey, key).Err()
}