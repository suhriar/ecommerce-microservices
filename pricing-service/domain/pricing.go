package domain

type PricingRule struct {
	ID                int     `json:"id"`
	ProductID         int     `json:"product_id"`
	ProductPrice      float64 `json:"product_price"`
	DefaultMarkup     float64 `json:"default_markup"`
	DefaultDiscount   float64 `json:"default_discount"`
	StockThreshold    int     `json:"stock_threshold"`    // If stock is less than this, apply price adjustments
	MarkupIncrease    float64 `json:"markup_increase"`    // Increase markup by this percentage
	DiscountReduction float64 `json:"discount_reduction"` // Reduce discount by this percentage
}

// Pricing represents the pricing data for a product.
type Pricing struct {
	ProductID  int     `json:"product_id"`
	Markup     float64 `json:"markup"`      // Markup percentage
	Discount   float64 `json:"discount"`    // Discount percentage
	FinalPrice float64 `json:"final_price"` // Calculated final price
}
