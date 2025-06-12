package user

import (
	"context"
	"github.com/hanifkf12/hanif_skeleton/internal/entity"
	"github.com/hanifkf12/hanif_skeleton/internal/repository"
	"github.com/hanifkf12/hanif_skeleton/pkg/databasex"
	"github.com/hanifkf12/hanif_skeleton/pkg/telemetry"
)

type userRepository struct {
	db databasex.Database
}

func (r *userRepository) GetUsers(ctx context.Context) ([]entity.User, error) {
	ctx, span := telemetry.StartSpan(ctx, "userRepository.GetUsers")
	defer span.End()

	var users []entity.User
	query := `SELECT id, name, email, username, created_at, updated_at FROM users`

	err := r.db.Select(ctx, &users, query)
	if err != nil {
		return nil, err
	}

	return users, nil
}

func NewUserRepository(db databasex.Database) repository.UserRepository {
	return &userRepository{
		db: db,
	}
}
