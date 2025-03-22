package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"user-service/domain"

	"github.com/Jobhun/mono-api/service/repository/cache"
	"github.com/go-redis/redis/v8"
	"github.com/rs/zerolog"
	"github.com/suhriar/e-commerce-high-traffic/product-catalog-service/internal/entity"
)

var logger = zerolog.New(os.Stdout).With().Timestamp().Logger()

type ProductUsecase interface {
	GetUserByID(ctx context.Context, id int) (*domain.User, error)
	CreateUser(ctx context.Context, user *domain.User) (*domain.User, error)
	Login(ctx context.Context, email, password string) (token string, err error)
	ValidateToken(ctx context.Context, email string) (string, error)
}

type productUsecaseImpl struct {
	repo  repo.UserRepository
	cache cache.UserCache
}

func NewUserService(repo repo.UserRepository, cache cache.UserCache) UserUsecase {
	return &userServiceImpl{
		repo:  repo,
		cache: cache,
	}
}

// GetProductStock retrieves the stock for a product.
func (p *ProductService) GetProductStock(ctx context.Context, productID int) (int, error) {
	// Read from cache
	key := fmt.Sprintf("product:%d", productID)
	productCache, err := p.rdb.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			logger.Warn().Msgf("Stock for product %d not found in cache", productID)
		} else {
			logger.Error().Err(err).Msgf("Error getting stock for product %d from cache", productID)
			return 0, err
		}
	}

	if productCache != "" {
		var product entity.Product
		err = json.Unmarshal([]byte(productCache), &product)
		if err != nil {
			logger.Error().Err(err).Msgf("Error unmarshalling product %d", productID)
			return 0, err
		}

		logger.Info().Msgf("Retrieved stock for product %d: %d", productID, product.Stock)
		return product.Stock, nil
	}

	product, err := p.productRepo.GetProductByID(ctx, productID)
	if err != nil {
		logger.Error().Err(err).Msgf("Error getting product by ID %d", productID)
		return 0, err
	}

	// Write to cache
	err = p.rdb.Set(ctx, key, product, 0).Err()
	if err != nil {
		logger.Error().Err(err).Msgf("Error setting product %d in cache", productID)
		return 0, err
	}

	return product.Stock, nil
}

// ReserveProductStock reserves stock for an order.
func (p *ProductService) ReserveProductStock(ctx context.Context, productID int, quantity int) error {
	// Get product from cache
	key := fmt.Sprintf("product:%d", productID)
	productCache, err := p.rdb.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			logger.Warn().Msgf("Product %d not found in cache", productID)
		} else {
			logger.Error().Err(err).Msgf("Error getting product %d from cache", productID)
			return err
		}
	}

	var productData entity.Product
	err = json.Unmarshal([]byte(productCache), &productData)
	if err != nil {
		logger.Error().Err(err).Msgf("Error unmarshalling product %d", productID)
		return err
	}

	if productData.ID == 0 {
		product, err := p.productRepo.GetProductByID(ctx, productID)
		if err != nil {
			logger.Error().Err(err).Msgf("Error getting product by ID %d", productData.ID)
			return err
		}
		productData = *product
	}

	if productData.Stock < quantity {
		logger.Warn().Msgf("Product %d out of stock", productID)
		return fmt.Errorf("product out of stock")
	}

	productData.Stock -= quantity
	_, err = p.productRepo.UpdateProduct(ctx, &productData)
	if err != nil {
		logger.Error().Err(err).Msgf("Error updating product %d", productID)
		return err
	}

	// Delete product from cache
	//err = p.rdb.Del(ctx, key).Err()
	//if err != nil {
	//	logger.Error().Err(err).Msgf("Error deleting product %d from cache", productData.ID)
	//	return err
	//}

	// Write to cache
	err = p.rdb.Set(ctx, key, productData, 0).Err()
	if err != nil {
		logger.Error().Err(err).Msgf("Error setting product %d in cache", productID)
	}

	return nil
}

// ReleaseProductStock releases reserved stock when an order is canceled.
func (p *ProductService) ReleaseProductStock(ctx context.Context, productID int, quantity int) error {
	// Get product from cache
	key := fmt.Sprintf("product:%d", productID)
	productCache, err := p.rdb.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			logger.Warn().Msgf("Product %d not found in cache", productID)
		} else {
			logger.Error().Err(err).Msgf("Error getting product %d from cache", productID)
			return err
		}
	}

	var productData entity.Product
	err = json.Unmarshal([]byte(productCache), &productData)
	if err != nil {
		logger.Error().Err(err).Msgf("Error unmarshalling product %d", productID)
		return err
	}

	if productData.ID == 0 {
		product, err := p.productRepo.GetProductByID(ctx, productID)
		if err != nil {
			logger.Error().Err(err).Msgf("Error getting product by ID %d", productData.ID)
		}
		productData = *product
	}

	productData.Stock += quantity
	_, err = p.productRepo.UpdateProduct(ctx, &productData)
	if err != nil {
		logger.Error().Err(err).Msgf("Error updating product %d", productID)
		return err
	}

	// Delete product from cache
	//err = p.rdb.Del(ctx, key).Err()
	//if err != nil {
	//	logger.Error().Err(err).Msgf("Error deleting product %d from cache", productData.ID)
	//	return err
	//}

	// Write to cache
	err = p.rdb.Set(ctx, key, productData, 0).Err()
	if err != nil {
		logger.Error().Err(err).Msgf("Error setting product %d in cache", productID)
	}

	return nil
}
