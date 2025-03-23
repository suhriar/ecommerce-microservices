package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"

	"user-service/domain"
	repo "user-service/internal/repository/mysql"
	cache "user-service/internal/repository/redis"
	"user-service/pkg/utils"

	"github.com/go-redis/redis/v8"
	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/bcrypt"
)

type UserUsecase interface {
	GetUserByID(ctx context.Context, id int) (user domain.User, err error)
	CreateUser(ctx context.Context, req domain.User) (user domain.User, err error)
	Login(ctx context.Context, email, password string) (token string, err error)
	ValidateToken(ctx context.Context, email string) (string, error)
}

type userUsecase struct {
	repo  repo.UserRepository
	cache cache.UserCache
}

func NewUserUsecase(repo repo.UserRepository, cache cache.UserCache) UserUsecase {
	return &userUsecase{
		repo:  repo,
		cache: cache,
	}
}

// GetUserByID retrieves a user by ID (stub for now).
func (u *userUsecase) GetUserByID(ctx context.Context, id int) (user domain.User, err error) {
	user, err = u.repo.GetUserByID(ctx, id)
	if err != nil {
		log.Error().Err(err).Msgf("Error getting user by ID %d", id)
		return user, err
	}

	return user, nil
}

// CreateUser creates a new user (stub for now).
func (u *userUsecase) CreateUser(ctx context.Context, req domain.User) (user domain.User, err error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Error().Err(err).Msg("Error hashing password")
		return user, err
	}

	req.Password = string(hashedPassword)

	createdUser, err := u.repo.CreateUser(ctx, req)
	if err != nil {
		log.Error().Err(err).Msg("Error creating user")
		return user, err
	}

	return createdUser, nil
}

//// Login logs in a user with the given email and password.
//func (u *userUsecase) Login(email string, password string) (user domain.User, err error) {
//	user, err := u.repo.GetUserByEmailAndPassword(email, password)
//	if err != nil {
//		log.Error().Err(err).Msg("Error logging in user")
//		return nil, err
//	}
//
//	return user, nil
//}

func (u *userUsecase) Login(ctx context.Context, email, password string) (token string, err error) {
	user, err := u.repo.GetUserByEmail(ctx, email)
	if err != nil {
		return "", err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return "", errors.New("invalid email or password")
	}

	// After validation, generate JWT token
	tokenString, err := utils.GenerateJWT(user)
	if err != nil {
		return "", err
	}
	// Store the JWT token in Redis with the user email as the key
	err = u.cache.SetUserTokenByEmail(ctx, email, tokenString, time.Hour*24) // Set expiration to 24 hours
	if err != nil {
		return "", err
	}

	// Return the user and the generated JWT token
	return tokenString, nil
}

func (u *userUsecase) ValidateToken(ctx context.Context, email string) (validateToken string, err error) {
	// Retrieve the JWT token from Redis
	validateToken, err = u.cache.GetUserTokenByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return "", fmt.Errorf("session not found")
		}
		return "", err
	}

	return validateToken, nil
}
