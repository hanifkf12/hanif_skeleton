package usecase

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/hanifkf12/hanif_skeleton/internal/appctx"
	"github.com/hanifkf12/hanif_skeleton/internal/entity"
	"github.com/hanifkf12/hanif_skeleton/internal/repository"
	"github.com/hanifkf12/hanif_skeleton/internal/usecase/contract"
	"github.com/hanifkf12/hanif_skeleton/pkg/logger"
	"github.com/hanifkf12/hanif_skeleton/pkg/telemetry"
)

type getUserByID struct {
	userRepo repository.UserRepository
}

func NewGetUserByID(userRepo repository.UserRepository) contract.UseCase {
	return &getUserByID{userRepo: userRepo}
}

func (u *getUserByID) Serve(data appctx.Data) appctx.Response {
	ctx := data.FiberCtx.UserContext()
	ctx, span := telemetry.StartSpan(ctx, "getUserByID.Serve")
	defer span.End()

	lf := logger.NewFields("GetUserByID").WithTrace(ctx)

	// Parse :id from URL param
	idStr := data.FiberCtx.Params("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		lf.Append(logger.Any("id", idStr))
		logger.Error("Invalid user ID", lf)
		return *appctx.NewResponse().WithCode(fiber.StatusBadRequest).WithErrors("invalid user id")
	}

	lf.Append(logger.Any("user_id", id))

	user, err := u.userRepo.GetUserByID(ctx, id)
	if err != nil {
		telemetry.SpanError(ctx, err)
		lf.Append(logger.Any("error", err.Error()))
		logger.Error("Failed to get user", lf)
		return *appctx.NewResponse().WithCode(fiber.StatusNotFound).WithErrors("user not found")
	}

	resp := entity.GetUserByIDResponse{
		ID:        user.Id,
		Name:      user.Name,
		Email:     user.Email,
		Username:  user.Username,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}

	logger.Info("Successfully retrieved user", lf)
	return *appctx.NewResponse().WithCode(fiber.StatusOK).WithData(resp)
}
