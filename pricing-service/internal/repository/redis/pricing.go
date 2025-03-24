package redis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"pricing-service/domain"
	"time"

	"github.com/go-redis/redis/v8"
)

type PricingCache interface {
	GetPricingRule(ctx context.Context, productID int) (rule domain.PricingRule, err error)
	SetProduct(ctx context.Context, rule domain.PricingRule, expiration time.Duration) (err error)
}

type pricingCache struct {
	rdb *redis.Client
}

func NewPricingCache(rdb *redis.Client) PricingCache {
	return &pricingCache{rdb}
}

func (r *pricingCache) GetPricingRule(ctx context.Context, productID int) (rule domain.PricingRule, err error) {
	key := fmt.Sprintf("pricing_rule:%d", productID)
	pricingCache, err := r.rdb.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return rule, nil
		} else {
			return rule, err
		}
	}

	err = json.Unmarshal([]byte(pricingCache), &rule)
	if err != nil {
		return rule, err
	}

	return rule, nil
}

func (r *pricingCache) SetProduct(ctx context.Context, rule domain.PricingRule, expiration time.Duration) (err error) {
	ruleByte, err := json.Marshal(rule)
	if err != nil {
		return err
	}

	key := fmt.Sprintf("pricing_rule:%d", rule.ID)
	err = r.rdb.Set(ctx, key, ruleByte, 0).Err()
	if err != nil {
		return err
	}
	return nil
}
