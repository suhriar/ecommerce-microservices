package mysql

import (
	"context"
	"database/sql"
	"order-service/domain"
	"order-service/internal/sharding"
)

type OrderRepository interface {
	GetOrderByID(ctx context.Context, id int) (order domain.Order, err error)
	CreateOrder(ctx context.Context, req domain.Order) (order domain.Order, err error)
	UpdateOrder(ctx context.Context, req domain.Order) (order domain.Order, err error)
	DeleteOrder(ctx context.Context, id, userID int) (err error)
	UpdateOrderStatus(ctx context.Context, id, userID int, status string) (err error)
}

type orderRepository struct {
	dbShards []*sql.DB
	shard    *sharding.ShardRouter
}

func NewOrderRepository(dbShards []*sql.DB, shard *sharding.ShardRouter) OrderRepository {
	return &orderRepository{dbShards, shard}
}

func (r *orderRepository) GetOrderByID(ctx context.Context, id int) (order domain.Order, err error) {
	orderQuery := `SELECT id, user_id, quantity, total, status, total_mark_up, total_discount FROM orders WHERE id = ?`
	productRequestQuery := `SELECT product_id, quantity, mark_up, discount, final_price FROM product_requests WHERE order_id = ?`

	// Loop semua database shard
	for _, db := range r.dbShards {
		err = db.QueryRowContext(ctx, orderQuery, id).Scan(&order.ID, &order.UserID, &order.Quantity, &order.Total, &order.Status, &order.TotalMarkUp, &order.TotalDiscount)
		if err == nil {
			break
		} else if err == sql.ErrNoRows {
			continue
		} else {
			return order, err
		}
	}

	if order.ID == 0 {
		return order, sql.ErrNoRows
	}

	for _, db := range r.dbShards {
		rows, err := db.QueryContext(ctx, productRequestQuery, id)
		if err != nil {
			continue
		}
		defer rows.Close()

		for rows.Next() {
			productRequest := domain.ProductRequest{}
			err := rows.Scan(&productRequest.ProductID, &productRequest.Quantity, &productRequest.MarkUp, &productRequest.Discount, &productRequest.FinalPrice)
			if err != nil {
				return order, err
			}
			order.ProductRequests = append(order.ProductRequests, productRequest)
		}
		break
	}

	return order, nil
}

func (r *orderRepository) CreateOrder(ctx context.Context, req domain.Order) (order domain.Order, err error) {
	dbIndex := r.shard.GetShard(req.UserID)
	db := r.dbShards[dbIndex]

	// Start a transaction
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return order, err
	}

	// Insert order
	orderQuery := `INSERT INTO orders (user_id, quantity, total, status, total_mark_up, total_discount, idempotent_key) VALUES (?, ?, ?, ?, ?, ?, ?)`
	res, err := tx.ExecContext(ctx, orderQuery, req.UserID, req.Quantity, req.Total, req.Status, req.TotalMarkUp, req.TotalDiscount, req.IdempotentKey)
	if err != nil {
		tx.Rollback()
		return order, err
	}

	orderID, err := res.LastInsertId()
	if err != nil {
		tx.Rollback()
		return order, err
	}

	//// Insert product requests
	//productQuery := `
	//	INSERT INTO product_requests (order_id, product_id, quantity, mark_up, discount, final_price)
	//	VALUES (?, ?, ?, ?, ?, ?)`
	//for _, product := range order.ProductRequests {
	//	_, err := tx.Exec(productQuery, orderID, product.ProductID, product.Quantity, product.MarkUp, product.Discount, product.FinalPrice)
	//	if err != nil {
	//		tx.Rollback()
	//		return nil, err
	//	}
	//}

	// Insert product requests with batch
	productQuery := `
		INSERT INTO product_requests (order_id, product_id, quantity, mark_up, discount, final_price)
		VALUES `

	// Build the query
	var values []interface{}
	for _, product := range req.ProductRequests {
		productQuery += "(?, ?, ?, ?, ?, ?),"
		values = append(values, orderID, product.ProductID, product.Quantity, product.MarkUp, product.Discount, product.FinalPrice)
	}

	// Remove the trailing comma
	productQuery = productQuery[:len(productQuery)-1]

	// Execute the query batch insert
	_, err = tx.ExecContext(ctx, productQuery, values...)
	if err != nil {
		tx.Rollback()
		return order, err
	}

	// Commit the transaction
	err = tx.Commit()
	if err != nil {
		return order, err
	}

	order.ID = int(orderID)
	return order, nil
}

func (r *orderRepository) UpdateOrder(ctx context.Context, req domain.Order) (order domain.Order, err error) {
	dbIndex := r.shard.GetShard(req.UserID)
	db := r.dbShards[dbIndex]

	// Start a transaction
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return order, err
	}

	// Update order
	orderQuery := `UPDATE orders SET user_id = ?, quantity = ?, total = ?, status = ?, total_mark_up = ?, total_discount = ? WHERE id = ?`
	_, err = tx.ExecContext(ctx, orderQuery, req.UserID, req.Quantity, req.Total, req.Status, req.TotalMarkUp, req.TotalDiscount, req.ID)
	if err != nil {
		tx.Rollback()
		return order, err
	}

	// Delete existing product requests
	deleteQuery := `DELETE FROM product_requests WHERE order_id = ?`
	_, err = tx.ExecContext(ctx, deleteQuery, req.ID)
	if err != nil {
		tx.Rollback()
		return order, err
	}

	// Insert product requests
	productQuery := `
		INSERT INTO product_requests (order_id, product_id, quantity, mark_up, discount, final_price)
		VALUES (?, ?, ?, ?, ?, ?)`
	for _, product := range req.ProductRequests {
		_, err := tx.ExecContext(ctx, productQuery, req.ID, product.ProductID, product.Quantity, product.MarkUp, product.Discount, product.FinalPrice)
		if err != nil {
			tx.Rollback()
			return order, err
		}
	}

	// Commit the transaction
	err = tx.Commit()
	if err != nil {
		return order, err
	}

	return order, nil
}

func (r *orderRepository) DeleteOrder(ctx context.Context, id, userID int) (err error) {
	dbIndex := r.shard.GetShard(userID)
	db := r.dbShards[dbIndex]

	// Start a transaction
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	// Delete product requests
	productQuery := `DELETE FROM product_requests WHERE order_id = ?`
	_, err = tx.ExecContext(ctx, productQuery, id)
	if err != nil {
		tx.Rollback()
		return err
	}

	// Delete order
	orderQuery := `DELETE FROM orders WHERE id = ?`
	_, err = tx.ExecContext(ctx, orderQuery, id)
	if err != nil {
		tx.Rollback()
		return err
	}

	// Commit the transaction
	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

func (r *orderRepository) UpdateOrderStatus(ctx context.Context, id, userID int, status string) (err error) {
	dbIndex := r.shard.GetShard(userID)
	db := r.dbShards[dbIndex]

	query := `UPDATE orders SET status = ? WHERE id = ?`
	_, err = db.ExecContext(ctx, query, status, id)
	if err != nil {
		return err
	}

	return nil
}
