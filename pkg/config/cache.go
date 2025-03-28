package config

import (
	"github.com/redis/go-redis/v9"
	"github.com/hftamayo/gotodo/pkg/utils"
)

func SetupCache(redisClient *redis.Client) *utils.Cache {
	return utils.NewCache(redisClient)
}
