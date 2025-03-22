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
	"user-service/pkg/utils"

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
	tokenString, err := utils.GenerateJWT(user)
	if err != nil {
		return "", err
	}
	// Store the JWT token in Redis with the user email as the key
	err = s.cache.SetUserTokenByEmail(ctx, email, tokenString, time.Hour*24) // Set expiration to 24 hours
	if err != nil {
		return "", err
	}

	// Return the user and the generated JWT token
	return tokenString, nil
}

func (s *userUsecaseImpl) ValidateToken(ctx context.Context, email string) (validateToken string, err error) {
	// Retrieve the JWT token from Redis
	validateToken, err = s.cache.GetUserTokenByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return "", fmt.Errorf("session not found")
		}
		return "", err
	}

	return validateToken, nil
}
