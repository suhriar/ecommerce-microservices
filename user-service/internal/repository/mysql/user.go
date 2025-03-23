package mysql

import (
	"context"
	"database/sql"

	"user-service/domain"
)

type UserRepository interface {
	GetUserByID(ctx context.Context, id int) (user domain.User, err error)
	CreateUser(ctx context.Context, req domain.User) (user domain.User, err error)
	GetUserByEmail(ctx context.Context, email string) (user domain.User, err error)
}

type userRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) UserRepository {
	return &userRepository{db}
}

func (r *userRepository) GetUserByID(ctx context.Context, id int) (user domain.User, err error) {
	query := `SELECT id, username, email, password FROM users WHERE id = ?`
	err = r.db.QueryRowContext(ctx, query, id).Scan(&user.ID, &user.Username, &user.Email, &user.Password)
	if err != nil {
		return user, err
	}

	return user, nil
}

func (r *userRepository) CreateUser(ctx context.Context, req domain.User) (user domain.User, err error) {
	query := `INSERT INTO users (username, email, password) VALUES (?, ?, ?)`
	res, err := r.db.ExecContext(ctx, query, req.Username, req.Email, req.Password)
	if err != nil {
		return user, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return user, err
	}

	user = domain.User{
		ID:       int(id),
		Username: req.Username,
		Email:    req.Email,
		Password: req.Password, // Hati-hati menyertakan password dalam respons
	}

	return user, nil
}

func (r *userRepository) GetUserByEmail(ctx context.Context, email string) (user domain.User, err error) {
	query := `SELECT id, username, email, password FROM users WHERE email = ?`
	err = r.db.QueryRowContext(ctx, query, email).Scan(&user.ID, &user.Username, &user.Email, &user.Password)
	if err != nil {
		return user, err
	}

	return user, nil
}
