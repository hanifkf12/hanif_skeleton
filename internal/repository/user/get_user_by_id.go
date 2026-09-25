package user

import (
	"context"

	"github.com/hanifkf12/hanif_skeleton/internal/entity"
	"github.com/hanifkf12/hanif_skeleton/internal/repository"
	"github.com/hanifkf12/hanif_skeleton/pkg/sqlbuilder"
	"github.com/hanifkf12/hanif_skeleton/pkg/telemetry"
)

func (u *userRepository) GetUserByID(ctx context.Context, id int64) (*entity.User, error) {
	ctx, span := telemetry.StartSpan(ctx, "userRepository.GetUserByID")
	defer span.End()

	var user entity.User

	model := sqlbuilder.NewModel(u.db, &entity.User{})
	err := model.
		Table("users").
		Select("id", "name", "email", "username", "created_at", "updated_at").
		Where("id = ?", id).
		First(ctx, &user)

	if err != nil {
		return nil, repository.MapSQLError(err, "User not found")
	}

	return &user, nil
}
