package redis

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

type UserCache interface {
	GetUserTokenByEmail(ctx context.Context, email string) (token string, err error)
	SetUserTokenByEmail(ctx context.Context, email, token string, expiration time.Duration) (err error)
}

type userCache struct {
	rdb *redis.Client
}

func NewUserCache(rdb *redis.Client) UserCache {
	return &userCache{rdb}
}

func (r *userCache) GetUserTokenByEmail(ctx context.Context, email string) (token string, err error) {
	token, err = r.rdb.Get(ctx, email).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return "", fmt.Errorf("session not found")
		}
		return "", err
	}

	return
}

func (r *userCache) SetUserTokenByEmail(ctx context.Context, email, token string, expiration time.Duration) (err error) {
	err = r.rdb.Set(ctx, email, token, expiration).Err() // Set expiration to 24 hours
	if err != nil {
		return err
	}
	return nil
}
