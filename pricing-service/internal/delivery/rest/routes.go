package rest

import (
	"net/http"
	"pricing-service/internal/delivery/middleware"

	"github.com/gorilla/mux"
)

// RegisterRoutes registers all API routes
func RegisterRoutes(router *mux.Router, pricingHandler *PricingHandler) {
	// Logger Middleware
	router.Use(middleware.LoggingMiddleware)

	// API Router
	apiRouter := router.PathPrefix("/api").Subrouter()

	// Health check
	apiRouter.HandleFunc("/health", HealthCheck).Methods("GET")

	// Inisialisasi JWT middleware
	jwtMiddleware := middleware.NewJWTMiddleware()

	// Register Pricing routes
	registerPricingRoutes(apiRouter, pricingHandler, jwtMiddleware)
}

// registerUserRoutes registers user related routes
func registerPricingRoutes(router *mux.Router, handler *PricingHandler, jwtMiddleware *middleware.JWTMiddleware) {
	// Public routes
	PricingRouter := router.PathPrefix("/pricing").Subrouter()

	// Protected routes
	protected := PricingRouter.PathPrefix("").Subrouter()
	protected.Use(jwtMiddleware.RequireAuth)
	protected.HandleFunc("", handler.GetPricing).Methods("POST")

}

// HealthCheck handler for the health endpoint
func HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}
