package rest

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"user-service/domain"
	"user-service/internal/delivery/middleware"
	"user-service/internal/usecase"
	"user-service/pkg/utils"

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
		utils.RespondWithJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid ID"})
		return
	}

	user, err := h.userUsecase.GetUserByID(r.Context(), id)
	if err != nil {
		utils.RespondWithJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, user)
}

// CreateUser creates a new user --> /users
func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var user domain.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		utils.RespondWithJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid request payload"})
		return
	}

	createdUser, err := h.userUsecase.CreateUser(r.Context(), user)
	if err != nil {
		utils.RespondWithJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, createdUser)
}

// Login logs in a user --> /users/login
func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	var login struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&login); err != nil {
		utils.RespondWithJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid request payload"})
		return
	}

	token, err := h.userUsecase.Login(r.Context(), login.Email, login.Password)
	if err != nil {
		utils.RespondWithJSON(w, http.StatusUnauthorized, map[string]string{"error": err.Error()})
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, map[string]string{"token": token})
}

// ValidateSession validates a session token --> /users/validate
func (h *UserHandler) ValidateSession(w http.ResponseWriter, r *http.Request) {
	tokenHeader := r.Header.Get("Authorization")
	if tokenHeader == "" {
		utils.RespondWithJSON(w, http.StatusUnauthorized, map[string]string{"error": "Unauthorized"})
		return
	}

	user, err := middleware.GetUserFromContext(r.Context())
	if err != nil {
		utils.RespondWithJSON(w, http.StatusUnauthorized, map[string]string{"error": "Unauthorized"})
		return
	}

	tokenString := strings.TrimPrefix(tokenHeader, "Bearer ")

	validateToken, err := h.userUsecase.ValidateToken(r.Context(), user.Email)
	if err != nil || validateToken != tokenString {
		utils.RespondWithJSON(w, http.StatusUnauthorized, map[string]string{"error": "Unauthorized"})
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, map[string]string{"message": "Session is valid"})
}
