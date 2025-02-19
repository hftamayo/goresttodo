package config

import (
	"github.com/go-redis/redis/v8"
	"github.com/hftamayo/gotodo/pkg/utils"
)

func SetupCache(redisClient *redis.Client) *utils.Cache {
	return utils.NewCache(redisClient)
}
