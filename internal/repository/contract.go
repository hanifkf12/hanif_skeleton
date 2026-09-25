package repository

import (
	"context"

	"github.com/hanifkf12/hanif_skeleton/internal/entity"
)

// UserRepository adalah port penyimpanan user.
//
// Parameter-nya adalah tipe domain (entity.UserInput / entity.UserUpdate),
// bukan DTO transport. entity.UserInput.PasswordHash sudah berupa hash, jadi
// repository tidak punya cara untuk menerima password plaintext.
type UserRepository interface {
	GetUsers(ctx context.Context) ([]entity.User, error)
	GetUserByID(ctx context.Context, id int64) (*entity.User, error)
	GetUserByUsername(ctx context.Context, username string) (*entity.User, error)
	CreateUser(ctx context.Context, input entity.UserInput) (int64, error)
	UpdateUser(ctx context.Context, id int64, update entity.UserUpdate) error
	DeleteUser(ctx context.Context, id int64) error
}

// CampaignRepository sudah menerima domain object, jadi tidak perlu diubah.
type CampaignRepository interface {
	Create(ctx context.Context, campaign *entity.Campaign) error
	Update(ctx context.Context, campaign *entity.Campaign) error
	Delete(ctx context.Context, id string) error
	GetByID(ctx context.Context, id string) (*entity.Campaign, error)
	GetAll(ctx context.Context) ([]entity.Campaign, error)
}
