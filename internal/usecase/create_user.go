package usecase

import (
	"github.com/hanifkf12/hanif_skeleton/internal/appctx"
	"github.com/hanifkf12/hanif_skeleton/internal/entity"
	"github.com/hanifkf12/hanif_skeleton/internal/repository"
	"github.com/hanifkf12/hanif_skeleton/internal/usecase/contract"
)

type createUser struct {
	userRepo repository.UserRepository
}

func NewCreateUser(userRepo repository.UserRepository) contract.UseCase {
	return &createUser{userRepo: userRepo}
}

func (u *createUser) Serve(data appctx.Data) appctx.Response {
	// Parse request body
	req := new(entity.CreateUserRequest)
	if err := data.FiberCtx.BodyParser(req); err != nil {
		return *appctx.NewResponse().WithErrors(err.Error())
	}

	// Create user in database
	userID, err := u.userRepo.CreateUser(data.FiberCtx.Context(), *req)
	if err != nil {
		return *appctx.NewResponse().WithErrors(err.Error())
	}

	// Prepare response
	resp := entity.CreateUserResponse{
		ID:       userID,
		Username: req.Username,
		Email:    req.Email,
	}

	return *appctx.NewResponse().WithData(resp)
}
