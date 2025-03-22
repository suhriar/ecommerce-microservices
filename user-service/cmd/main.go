package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"user-service/cmd/app"
	"user-service/config"
	"user-service/config/cache"
	"user-service/config/database"

	"github.com/gorilla/mux"
)

func main() {
	// Load environment variables
	conf := config.LoadConfig()

	// Initialize DB
	db, err := database.NewMySQLConnection(conf)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Initialize DB
	rdb, err := cache.NewRedisClient(conf)
	if err != nil {
		log.Fatalf("Failed to connect to redis: %v", err)
	}
	defer db.Close()

	// Router setup
	router := mux.NewRouter()
	// API Routes
	apiRouter := router.PathPrefix("/api").Subrouter()
	app.NewApp(apiRouter, db, rdb)

	// Start server
	server := &http.Server{
		Addr:         fmt.Sprintf(":%s", getEnv("PORT", "8000")),
		Handler:      apiRouter,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	// Server in a goroutine
	go func() {
		log.Printf("Server running on port %s\n", getEnv("PORT", "8000"))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Could not listen on %s: %v\n", getEnv("PORT", "8080"), err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Server is shutting down...")
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
