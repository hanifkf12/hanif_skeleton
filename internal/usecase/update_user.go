package usecase

import (
	"github.com/gofiber/fiber/v2"
	"github.com/hanifkf12/hanif_skeleton/internal/appctx"
	"github.com/hanifkf12/hanif_skeleton/internal/entity"
	"github.com/hanifkf12/hanif_skeleton/internal/repository"
	"github.com/hanifkf12/hanif_skeleton/internal/usecase/contract"
	"github.com/hanifkf12/hanif_skeleton/pkg/logger"
	"strconv"
)

type updateUser struct {
	userRepo repository.UserRepository
}

func NewUpdateUser(userRepo repository.UserRepository) contract.UseCase {
	return &updateUser{userRepo: userRepo}
}

func (u *updateUser) Serve(data appctx.Data) appctx.Response {
	var (
		lf = logger.NewFields("UpdateUser")
	)

	// Parse user ID from path parameter
	userID := data.FiberCtx.Params("id")
	if userID == "" {
		return *appctx.NewResponse().WithCode(fiber.StatusBadRequest).WithErrors("User ID is required")
	}

	// Convert user ID to int64
	id, err := strconv.ParseInt(userID, 10, 64)
	if err != nil {
		return *appctx.NewResponse().WithCode(fiber.StatusBadRequest).WithErrors("Invalid user ID format")
	}

	// Parse request body
	req := new(entity.UpdateUserRequest)
	if err := data.FiberCtx.BodyParser(req); err != nil {
		return *appctx.NewResponse().WithCode(fiber.StatusBadRequest).WithErrors(err.Error())
	}

	// Set the ID from the path parameter
	req.ID = id

	// Update user in database
	err = u.userRepo.UpdateUser(data.FiberCtx.Context(), *req)
	if err != nil {
		lf.Append(logger.Any("error", err.Error()))
		logger.Error("Failed to update user", lf)
		return *appctx.NewResponse().WithCode(fiber.StatusInternalServerError).WithErrors(err.Error())
	}

	// Prepare response
	resp := entity.UpdateUserResponse{
		ID: id,
	}

	// Include updated fields in response
	if req.Username != "" {
		resp.Username = req.Username
	}

	if req.Email != "" {
		resp.Email = req.Email
	}

	logger.Info("User updated successfully", lf)
	return *appctx.NewResponse().WithCode(fiber.StatusOK).WithData(resp)
}
