package redis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"product-service/domain"
	"time"

	"github.com/go-redis/redis/v8"
)

type ProductCache interface {
	GetProductByID(ctx context.Context, productID int) (product domain.Product, err error)
	SetProduct(ctx context.Context, product domain.Product, expiration time.Duration) (err error)
}

type productCache struct {
	rdb *redis.Client
}

func NewProductCache(rdb *redis.Client) ProductCache {
	return &productCache{rdb}
}

func (r *productCache) GetProductByID(ctx context.Context, productID int) (product domain.Product, err error) {
	key := fmt.Sprintf("product:%d", productID)
	productCache, err := r.rdb.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return product, nil
		} else {
			return product, err
		}
	}

	err = json.Unmarshal([]byte(productCache), &product)
	if err != nil {
		return product, err
	}

	return product, nil
}

func (r *productCache) SetProduct(ctx context.Context, product domain.Product, expiration time.Duration) (err error) {
	productByte, err := json.Marshal(product)
	if err != nil {
		return err
	}

	key := fmt.Sprintf("product:%d", product.ID)
	err = r.rdb.Set(ctx, key, productByte, 0).Err()
	if err != nil {
		return err
	}
	return nil
}
