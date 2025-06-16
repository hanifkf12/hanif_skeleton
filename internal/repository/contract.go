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
	CreateUser(ctx context.Context, user entity.CreateUserRequest) (int64, error)
	UpdateUser(ctx context.Context, user entity.UpdateUserRequest) error
	DeleteUser(ctx context.Context, id int64) error
}

type CampaignRepository interface {
	Create(ctx context.Context, campaign *entity.Campaign) error
	Update(ctx context.Context, campaign *entity.Campaign) error
	Delete(ctx context.Context, id string) error
	GetByID(ctx context.Context, id string) (*entity.Campaign, error)
	GetAll(ctx context.Context) ([]entity.Campaign, error)
}
