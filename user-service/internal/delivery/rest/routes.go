package rest

import (
	"net/http"
	"user-service/internal/delivery/middleware"

	"github.com/gorilla/mux"
)

// RegisterRoutes registers all API routes
func RegisterRoutes(router *mux.Router, userHandler *UserHandler) {
	// Health check
	router.HandleFunc("/health", HealthCheck).Methods("GET")

	// Inisialisasi JWT middleware
	jwtMiddleware := middleware.NewJWTMiddleware()

	// Register user routes
	registerUserRoutes(router, userHandler, jwtMiddleware)
}

// registerUserRoutes registers user related routes
func registerUserRoutes(router *mux.Router, handler *UserHandler, jwtMiddleware *middleware.JWTMiddleware) {
	userRouter := router.PathPrefix("/users").Subrouter()

	// Public routes (tidak memerlukan autentikasi)
	userRouter.HandleFunc("", handler.CreateUser).Methods("POST")
	userRouter.HandleFunc("/login", handler.Login).Methods("POST")

	// Protected routes (memerlukan autentikasi)
	protected := userRouter.PathPrefix("").Subrouter()
	protected.Use(jwtMiddleware.RequireAuth) // Menerapkan JWT middleware untuk semua routes di bawah ini

	protected.HandleFunc("/{id:[0-9]+}", handler.GetUserByID).Methods("GET")
	protected.HandleFunc("/validate", handler.ValidateSession).Methods("GET")
}

// HealthCheck handler for the health endpoint
func HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}
