package user

import (
	"context"

	"github.com/hanifkf12/hanif_skeleton/internal/entity"
	"github.com/hanifkf12/hanif_skeleton/internal/repository"
)

func (u *userRepository) CreateUser(ctx context.Context, input entity.UserInput) (int64, error) {
	// Kolom `name` diisi dari Username — perilaku lama yang dipertahankan.
	const query = `
		INSERT INTO users (name, username, email, password_hash, role)
		VALUES ($1, $2, $3, $4, 'user')
		RETURNING id
	`

	var id int64
	err := u.db.QueryRowX(ctx, query, input.Username, input.Username, input.Email, input.PasswordHash).Scan(&id)
	if err != nil {
		return 0, repository.MapSQLError(err, "User not found")
	}

	return id, nil
}
