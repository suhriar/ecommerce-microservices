package redis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"product-service/domain"
	"time"

	"github.com/go-redis/redis/v8"
	"gorm.io/gorm/logger"
)

type ProductCache interface {
}

type productCacheImpl struct {
	rdb *redis.Client
}

func NewProductCache(rdb *redis.Client) ProductCache {
	return &productCacheImpl{rdb}
}

func (r *productCacheImpl) GetProductByID(ctx context.Context, productID int) (product domain.Product, err error) {
	key := fmt.Sprintf("product:%d", productID)
	productCache, err := p.cache.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			logger.Warn().Msgf("Stock for product %d not found in cache", productID)
		} else {
			logger.Error().Err(err).Msgf("Error getting stock for product %d from cache", productID)
			return 0, err
		}
	}

	if productCache != "" {
		var product domain.Product
		err = json.Unmarshal([]byte(productCache), &product)
		if err != nil {
			logger.Error().Err(err).Msgf("Error unmarshalling product %d", productID)
			return 0, err
		}

		logger.Info().Msgf("Retrieved stock for product %d: %d", productID, product.Stock)
		return product.Stock, nil
	}

	return
}

func (r *productCacheImpl) SetUserTokenByEmail(ctx context.Context, email, token string, expiration time.Duration) (err error) {
	err = r.rdb.Set(ctx, email, token, expiration).Err() // Set expiration to 24 hours
	if err != nil {
		return err
	}
	return nil
}
