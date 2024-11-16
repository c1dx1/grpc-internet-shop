package repository

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v4/pgxpool"
	"golang.org/x/crypto/bcrypt"
	"internet-shop/shared/models"
)

type UserRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) SignInUser(ctx context.Context, email, password string) (int64, error) {
	conn, err := r.db.Acquire(ctx)
	if err != nil {
		return 0, err
	}
	defer conn.Release()

	var user models.User
	err = conn.QueryRow(ctx, "SELECT id, password_hash FROM users WHERE email=$1", email).Scan(&user.ID, &user.Password)
	if err != nil {
		return 0, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return 0, err
	}

	return user.ID, nil
}

func (r *UserRepository) SignUpUser(ctx context.Context, email, password string) (int64, error) {
	conn, err := r.db.Acquire(ctx)
	if err != nil {
		return 0, err
	}
	defer conn.Release()

	var flag bool

	err = conn.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM users WHERE email=$1)", email).Scan(&flag)
	if flag {
		return 0, fmt.Errorf("User is alredy exists")
	}

	var userID int64
	err = conn.QueryRow(ctx, "INSERT INTO users(email, password_hash) VALUES($1, $2) RETURNING id", email, password).Scan(&userID)
	if err != nil {
		return 0, err
	}

	return userID, nil
}

func (r *UserRepository) GetEmailById(ctx context.Context, userID int64) (string, error) {
	conn, err := r.db.Acquire(ctx)
	if err != nil {
		return "", err
	}
	defer conn.Release()

	var email string

	err = conn.QueryRow(ctx, "SELECT email FROM users WHERE id=$1", userID).Scan(&email)
	if err != nil {
		return "", err
	}

	return email, nil
}
