package middleware

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
	"user-service/config"
	"user-service/pkg/utils"

	"github.com/golang-jwt/jwt/v4"
)

// contextKey adalah tipe untuk key dalam context
type contextKey string

const (
	UserIDlKey   contextKey = "user_id"
	UserNameKey  contextKey = "username"
	UserEmailKey contextKey = "email"
)

// JWTMiddleware adalah middleware untuk JWT authentication
type JWTMiddleware struct {
	secretKey []byte
}

// NewJWTMiddleware membuat instance baru dari JWTMiddleware
func NewJWTMiddleware() *JWTMiddleware {
	return &JWTMiddleware{
		secretKey: []byte(config.AppConfig.Jwt.Secret),
	}
}

type JwtCustomClaims struct {
	UserID   int    `json:"user_id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	jwt.RegisteredClaims
}

type UserAuth struct {
	ID       int    `json:"ud"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

// Middleware mengecek JWT token untuk endpoint yang terproteksi
func (m *JWTMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			utils.RespondWithJSON(w, http.StatusUnauthorized, map[string]string{"message": "Authorization header is required"})
			return
		}

		// Format harus "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			utils.RespondWithJSON(w, http.StatusUnauthorized, map[string]string{"message": "Authorization header is required"})
			return
		}

		tokenString := parts[1]

		// Parse token dengan custom claims
		claims := &JwtCustomClaims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return m.secretKey, nil
		})

		if err != nil || !token.Valid {
			utils.RespondWithJSON(w, http.StatusUnauthorized, map[string]string{"message": "Invalid token", "error": err.Error()})
			return
		}

		// Cek apakah token sudah expired
		if claims.ExpiresAt.Time.Before(time.Now()) {
			utils.RespondWithJSON(w, http.StatusUnauthorized, map[string]string{"message": "Token expired"})
			http.Error(w, "Token expired", http.StatusUnauthorized)
			return
		}

		// Tambahkan data user ke context
		ctx := context.WithValue(r.Context(), UserNameKey, claims.Username)
		ctx = context.WithValue(ctx, UserEmailKey, claims.Email)
		ctx = context.WithValue(ctx, UserIDlKey, claims.UserID)

		// Lanjutkan request dengan context yang telah diperbarui
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequireAuth adalah middleware yang memastikan user sudah terautentikasi
func (m *JWTMiddleware) RequireAuth(next http.Handler) http.Handler {
	return m.Middleware(next)
}

// GetUserFromContext mengambil username dan email dari context
func GetUserFromContext(ctx context.Context) (user UserAuth, err error) {
	username, ok1 := ctx.Value(UserNameKey).(string)
	email, ok2 := ctx.Value(UserEmailKey).(string)
	id, ok3 := ctx.Value(UserIDlKey).(int)

	if !ok1 || !ok2 || !ok3 {
		return user, errors.New("could not get user data from context")
	}

	user.ID = id
	user.Username = username
	user.Email = email

	return user, nil
}
