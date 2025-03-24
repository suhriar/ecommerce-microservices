package migration

import (
	"database/sql"
	"fmt"
	"time"
)

// AutoMigrateOrders creates the orders table if it does not exist and sets AUTO_INCREMENT for sharding.
func AutoMigrateOrders(retries int, dbs ...*sql.DB) error {
	query := `
		CREATE TABLE IF NOT EXISTS orders (
			id INT AUTO_INCREMENT PRIMARY KEY,
			user_id INT NOT NULL,
			quantity INT NOT NULL,
			total DOUBLE NOT NULL,
			total_mark_up DOUBLE NOT NULL,
			total_discount DOUBLE NOT NULL,
			status VARCHAR(20) NOT NULL,
			idempotent_key VARCHAR(255) UNIQUE NOT NULL
		);
	`

	for shardIndex, db := range dbs {
		_, err := db.Exec(query)
		if err != nil {
			// Retry jika gagal
			for i := 0; i < retries; i++ {
				time.Sleep(1 * time.Second)
				_, err = db.Exec(query)
				if err == nil {
					break
				}
			}
		}

		// Atur AUTO_INCREMENT
		autoIncrementQuery := fmt.Sprintf("ALTER TABLE orders AUTO_INCREMENT = %d;", (shardIndex+1)*1000001)
		_, err = db.Exec(autoIncrementQuery)
		if err != nil {
			fmt.Println("Error setting AUTO_INCREMENT for orders:", err)
		}
	}
	return nil
}

// AutoMigrateProductRequests creates the product_requests table if it does not exist and sets AUTO_INCREMENT for sharding.
func AutoMigrateProductRequests(retries int, dbs ...*sql.DB) error {
	query := `
		CREATE TABLE IF NOT EXISTS product_requests (
			id INT AUTO_INCREMENT PRIMARY KEY,
			order_id INT NOT NULL,
			product_id INT NOT NULL,
			quantity INT NOT NULL,
			mark_up DOUBLE NOT NULL,
			discount DOUBLE NOT NULL,
			final_price DOUBLE NOT NULL,
			FOREIGN KEY (order_id) REFERENCES orders(id) ON DELETE CASCADE
		);
	`
	for shardIndex, db := range dbs {
		_, err := db.Exec(query)
		if err != nil {
			// Retry jika gagal
			for i := 0; i < retries; i++ {
				time.Sleep(1 * time.Second)
				_, err = db.Exec(query)
				if err == nil {
					break
				}
			}
		}

		// Atur AUTO_INCREMENT
		autoIncrementQuery := fmt.Sprintf("ALTER TABLE product_requests AUTO_INCREMENT = %d;", (shardIndex+1)*1000001)
		_, err = db.Exec(autoIncrementQuery)
		if err != nil {
			fmt.Println("Error setting AUTO_INCREMENT for product_requests:", err)
		}
	}
	return nil
}
