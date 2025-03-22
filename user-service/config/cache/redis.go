package cache

import (
	"context"
	"fmt"
	"time"

	"user-service/config"

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

	fmt.Println("nknk")
	err = rdb.Set(context.Background(), "email", "dddt", time.Hour*24).Err() // Set expiration to 24 hours
	if err != nil {
		return
	}

	return
}
