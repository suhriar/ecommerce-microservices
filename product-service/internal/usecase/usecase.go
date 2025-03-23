package usecase

import (
	"context"
	"fmt"
	"time"

	"product-service/domain"
	repo "product-service/internal/repository/mysql"
	cache "product-service/internal/repository/redis"

	"github.com/rs/zerolog/log"
)

type ProductUsecase interface {
	GetProductStock(ctx context.Context, productID int) (stock int, err error)
	ReserveProductStock(ctx context.Context, productID int, quantity int) (err error)
	ReleaseProductStock(ctx context.Context, productID int, quantity int) (err error)
	PreWarmCache(ctx context.Context) (err error)
	PreWarmCacheAsync(ctx context.Context) (err error)
}

type productUsecase struct {
	repo  repo.ProductRepository
	cache cache.ProductCache
}

func NewProductService(repo repo.ProductRepository, cache cache.ProductCache) ProductUsecase {
	return &productUsecase{
		repo:  repo,
		cache: cache,
	}
}

// GetProductStock retrieves the stock for a product.
func (u *productUsecase) GetProductStock(ctx context.Context, productID int) (stock int, err error) {
	// Read from cache
	product, err := u.cache.GetProductByID(ctx, productID)
	if err != nil {
		log.Error().Err(err).Msgf("Error getting stock for product %d from cache", productID)
		return 0, err
	}

	if product.ID == 0 {
		product, err = u.repo.GetProductByID(ctx, productID)
		if err != nil {
			log.Error().Err(err).Msgf("Error getting product by ID %d", productID)
			return 0, err
		}

		// Write to cache
		err = u.cache.SetProduct(ctx, product, 0)
		if err != nil {
			log.Error().Err(err).Msgf("Error setting product %d in cache", productID)
			return 0, err
		}

	}

	return product.Stock, nil
}

// ReserveProductStock reserves stock for an order.
func (u *productUsecase) ReserveProductStock(ctx context.Context, productID int, quantity int) (err error) {
	// Get product from cache
	product, err := u.cache.GetProductByID(ctx, productID)
	if err != nil {
		log.Error().Err(err).Msgf("Error getting stock for product %d from cache", productID)
		return err
	}

	if product.ID == 0 {
		product, err = u.repo.GetProductByID(ctx, productID)
		if err != nil {
			log.Error().Err(err).Msgf("Error getting product by ID %d", product.ID)
			return err
		}
	}

	if product.Stock < quantity {
		log.Warn().Msgf("Product %d out of stock", productID)
		return fmt.Errorf("product out of stock")
	}

	product.Stock -= quantity
	err = u.repo.UpdateProduct(ctx, product)
	if err != nil {
		log.Error().Err(err).Msgf("Error updating product %d", productID)
		return err
	}

	// Write to cache
	err = u.cache.SetProduct(ctx, product, 0)
	if err != nil {
		log.Error().Err(err).Msgf("Error setting product %d in cache", productID)
		return err
	}

	return nil
}

// ReleaseProductStock releases reserved stock when an order is canceled.
func (u *productUsecase) ReleaseProductStock(ctx context.Context, productID int, quantity int) (err error) {
	// Get product from cache
	product, err := u.cache.GetProductByID(ctx, productID)
	if err != nil {
		log.Error().Err(err).Msgf("Error getting stock for product %d from cache", productID)
		return err
	}

	if product.ID == 0 {
		product, err := u.repo.GetProductByID(ctx, productID)
		if err != nil {
			log.Error().Err(err).Msgf("Error getting product by ID %d", product.ID)
		}
	}

	product.Stock += quantity
	err = u.repo.UpdateProduct(ctx, product)
	if err != nil {
		log.Error().Err(err).Msgf("Error updating product %d", productID)
		return err
	}

	// Write to cache
	err = u.cache.SetProduct(ctx, product, 0)
	if err != nil {
		log.Error().Err(err).Msgf("Error setting product %d in cache", productID)
		return err
	}
	return nil
}

// PreWarmCache pre-warms the cache with product data.
func (u *productUsecase) PreWarmCache(ctx context.Context) (err error) {
	products, err := u.repo.GetProducts(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Error getting products")
		return err
	}

	for _, product := range products {
		err = u.cache.SetProduct(ctx, product, 1*time.Minute)
		if err != nil {
			log.Error().Err(err).Msgf("Error setting product %d in cache", product.ID)
			return err
		}
	}

	return nil
}

// PreWarmCacheAsync pre-warms the cache with product data asynchronously.
func (u *productUsecase) PreWarmCacheAsync(ctx context.Context) (err error) {
	products, err := u.repo.GetProducts(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Error getting products")
		return err
	}

	for _, product := range products {
		go func(product domain.Product) {
			err = u.cache.SetProduct(ctx, product, 1*time.Minute)
			if err != nil {
				log.Error().Err(err).Msgf("Error setting product %d in cache", product.ID)
			}
		}(product)
	}

	return nil
}
