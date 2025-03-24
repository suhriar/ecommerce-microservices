package utils

import (
	"fmt"
	"time"
	"user-service/config"
	"user-service/domain"

	"github.com/golang-jwt/jwt/v4"
)

type JwtCustomClaims struct {
	UserID   int    `json:"user_id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	jwt.RegisteredClaims
}

// GenerateJWT membuat token JWT baru
func GenerateJWT(user domain.User) (tokenString string, err error) {
	// Buat claims dengan data user
	claims := &JwtCustomClaims{
		UserID:   user.ID,
		Username: user.Username,
		Email:    user.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24)),
		},
	}
	// Buat token dengan signing method HMAC SHA256
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Tanda tangani token
	tokenString, err = token.SignedString([]byte(config.AppConfig.Jwt.Secret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// ValidateJWT memvalidasi token JWT
func ValidateJWT(tokenString string) (claims JwtCustomClaims, err error) {
	t, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte("secret"), nil // Kunci yang sama saat membuat token
	})

	if err != nil {
		return claims, err
	}

	if !t.Valid {
		return claims, fmt.Errorf("invalid token")
	}

	return claims, nil
}
