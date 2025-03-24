// File: config/database/mysql.go
package database

import (
	"database/sql"
	"fmt"
	"time"

	"order-service/config"

	_ "github.com/go-sql-driver/mysql"
)

func NewMySQLShardConnection(cfg *config.Config) (dbs []*sql.DB, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("error occured %+v", r)
		}
	}()
	dbConfig := cfg.MySql
	dbConfig2 := cfg.MySql2
	dbConfig3 := cfg.MySql3

	shardConfigs := []string{
		fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true",
			dbConfig.User, dbConfig.Password, dbConfig.Host, dbConfig.Port, dbConfig.Name),
		fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true",
			dbConfig2.User, dbConfig2.Password, dbConfig2.Host, dbConfig2.Port, dbConfig2.Name),
		fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true",
			dbConfig3.User, dbConfig3.Password, dbConfig3.Host, dbConfig3.Port, dbConfig3.Name),
	}

	for _, dsn := range shardConfigs {
		db, err := sql.Open("mysql", dsn)
		if err != nil {
			return nil, fmt.Errorf("failed to connect to shard %s: %w", dbConfig.Name, err)
		}

		// Configure connection pool
		db.SetMaxIdleConns(10)
		db.SetMaxOpenConns(100)
		db.SetConnMaxLifetime(time.Hour)

		// Verify connection
		if err = db.Ping(); err != nil {
			return nil, fmt.Errorf("failed to ping shard %s: %w", dbConfig.Name, err)
		}

		dbs = append(dbs, db)
	}

	return dbs, nil
}
