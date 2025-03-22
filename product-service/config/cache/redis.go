package cache

import (
	"context"
	"fmt"

	"product-service/config"

	"github.com/go-redis/redis/v8"
)

func NewRedisClient(cfg *config.Config) (rdb *redis.Client, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("error occured %+v", r)
		}
	}()

	rdbConfig := cfg.Redis

	redisClient := redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%s", rdbConfig.Host, rdbConfig.Port),
	})

	statCmd := redisClient.Ping(context.Background())
	err = statCmd.Err()
	if err != nil {
		return
	}

	return
}
