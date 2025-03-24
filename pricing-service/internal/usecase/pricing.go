package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"pricing-service/domain"
	repo "pricing-service/internal/repository/mysql"
	cache "pricing-service/internal/repository/redis"
	"pricing-service/pkg/utils"
)

type PricingUsecase interface {
	CalculatePricing(ctx context.Context, productID int) (price domain.Pricing, err error)
}

type pricingUsecase struct {
	repo              repo.PricingRepository
	cache             cache.PricingCache
	productServiceURL string
}

func NewPricingUsecase(repo repo.PricingRepository, cache cache.PricingCache, productServiceURL string) PricingUsecase {
	return &pricingUsecase{
		repo:              repo,
		cache:             cache,
		productServiceURL: productServiceURL,
	}
}

// CalculatePricing calculates the final price for a product based on pricing rules.
func (u *pricingUsecase) CalculatePricing(ctx context.Context, productID int) (price domain.Pricing, err error) {
	//  Get the pricing rule for the product
	pricingRule, err := u.cache.GetPricingRule(ctx, productID)
	if err != nil {
		return
	}

	if pricingRule.ID == 0 {
		pricingRule, err = u.repo.GetPricingRule(ctx, productID)
		if err != nil {
			return price, fmt.Errorf("pricing rule not found for product %d", productID)
		}

		err = u.cache.SetProduct(ctx, pricingRule, 0)
		if err != nil {
			return
		}

	}

	// Step 2: Check product stock
	available, err := u.checkProductStock(ctx, productID)
	if err != nil {
		return
	}

	// Step 3: Calculate price based on stock
	markup := pricingRule.DefaultMarkup
	discount := pricingRule.DefaultDiscount

	// If stock is below the threshold, apply price adjustments
	if available < pricingRule.StockThreshold {
		markup += pricingRule.MarkupIncrease
		discount -= pricingRule.DiscountReduction
	}

	// Step 4: Calculate the final price
	productPrice := pricingRule.ProductPrice
	finalPrice := productPrice * (1 + markup) * (1 - discount)

	// Step 5: Return the calculated pricing
	price = domain.Pricing{
		ProductID:  productID,
		Markup:     markup,
		Discount:   discount,
		FinalPrice: finalPrice,
	}
	return price, nil
}

// checkProductStock checks if the product is available in the required quantity.
func (u *pricingUsecase) checkProductStock(ctx context.Context, productID int) (availableStock int, err error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s/products/%d/stock", u.productServiceURL, productID), nil)
	if err != nil {
		return 0, err
	}

	token, err := utils.GetTokenFromContext(ctx)
	if err != nil {
		return
	}

	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("product not available")
	}

	var stockData map[string]int
	if err := json.NewDecoder(resp.Body).Decode(&stockData); err != nil {
		return 0, err
	}

	availableStock = stockData["stock"]
	return availableStock, nil
}
