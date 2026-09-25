package user

import (
	"context"
	"fmt"
	"strings"

	"github.com/hanifkf12/hanif_skeleton/internal/entity"
	"github.com/hanifkf12/hanif_skeleton/internal/repository"
	"github.com/hanifkf12/hanif_skeleton/pkg/apperror"
)

func (u *userRepository) UpdateUser(ctx context.Context, id int64, update entity.UserUpdate) error {
	if update.IsEmpty() {
		return apperror.Invalid("No fields to update")
	}

	// Bangun klausa SET dinamis dari field yang benar-benar di-set. Field
	// bernilai nil berarti "jangan diubah".
	setClauses := []string{}
	args := []interface{}{}

	if update.Username != nil {
		setClauses = append(setClauses, fmt.Sprintf("username = $%d", len(args)+1))
		args = append(args, *update.Username)
	}

	if update.Email != nil {
		setClauses = append(setClauses, fmt.Sprintf("email = $%d", len(args)+1))
		args = append(args, *update.Email)
	}

	if update.PasswordHash != nil {
		setClauses = append(setClauses, fmt.Sprintf("password_hash = $%d", len(args)+1))
		args = append(args, *update.PasswordHash)
	}

	query := fmt.Sprintf("UPDATE users SET %s WHERE id = $%d", strings.Join(setClauses, ", "), len(args)+1)
	args = append(args, id)

	result, err := u.db.Exec(ctx, query, args...)
	if err != nil {
		return repository.MapSQLError(err, "User not found")
	}

	// Tanpa pemeriksaan ini, mengubah id yang tidak ada akan tetap dijawab 200.
	affected, err := result.RowsAffected()
	if err != nil {
		return apperror.Internal(err)
	}
	if affected == 0 {
		return apperror.NotFound("User not found")
	}

	return nil
}
