package app

import (
	"database/sql"

	"product-service/internal/consumer"
	"product-service/internal/delivery/rest"
	repo "product-service/internal/repository/mysql"
	cache "product-service/internal/repository/redis"
	"product-service/internal/usecase"

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"
)

func NewApp(router *mux.Router, db *sql.DB, rdb *redis.Client) {
	productRepo := repo.NewProductRepository(db)
	productCache := cache.NewProductCache(rdb)
	productUsecase := usecase.NewProductService(productRepo, productCache)

	productHandler := rest.NewproductHandler(productUsecase)

	consumer := consumer.NewConsumer(productUsecase)
	go consumer.StartKafkaConsumer()

	rest.RegisterRoutes(router, productHandler)
}
