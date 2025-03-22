package usecase

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"user-service/domain"
	repo "user-service/internal/repository/mysql"
	cache "user-service/internal/repository/redis"

	"github.com/go-redis/redis/v8"
	"github.com/golang-jwt/jwt/v4"
	"github.com/rs/zerolog"
)

var logger = zerolog.New(os.Stdout).With().Timestamp().Logger()

type UserUsecase interface {
	GetUserByID(ctx context.Context, id int) (user domain.User, err error)
	CreateUser(ctx context.Context, req domain.User) (user domain.User, err error)
	Login(ctx context.Context, email, password string) (token string, err error)
	ValidateToken(ctx context.Context, email string) (string, error)
}

type userUsecaseImpl struct {
	repo  repo.UserRepository
	cache cache.UserCache
}

func NewUserUsecase(repo repo.UserRepository, cache cache.UserCache) UserUsecase {
	return &userUsecaseImpl{
		repo:  repo,
		cache: cache,
	}
}

type JwtCustomClaims struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	jwt.RegisteredClaims
}

// GetUserByID retrieves a user by ID (stub for now).
func (s *userUsecaseImpl) GetUserByID(ctx context.Context, id int) (user domain.User, err error) {
	user, err = s.repo.GetUserByID(ctx, id)
	if err != nil {
		logger.Error().Err(err).Msgf("Error getting user by ID %d", id)
		return user, err
	}

	return user, nil
}

// CreateUser creates a new user (stub for now).
func (s *userUsecaseImpl) CreateUser(ctx context.Context, req domain.User) (user domain.User, err error) {
	createdUser, err := s.repo.CreateUser(ctx, req)
	if err != nil {
		logger.Error().Err(err).Msg("Error creating user")
		return user, err
	}

	return createdUser, nil
}

//// Login logs in a user with the given email and password.
//func (s *userUsecaseImpl) Login(email string, password string) (user domain.User, err error) {
//	user, err := s.repo.GetUserByEmailAndPassword(email, password)
//	if err != nil {
//		logger.Error().Err(err).Msg("Error logging in user")
//		return nil, err
//	}
//
//	return user, nil
//}

func (s *userUsecaseImpl) Login(ctx context.Context, email, password string) (token string, err error) {
	user, err := s.repo.GetUserByEmailAndPassword(ctx, email, password)
	if err != nil {
		return "", err
	}

	// After validation, generate JWT token
	claims := &JwtCustomClaims{
		Name:  user.Username,
		Email: user.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24)),
		},
	}

	tkn := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	t, err := tkn.SignedString([]byte("secret"))
	if err != nil {
		return "", err
	}

	// Store the JWT token in Redis with the user email as the key
	err = s.cache.SetUserTokenByEmail(ctx, email, t, time.Hour*24) // Set expiration to 24 hours
	if err != nil {
		return "", err
	}

	// Return the user and the generated JWT token
	return t, nil
}

func (s *userUsecaseImpl) ValidateToken(ctx context.Context, token string) (validateToken string, err error) {
	claims := &JwtCustomClaims{}
	fmt.Println(token)
	t, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte("secret"), nil // Kunci yang sama saat membuat token
	})

	if err != nil {
		return validateToken, err
	}

	if !t.Valid {
		return validateToken, fmt.Errorf("invalid token")
	}
	fmt.Println(claims.Email)
	// Retrieve the JWT token from Redis
	validateToken, err = s.cache.GetUserTokenByEmail(ctx, claims.Email)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return "", fmt.Errorf("session not found")
		}
		return "", err
	}
	fmt.Println(validateToken)

	return validateToken, nil
}
