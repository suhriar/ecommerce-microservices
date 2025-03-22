package app

import (
	"database/sql"

	"user-service/internal/delivery/rest"
	repo "user-service/internal/repository/mysql"
	cache "user-service/internal/repository/redis"
	"user-service/internal/usecase"

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"
)

func NewApp(router *mux.Router, db *sql.DB, rdb *redis.Client) {
	userRepo := repo.NewUserRepository(db)
	userCache := cache.NewUserCache(rdb)
	userUsecase := usecase.NewUserUsecase(userRepo, userCache)

	userHandler := rest.NewUserHandler(userUsecase)

	rest.RegisterRoutes(router, userHandler)
	// return
}
