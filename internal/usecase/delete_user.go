package usecase

import (
	"strconv"

	"github.com/hanifkf12/hanif_skeleton/internal/appctx"
	"github.com/hanifkf12/hanif_skeleton/internal/entity"
	"github.com/hanifkf12/hanif_skeleton/internal/repository"
	"github.com/hanifkf12/hanif_skeleton/internal/usecase/contract"
	"github.com/hanifkf12/hanif_skeleton/pkg/apperror"
	"github.com/hanifkf12/hanif_skeleton/pkg/logger"
	"github.com/hanifkf12/hanif_skeleton/pkg/telemetry"
)

type deleteUser struct {
	userRepo repository.UserRepository
}

func NewDeleteUser(userRepo repository.UserRepository) contract.UseCase {
	return &deleteUser{userRepo: userRepo}
}

func (u *deleteUser) Serve(data appctx.Data) appctx.Response {
	ctx := data.Request.Context()
	ctx, span := telemetry.StartSpan(ctx, "deleteUser.Serve")
	defer span.End()

	lf := logger.NewFields("DeleteUser").WithTrace(ctx)

	// Parse user ID from path parameter
	userID := data.Request.Param("id")
	if userID == "" {
		return *appctx.ResponseFromError(apperror.Invalid("User ID is required"))
	}

	id, err := strconv.ParseInt(userID, 10, 64)
	if err != nil {
		return *appctx.ResponseFromError(apperror.Invalid("Invalid user ID format"))
	}

	if err := u.userRepo.DeleteUser(ctx, id); err != nil {
		telemetry.SpanError(ctx, err)
		lf.Append(logger.Any("error", err.Error()))
		logger.Error("Failed to delete user", lf)
		return *appctx.ResponseFromError(err)
	}

	resp := entity.DeleteUserResponse{
		Message: "User deleted successfully",
		ID:      id,
	}

	logger.Info("User deleted successfully", lf)
	return *appctx.NewResponse().WithCode(appctx.StatusOK).WithData(resp)
}
