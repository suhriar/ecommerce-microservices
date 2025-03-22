package rest

import (
	"encoding/json"
	"net/http"
	"strconv"

	"user-service/domain"
	"user-service/internal/usecase"

	"github.com/gorilla/mux"
)

type UserHandler struct {
	userUsecase usecase.UserUsecase
}

// NewUserHandler creates a new instance of UserHandler
func NewUserHandler(userUsecase usecase.UserUsecase) *UserHandler {
	return &UserHandler{userUsecase: userUsecase}
}

// GetUserByID retrieves a user by ID --> /users/{id}
func (h *UserHandler) GetUserByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.respondWithJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid ID"})
		return
	}

	user, err := h.userUsecase.GetUserByID(r.Context(), id)
	if err != nil {
		h.respondWithJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	h.respondWithJSON(w, http.StatusOK, user)
}

// CreateUser creates a new user --> /users
func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var user domain.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		h.respondWithJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid request payload"})
		return
	}

	createdUser, err := h.userUsecase.CreateUser(r.Context(), user)
	if err != nil {
		h.respondWithJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	h.respondWithJSON(w, http.StatusOK, createdUser)
}

// Login logs in a user --> /users/login
func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	var login struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&login); err != nil {
		h.respondWithJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid request payload"})
		return
	}

	token, err := h.userUsecase.Login(r.Context(), login.Email, login.Password)
	if err != nil {
		h.respondWithJSON(w, http.StatusUnauthorized, map[string]string{"error": err.Error()})
		return
	}

	h.respondWithJSON(w, http.StatusOK, map[string]string{"token": token})
}

// ValidateSession validates a session token --> /users/validate
func (h *UserHandler) ValidateSession(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("Authorization")
	if token == "" {
		h.respondWithJSON(w, http.StatusUnauthorized, map[string]string{"error": "Unauthorized"})
		return
	}

	validateToken, err := h.userUsecase.ValidateToken(r.Context(), token)
	if err != nil || validateToken != token {
		h.respondWithJSON(w, http.StatusUnauthorized, map[string]string{"error": "Unauthorized"})
		return
	}

	h.respondWithJSON(w, http.StatusOK, map[string]string{"message": "Session is valid"})
}

// Helper function to respond with JSON
func (h *UserHandler) respondWithJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(payload)
}
