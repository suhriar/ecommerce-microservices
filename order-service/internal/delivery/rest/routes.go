package rest

import (
	"net/http"
	"order-service/internal/delivery/middleware"

	"github.com/gorilla/mux"
)

// RegisterRoutes registers all API routes
func RegisterRoutes(router *mux.Router, orderHandler *OrderHandler) {
	// Logger Middleware
	router.Use(middleware.LoggingMiddleware)

	// API Router
	apiRouter := router.PathPrefix("/api").Subrouter()

	// Health check
	apiRouter.HandleFunc("/health", HealthCheck).Methods("GET")

	// Inisialisasi JWT middleware
	jwtMiddleware := middleware.NewJWTMiddleware()

	// Register order routes
	registerOrderRoutes(apiRouter, orderHandler, jwtMiddleware)
}

func registerOrderRoutes(router *mux.Router, handler *OrderHandler, jwtMiddleware *middleware.JWTMiddleware) {
	// Public routes
	orderRouter := router.PathPrefix("/orders").Subrouter()

	// Protected routes
	protected := orderRouter.PathPrefix("").Subrouter()
	protected.Use(jwtMiddleware.RequireAuth)
	protected.HandleFunc("/orders", handler.CreateOrder).Methods("POST")
	protected.HandleFunc("/orders", handler.UpdateOrder).Methods("PUT")
	protected.HandleFunc("/orders/{id:[0-9]+}", handler.CancelOrder).Methods("DELETE")
}

// HealthCheck handler for the health endpoint
func HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}
