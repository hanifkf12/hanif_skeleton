package user

import (
	"context"
	"github.com/hanifkf12/hanif_skeleton/internal/entity"
)

func (u *userRepository) CreateUser(ctx context.Context, user entity.CreateUserRequest) (int64, error) {
	// Define insert query
	query := "INSERT INTO users (username, email, password) VALUES ($1, $2, $3)"

	// Execute query and get result
	result, err := u.db.Exec(ctx, query, user.Username, user.Email, user.Password)
	if err != nil {
		return 0, err
	}

	// Get last inserted ID
	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return id, nil
}
