package middleware

import (
	"context"
	"fmt"
	"net/http"
	"product-service/config"
	"product-service/domain"
	"product-service/pkg/utils"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v4"
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
		ctx := context.WithValue(r.Context(), domain.UserNameKey, claims.Username)
		ctx = context.WithValue(ctx, domain.UserEmailKey, claims.Email)
		ctx = context.WithValue(ctx, domain.UserIDlKey, claims.UserID)
		ctx = context.WithValue(ctx, domain.AuthorizationKey, token)

		// Lanjutkan request dengan context yang telah diperbarui
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequireAuth adalah middleware yang memastikan user sudah terautentikasi
func (m *JWTMiddleware) RequireAuth(next http.Handler) http.Handler {
	return m.Middleware(next)
}
