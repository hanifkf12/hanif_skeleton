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

type deleteUser struct {
	userRepo repository.UserRepository
}

func NewDeleteUser(userRepo repository.UserRepository) contract.UseCase {
	return &deleteUser{userRepo: userRepo}
}

func (u *deleteUser) Serve(data appctx.Data) appctx.Response {
	var (
		lf = logger.NewFields("DeleteUser")
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

	// Delete user from database
	err = u.userRepo.DeleteUser(data.FiberCtx.Context(), id)
	if err != nil {
		lf.Append(logger.Any("error", err.Error()))
		logger.Error("Failed to delete user", lf)
		return *appctx.NewResponse().WithCode(fiber.StatusInternalServerError).WithErrors(err.Error())
	}

	// Prepare response
	resp := entity.DeleteUserResponse{
		Message: "User deleted successfully",
		ID:      id,
	}

	logger.Info("User deleted successfully", lf)
	return *appctx.NewResponse().WithCode(fiber.StatusOK).WithData(resp)
}
