package user

import (
	"context"

	"github.com/hanifkf12/hanif_skeleton/internal/repository"
	"github.com/hanifkf12/hanif_skeleton/pkg/apperror"
)

func (u *userRepository) DeleteUser(ctx context.Context, id int64) error {
	const query = "DELETE FROM users WHERE id = $1"

	result, err := u.db.Exec(ctx, query, id)
	if err != nil {
		return repository.MapSQLError(err, "User not found")
	}

	// Tanpa pemeriksaan ini, menghapus id yang tidak ada akan tetap dijawab 200.
	affected, err := result.RowsAffected()
	if err != nil {
		return apperror.Internal(err)
	}
	if affected == 0 {
		return apperror.NotFound("User not found")
	}

	return nil
}
