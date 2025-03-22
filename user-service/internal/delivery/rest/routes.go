package rest

import (
	"net/http"
	"user-service/internal/delivery/middleware"

	"github.com/gorilla/mux"
)

// RegisterRoutes registers all API routes
func RegisterRoutes(router *mux.Router, userHandler *UserHandler) {
	// Logger Middleware
	router.Use(middleware.LoggingMiddleware)

	// API Router
	apiRouter := router.PathPrefix("/api").Subrouter()

	// Health check
	apiRouter.HandleFunc("/health", HealthCheck).Methods("GET")

	// Inisialisasi JWT middleware
	jwtMiddleware := middleware.NewJWTMiddleware()

	// Register user routes
	registerUserRoutes(apiRouter, userHandler, jwtMiddleware)
}

// registerUserRoutes registers user related routes
func registerUserRoutes(router *mux.Router, handler *UserHandler, jwtMiddleware *middleware.JWTMiddleware) {
	userRouter := router.PathPrefix("/users").Subrouter()

	// Public routes
	userRouter.HandleFunc("", handler.CreateUser).Methods("POST")
	userRouter.HandleFunc("/login", handler.Login).Methods("POST")

	// Protected routes
	protected := userRouter.PathPrefix("").Subrouter()
	protected.Use(jwtMiddleware.RequireAuth)

	protected.HandleFunc("/{id:[0-9]+}", handler.GetUserByID).Methods("GET")
	protected.HandleFunc("/validate", handler.ValidateSession).Methods("GET")
}

// HealthCheck handler for the health endpoint
func HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}
