package cache

import (
	"context"
	"fmt"

	"order-service/config"

	"github.com/go-redis/redis/v8"
)

func NewRedisClient(cfg *config.Config) (rdb *redis.Client, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("error occured %+v", r)
		}
	}()

	rdbConfig := cfg.Redis

	rdb = redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%s", rdbConfig.Host, rdbConfig.Port),
	})

	statCmd := rdb.Ping(context.Background())
	err = statCmd.Err()
	if err != nil {
		return
	}

	return
}
