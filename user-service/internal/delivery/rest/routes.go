package rest

import (
	"net/http"

	"github.com/gorilla/mux"
)

// RegisterRoutes registers all API routes
func RegisterRoutes(router *mux.Router, userHandler *UserHandler) {
	// Health check
	router.HandleFunc("/health", HealthCheck).Methods("GET")

	// Register user routes
	registerUserRoutes(router, userHandler)
}

// registerUserRoutes registers user related routes
func registerUserRoutes(router *mux.Router, handler *UserHandler) {
	userRouter := router.PathPrefix("/users").Subrouter()
	userRouter.HandleFunc("", handler.CreateUser).Methods("POST")
	userRouter.HandleFunc("/{id:[0-9]+}", handler.GetUserByID).Methods("GET")
	userRouter.HandleFunc("/login", handler.Login).Methods("POST")
	userRouter.HandleFunc("/validate", handler.ValidateSession).Methods("GET")
}

// HealthCheck handler for the health endpoint
func HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}
