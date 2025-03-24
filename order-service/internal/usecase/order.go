package usecase

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"order-service/domain"
	repo "order-service/internal/repository/mysql"
	cache "order-service/internal/repository/redis"
	"order-service/pkg/utils"

	"github.com/rs/zerolog/log"
	"github.com/segmentio/kafka-go"
)

type OrderUsecase interface {
	CreateOrder(ctx context.Context, req domain.OrderRequest) (order domain.Order, err error)
	UpdateOrder(ctx context.Context, req domain.Order) (updateOrder domain.Order, err error)
	CancelOrder(ctx context.Context, id int) (updatedOrder domain.Order, err error)
	checkProductStock(ctx context.Context, productId int, quantity int) (avail bool, err error)
}

type orderUsecase struct {
	repo              repo.OrderRepository
	cache             cache.OrderCache
	kafkaWriter       *kafka.Writer
	productServiceURL string
	pricingServiceURL string
}

func NewOrderUsecase(repo repo.OrderRepository, cache cache.OrderCache, kafkaWriter *kafka.Writer, productServiceURL, pricingServiceURL string) OrderUsecase {
	return &orderUsecase{
		repo:              repo,
		cache:             cache,
		kafkaWriter:       kafkaWriter,
		productServiceURL: productServiceURL,
		pricingServiceURL: pricingServiceURL,
	}
}

func (u *orderUsecase) CreateOrder(ctx context.Context, req domain.OrderRequest) (createdOrder domain.Order, err error) {

	// get the idempotent key from order
	// validate, err := u.validateIdempotentKey(ctx, req.IdempotentKey)
	// if err != nil {
	// 	return order, err
	// }

	// if !validate {
	// 	return order, errors.New("idempotent key already exists")
	// }

	user, err := utils.GetUserFromContext(ctx)
	if err != nil {
		return createdOrder, err
	}

	var orderReq domain.Order

	availabilityCh := make(chan struct {
		ProductID int
		Available bool
		Error     error
	}, len(req.ProductRequests))

	pricingCh := make(chan struct {
		ProductID  int
		FinalPrice float64
		MarkUp     float64
		Discount   float64
		Error      error
	}, len(req.ProductRequests))

	for _, productRequest := range req.ProductRequests {
		//// check product availability
		//available, err := u.checkProductStock(productRequest.ProductID, productRequest.Quantity)
		//if err != nil {
		//	log.Error().Err(err).Msgf("Error checking product stock for product %d", productRequest.ProductID)
		//	return nil, err
		//}
		//
		//// get pricing
		//pricing, err := u.getPricing(productRequest.ProductID)
		//if err != nil {
		//	log.Error().Err(err).Msgf("Error getting pricing for product %d", productRequest.ProductID)
		//	return nil, err
		//}
		//
		//if !available {
		//	log.Warn().Msgf("Product %d out of stock", productRequest.ProductID)
		//	return nil, fmt.Errorf("product out of stock")
		//}
		//
		//productRequest.FinalPrice = float64(productRequest.Quantity) * pricing.FinalPrice
		//productRequest.MarkUp = float64(productRequest.Quantity) * pricing.Markup
		//productRequest.Discount = float64(productRequest.Quantity) * pricing.Discount

		go func(productID, quantity int) {
			available, err := u.checkProductStock(ctx, productID, quantity)
			availabilityCh <- struct {
				ProductID int
				Available bool
				Error     error
			}{
				ProductID: productRequest.ProductID,
				Available: available,
				Error:     err,
			}
			fmt.Println(availabilityCh)
		}(productRequest.ProductID, productRequest.Quantity)

		go func(productID int) {
			pricing, err := u.getPricing(ctx, productID)
			pricingCh <- struct {
				ProductID  int
				FinalPrice float64
				MarkUp     float64
				Discount   float64
				Error      error
			}{
				ProductID:  productRequest.ProductID,
				FinalPrice: pricing.FinalPrice,
				MarkUp:     pricing.Markup,
				Discount:   pricing.Discount,
				Error:      err,
			}
		}(productRequest.ProductID)
	}

	for range req.ProductRequests {
		availabilityResult := <-availabilityCh
		pricingResult := <-pricingCh

		if availabilityResult.Error != nil {
			log.Error().Err(availabilityResult.Error).Msgf("Error checking product stock for product %d", availabilityResult.ProductID)
			return createdOrder, availabilityResult.Error
		}

		if !availabilityResult.Available {
			log.Warn().Msgf("Product %d out of stock", availabilityResult.ProductID)
			return createdOrder, fmt.Errorf("product out of stock")
		}

		if pricingResult.Error != nil {
			log.Error().Err(pricingResult.Error).Msgf("Error getting pricing for product %d", pricingResult.ProductID)
			return createdOrder, pricingResult.Error
		}

		for _, productRequest := range req.ProductRequests {
			if productRequest.ProductID == availabilityResult.ProductID {
				productRequestReq := domain.ProductRequest{}
				productRequestReq.ProductID = productRequest.ProductID
				productRequestReq.Quantity = productRequest.Quantity
				productRequestReq.FinalPrice = float64(productRequest.Quantity) * pricingResult.FinalPrice
				productRequestReq.MarkUp = float64(productRequest.Quantity) * pricingResult.MarkUp
				productRequestReq.Discount = float64(productRequest.Quantity) * pricingResult.Discount
				orderReq.ProductRequests = append(orderReq.ProductRequests, productRequestReq)
			}
		}
	}

	orderReq.UserID = user.ID
	orderReq.Total = 0
	orderReq.Status = "created"
	orderReq.IdempotentKey = req.IdempotentKey
	for _, productRequest := range orderReq.ProductRequests {
		orderReq.TotalDiscount += productRequest.Discount
		orderReq.TotalMarkUp += productRequest.MarkUp
		orderReq.Total += productRequest.FinalPrice
		orderReq.Quantity += productRequest.Quantity
	}

	createdOrder, err = u.repo.CreateOrder(ctx, orderReq)
	if err != nil {
		log.Error().Err(err).Msg("Error creating order")
		return createdOrder, err
	}

	err = u.publishOrderEvent(ctx, &createdOrder, "created")
	if err != nil {
		return createdOrder, err
	}

	return createdOrder, nil
}

