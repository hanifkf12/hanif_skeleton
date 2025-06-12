package user

import (
	"context"
	"fmt"
	"github.com/hanifkf12/hanif_skeleton/internal/entity"
	"strings"
)

func (u *userRepository) UpdateUser(ctx context.Context, user entity.UpdateUserRequest) error {
	// Build the dynamic update query based on which fields are provided
	setClauses := []string{}
	args := []interface{}{}

	if user.Username != "" {
		setClauses = append(setClauses, "username = ?")
		args = append(args, user.Username)
	}

	if user.Email != "" {
		setClauses = append(setClauses, "email = ?")
		args = append(args, user.Email)
	}

	if user.Password != "" {
		setClauses = append(setClauses, "password = ?")
		args = append(args, user.Password)
	}

	// If no fields to update
	if len(setClauses) == 0 {
		return fmt.Errorf("no fields to update")
	}

	// Build the final query
	query := fmt.Sprintf("UPDATE users SET %s WHERE id = ?", strings.Join(setClauses, ", "))

	// Add the ID to args
	args = append(args, user.ID)

	// Execute the query
	_, err := u.db.Exec(ctx, query, args...)
	return err
}
