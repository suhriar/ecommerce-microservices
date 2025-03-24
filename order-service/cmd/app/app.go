package app

import (
	"database/sql"

	"order-service/internal/delivery/rest"
	repo "order-service/internal/repository/mysql"
	cache "order-service/internal/repository/redis"
	shard "order-service/internal/sharding"
	"order-service/internal/usecase"

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"
	"github.com/segmentio/kafka-go"
)

func NewApp(router *mux.Router, dbShards []*sql.DB, rdb *redis.Client, kafkaWriter *kafka.Writer) {
	orderShard := shard.NewShardRouter(len(dbShards))
	orderRepo := repo.NewOrderRepository(dbShards, orderShard)
	orderCache := cache.NewOrderCache(rdb)
	orderUsecase := usecase.NewOrderUsecase(orderRepo, orderCache, kafkaWriter, "http://localhost:8081", "http://localhost:8083")

	orderHandler := rest.NewOrderHandler(orderUsecase)

	rest.RegisterRoutes(router, orderHandler)
}
