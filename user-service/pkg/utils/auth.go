package utils

import (
	"context"
	"errors"
	"user-service/domain"
)

type UserAuth struct {
	ID       int    `json:"ud"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

// GetUserFromContext mengambil username dan email dari context
func GetUserFromContext(ctx context.Context) (user UserAuth, err error) {
	username, ok1 := ctx.Value(domain.UserNameKey).(string)
	email, ok2 := ctx.Value(domain.UserEmailKey).(string)
	id, ok3 := ctx.Value(domain.UserIDlKey).(int)

	if !ok1 || !ok2 || !ok3 {
		return user, errors.New("could not get user data from context")
	}

	user.ID = id
	user.Username = username
	user.Email = email

	return user, nil
}
