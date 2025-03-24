package app

import (
	"database/sql"

	"pricing-service/internal/delivery/rest"
	repo "pricing-service/internal/repository/mysql"
	cache "pricing-service/internal/repository/redis"
	"pricing-service/internal/usecase"

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"
)

func NewApp(router *mux.Router, db *sql.DB, rdb *redis.Client) {
	pricingRepo := repo.NewPricingRepository(db)
	pricingCache := cache.NewPricingCache(rdb)
	pricingUsecase := usecase.NewPricingUsecase(pricingRepo, pricingCache, "http://localhost:8001")

	pricingHandler := rest.NewPricingHandler(pricingUsecase)

	rest.RegisterRoutes(router, pricingHandler)
}