// UpdateOrder updates an existing order
func (u *orderUsecase) UpdateOrder(ctx context.Context, req domain.Order) (updateOrder domain.Order, err error) {
	if req.Status == "paid" {
		// check product availability
		for _, productRequest := range req.ProductRequests {
			available, err := u.checkProductStock(ctx, productRequest.ProductID, productRequest.Quantity)
			if err != nil {
				log.Error().Err(err).Msgf("Error checking product stock for product %d", productRequest.ProductID)
				return updateOrder, err
			}

			if !available {
				log.Warn().Msgf("Product %d out of stock", productRequest.ProductID)
				return updateOrder, fmt.Errorf("product out of stock")
			}
		}
	}
	updateOrder, err = u.repo.UpdateOrder(ctx, req)
	if err != nil {
		log.Error().Err(err).Msg("Error updating order")
		return updateOrder, err
	}

	err = u.publishOrderEvent(ctx, &updateOrder, "updated")
	if err != nil {
		return updateOrder, err
	}

	return updateOrder, nil
}

// CancelOrder cancels an existing order
func (u *orderUsecase) CancelOrder(ctx context.Context, id int) (updatedOrder domain.Order, err error) {
	order, err := u.repo.GetOrderByID(ctx, id)
	if err != nil {
		log.Error().Err(err).Msgf("Error getting order by ID %d", id)
		return updatedOrder, err
	}

	order.Status = "cancelled"

	updatedOrder, err = u.repo.UpdateOrder(ctx, order)
	if err != nil {
		log.Error().Err(err).Msg("Error updating order")
		return updatedOrder, err
	}

	err = u.publishOrderEvent(ctx, &updatedOrder, "cancelled")
	if err != nil {
		return updatedOrder, err
	}

	return updatedOrder, nil
}

func (u *orderUsecase) checkProductStock(ctx context.Context, productId int, quantity int) (avail bool, err error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s/api/products/%d/stock", u.productServiceURL, productId), nil)
	if err != nil {
		return false, err
	}

	token, err := utils.GetTokenFromContext(ctx)
	if err != nil {
		return
	}

	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("product not available")
	}

	var stockData map[string]int
	if err := json.NewDecoder(resp.Body).Decode(&stockData); err != nil {
		return false, err
	}

	availableStock := stockData["stock"]

	return availableStock >= quantity, nil
}

func (u *orderUsecase) getPricing(ctx context.Context, productId int) (pricing domain.Pricing, err error) {
	payload, err := json.Marshal(map[string]int{"product_id": productId})
	if err != nil {
		return pricing, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("%s/pricing", u.pricingServiceURL), bytes.NewBuffer(payload))
	if err != nil {
		return pricing, err
	}

	token, err := utils.GetTokenFromContext(ctx)
	if err != nil {
		return
	}

	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return pricing, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return pricing, fmt.Errorf("failed to get pricing")
	}

	if err := json.NewDecoder(resp.Body).Decode(&pricing); err != nil {
		return pricing, err
	}

	return pricing, nil
}

func (u *orderUsecase) publishOrderEvent(ctx context.Context, order *domain.Order, key string) (err error) {
	orderJSON, err := json.Marshal(order)
	if err != nil {
		return err
	}

	// order-created-1 or order-updated-1
	msg := kafka.Message{
		Key:   []byte(fmt.Sprintf("order-%s-%d", key, order.ID)),
		Value: orderJSON,
	}

	err = u.kafkaWriter.WriteMessages(ctx, msg)
	if err != nil {
		return err
	}

	return nil
}

func (u *orderUsecase) validateIdempotentKey(ctx context.Context, key string) (exist bool, err error) {
	_, err = u.cache.GetIdempotentKeyIsNotExist(ctx, key)
	if err != nil {
		return false, err
	}

	// if it doesn't exist, add the key to the cache with a TTL of 24 hours
	// and return true
	err = u.cache.SetIdempotentKey(ctx, key, 24*time.Hour)
	if err != nil {
		return false, err
	}

	return true, nil
}
