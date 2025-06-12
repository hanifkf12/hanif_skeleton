package user

import (
	"context"
)

func (u *userRepository) DeleteUser(ctx context.Context, id int64) error {
	// Define delete query
	query := "DELETE FROM users WHERE id = ?"

	// Execute query
	_, err := u.db.Exec(ctx, query, id)
	return err
}
