package mysql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"pricing-service/domain"
)

type PricingRepository interface {
	CreatePricingRule(ctx context.Context, rule domain.PricingRule) (err error)
	UpdatePricingRule(ctx context.Context, rule domain.PricingRule) (err error)
	DeletePricingRule(ctx context.Context, productID int) (err error)
	GetPricingRule(ctx context.Context, productID int) (rule domain.PricingRule, err error)
}

type pricingRepository struct {
	db *sql.DB
}

func NewPricingRepository(db *sql.DB) PricingRepository {
	return &pricingRepository{db}
}

// CreatePricingRule creates a new pricing rule in the database
func (r *pricingRepository) CreatePricingRule(ctx context.Context, rule domain.PricingRule) (err error) {
	query := `INSERT INTO pricing_rules (product_id, product_price, default_markup, default_discount, stock_threshold, markup_increase, discount_reduction)
		VALUES (?, ?, ?, ?, ?, ?, ?)`
	_, err = r.db.ExecContext(ctx, query, rule.ProductID, rule.ProductPrice, rule.DefaultMarkup, rule.DefaultDiscount, rule.StockThreshold, rule.MarkupIncrease, rule.DiscountReduction)
	return err
}

// UpdatePricingRule updates an existing pricing rule in the database
func (r *pricingRepository) UpdatePricingRule(ctx context.Context, rule domain.PricingRule) (err error) {
	query := `UPDATE pricing_rules SET product_price = ?, default_markup = ?, default_discount = ?, stock_threshold = ?, markup_increase = ?, discount_reduction = ? WHERE product_id = ?`
	_, err = r.db.ExecContext(ctx, query, rule.ProductPrice, rule.DefaultMarkup, rule.DefaultDiscount, rule.StockThreshold, rule.MarkupIncrease, rule.DiscountReduction, rule.ProductID)
	return err
}

// DeletePricingRule deletes a pricing rule from the database
func (r *pricingRepository) DeletePricingRule(ctx context.Context, productID int) (err error) {
	query := `DELETE FROM pricing_rules WHERE product_id = ?`
	_, err = r.db.ExecContext(ctx, query, productID)
	return err
}

// GetPricingRule fetches the pricing rule for a specific product from the database
func (r *pricingRepository) GetPricingRule(ctx context.Context, productID int) (rule domain.PricingRule, err error) {
	query := `SELECT id, product_id, product_price, default_markup, default_discount, stock_threshold, markup_increase, discount_reduction 
		FROM pricing_rules WHERE product_id = ?`
	row := r.db.QueryRowContext(ctx, query, productID)

	err = row.Scan(&rule.ID, &rule.ProductID, &rule.ProductPrice, &rule.DefaultMarkup, &rule.DefaultDiscount, &rule.StockThreshold, &rule.MarkupIncrease, &rule.DiscountReduction)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return rule, fmt.Errorf("pricing rule not found for product %d", productID)
		}
		return rule, err
	}
	return rule, nil
}
