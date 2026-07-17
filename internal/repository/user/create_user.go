package user

import (
	"context"
	"github.com/hanifkf12/hanif_skeleton/internal/entity"
)

func (u *userRepository) CreateUser(ctx context.Context, user entity.CreateUserRequest) (int64, error) {
	const query = `
		INSERT INTO users (name, username, email, password_hash, role)
		VALUES ($1, $2, $3, $4, 'user')
		RETURNING id
	`

	var id int64
	err := u.db.QueryRowX(ctx, query, user.Username, user.Username, user.Email, user.Password).Scan(&id)
	return id, err
}
