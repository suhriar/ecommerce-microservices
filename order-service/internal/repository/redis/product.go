package redis

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

type OrderCache interface {
	GetIdempotentKeyIsNotExist(ctx context.Context, idempotentKey string) (exist bool, err error)
	SetIdempotentKey(ctx context.Context, idempotentKey string, expiration time.Duration) (err error)
}

type orderCache struct {
	rdb *redis.Client
}

func NewOrderCache(rdb *redis.Client) OrderCache {
	return &orderCache{rdb}
}

func (r *orderCache) GetIdempotentKeyIsNotExist(ctx context.Context, idempotentKey string) (exist bool, err error) {
	key := fmt.Sprintf("idempotent-key:%s", idempotentKey)
	val, err := r.rdb.Get(ctx, key).Result()
	if err != nil && !errors.Is(err, redis.Nil) {
		return false, err
	}

	if val != "" {
		return false, errors.New("idempotent key already exists")
	}

	return true, nil
}

func (r *orderCache) SetIdempotentKey(ctx context.Context, idempotentKey string, expiration time.Duration) (err error) {
	key := fmt.Sprintf("idempotent-key:%s", idempotentKey)
	err = r.rdb.Set(ctx, key, "exists", 0).Err()
	if err != nil {
		return err
	}
	return nil
}
