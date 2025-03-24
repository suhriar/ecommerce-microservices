package rest

import (
	"net/http"
	"product-service/internal/delivery/middleware"

	"github.com/gorilla/mux"
)

// RegisterRoutes registers all API routes
func RegisterRoutes(router *mux.Router, productHandler *ProductHandler) {
	// Logger Middleware
	router.Use(middleware.LoggingMiddleware)

	// API Router
	apiRouter := router.PathPrefix("/api").Subrouter()

	// Health check
	apiRouter.HandleFunc("/health", HealthCheck).Methods("GET")

	// Inisialisasi JWT middleware
	jwtMiddleware := middleware.NewJWTMiddleware()

	// Register product routes
	registerProductRoutes(apiRouter, productHandler, jwtMiddleware)
}

// registerUserRoutes registers user related routes
func registerProductRoutes(router *mux.Router, handler *ProductHandler, jwtMiddleware *middleware.JWTMiddleware) {
	// Public routes
	productRouter := router.PathPrefix("/products").Subrouter()

	// Protected routes
	protected := productRouter.PathPrefix("").Subrouter()
	protected.Use(jwtMiddleware.RequireAuth)
	protected.HandleFunc("/{id:[0-9]+}/stock", handler.GetProductStock).Methods("GET")
	protected.HandleFunc("/reserve", handler.ReserveProductStock).Methods("POST")
	protected.HandleFunc("/release", handler.ReleaseProductStock).Methods("POST")
	protected.HandleFunc("/warmup-cache", handler.PreWarmupCache).Methods("GET")

}

// HealthCheck handler for the health endpoint
func HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}
