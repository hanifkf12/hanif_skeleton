package repository

import (
	"context"
	"github.com/hanifkf12/hanif_skeleton/internal/entity"
)

type HomeRepository interface {
	GetAdmin(ctx context.Context, data any) ([]entity.Admin, error)
}

type UserRepository interface {
	GetUsers(ctx context.Context) ([]entity.User, error)
}
