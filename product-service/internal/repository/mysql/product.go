package mysql

import (
	"context"
	"database/sql"
	"product-service/domain"
)

type ProductRepository interface {
	GetProductByID(ctx context.Context, id int) (product domain.Product, err error)
	CreateProduct(ctx context.Context, req domain.Product) (product domain.Product, err error)
	UpdateProduct(ctx context.Context, req domain.Product) (err error)
	DeleteProduct(ctx context.Context, id int) (err error)
	GetProducts(ctx context.Context) (products []domain.Product, err error)
}

type productRepository struct {
	db *sql.DB
}

func NewProductRepository(db *sql.DB) ProductRepository {
	return &productRepository{db}
}

func (r *productRepository) GetProductByID(ctx context.Context, id int) (product domain.Product, err error) {
	query := `SELECT id, name, description, price, stock FROM products WHERE id = ?`
	err = r.db.QueryRowContext(ctx, query, id).Scan(&product.ID, &product.Name, &product.Description, &product.Price, &product.Stock)
	if err != nil {
		return
	}

	return
}

func (r *productRepository) CreateProduct(ctx context.Context, req domain.Product) (product domain.Product, err error) {
	query := `INSERT INTO products (name, description, price, stock) VALUES (?, ?, ?, ?)`
	res, err := r.db.ExecContext(ctx, query, req.Name, req.Description, req.Price, req.Stock)
	if err != nil {
		return
	}

	id, err := res.LastInsertId()
	if err != nil {
		return
	}

	product = domain.Product{
		ID:          int(id),
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		Stock:       req.Stock,
	}

	return
}

func (r *productRepository) UpdateProduct(ctx context.Context, req domain.Product) (err error) {
	query := `UPDATE products SET name = ?, description = ?, price = ?, stock = ? WHERE id = ?`
	_, err = r.db.ExecContext(ctx, query, req.Name, req.Description, req.Price, req.Stock, req.ID)
	if err != nil {
		return
	}
	return
}

func (r *productRepository) DeleteProduct(ctx context.Context, id int) (err error) {
	query := `DELETE FROM products WHERE id = ?`
	_, err = r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}
	return nil
}

func (r *productRepository) GetProducts(ctx context.Context) (products []domain.Product, err error) {
	query := `SELECT id, name, description, price, stock FROM products`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		var product domain.Product
		err = rows.Scan(&product.ID, &product.Name, &product.Description, &product.Price, &product.Stock)
		if err != nil {
			return
		}
		products = append(products, product)
	}

	return
}
