package usecase

import (
	"context"

	"github.com/hanifkf12/hanif_skeleton/internal/entity"
)

// fakeUserRepository adalah UserRepository in-memory untuk unit test. Setiap
// method mencatat panggilannya, sehingga test bisa membuktikan bahwa sebuah
// request yang ditolak memang tidak pernah menyentuh penyimpanan — bukan
// sekadar membuktikan kode statusnya benar.
type fakeUserRepository struct {
	users     []entity.User
	usersErr  error
	user      *entity.User // hasil GetUserByUsername
	userErr   error
	byID      *entity.User // hasil GetUserByID
	byIDErr   error
	createID  int64
	createErr error
	updateErr error
	deleteErr error

	createdUser  *entity.UserInput
	updatedID    int64
	updatedWith  *entity.UserUpdate
	deletedID    int64
	updateCalls  int
	deleteCalls  int
	getByIDCalls int
}

func (r *fakeUserRepository) GetUsers(context.Context) ([]entity.User, error) {
	return r.users, r.usersErr
}

func (r *fakeUserRepository) GetUserByID(_ context.Context, id int64) (*entity.User, error) {
	r.getByIDCalls++
	if r.byIDErr != nil {
		return nil, r.byIDErr
	}
	if r.byID == nil {
		return &entity.User{}, nil
	}
	return r.byID, nil
}

func (r *fakeUserRepository) GetUserByUsername(context.Context, string) (*entity.User, error) {
	return r.user, r.userErr
}

func (r *fakeUserRepository) CreateUser(_ context.Context, input entity.UserInput) (int64, error) {
	r.createdUser = &input
	return r.createID, r.createErr
}

func (r *fakeUserRepository) UpdateUser(_ context.Context, id int64, update entity.UserUpdate) error {
	r.updateCalls++
	r.updatedID = id
	r.updatedWith = &update
	return r.updateErr
}

func (r *fakeUserRepository) DeleteUser(_ context.Context, id int64) error {
	r.deleteCalls++
	r.deletedID = id
	return r.deleteErr
}

var _ interface {
	GetUsers(context.Context) ([]entity.User, error)
	GetUserByID(context.Context, int64) (*entity.User, error)
	GetUserByUsername(context.Context, string) (*entity.User, error)
	CreateUser(context.Context, entity.UserInput) (int64, error)
	UpdateUser(context.Context, int64, entity.UserUpdate) error
	DeleteUser(context.Context, int64) error
} = (*fakeUserRepository)(nil)
