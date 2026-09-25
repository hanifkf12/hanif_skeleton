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

type getUserByID struct {
	userRepo repository.UserRepository
}

func NewGetUserByID(userRepo repository.UserRepository) contract.UseCase {
	return &getUserByID{userRepo: userRepo}
}

func (u *getUserByID) Serve(data appctx.Data) appctx.Response {
	ctx := data.Request.Context()
	ctx, span := telemetry.StartSpan(ctx, "getUserByID.Serve")
	defer span.End()

	lf := logger.NewFields("GetUserByID").WithTrace(ctx)

	// Parse :id from URL param
	idStr := data.Request.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		lf.Append(logger.Any("id", idStr))
		logger.Error("Invalid user ID", lf)
		return *appctx.ResponseFromError(apperror.Invalid("Invalid user ID"))
	}

	// Otorisasi dijalankan sebelum query, sehingga user biasa tidak bisa
	// membedakan "user lain tidak ada" dari "user lain ada" lewat kode status.
	actor, ok := data.Request.Principal()
	if !ok {
		return *appctx.ResponseFromError(apperror.Unauthorized("Authentication required"))
	}
	if !actor.CanAccessUser(id) {
		lf.Append(logger.Any("actor_user_id", actor.UserID))
		lf.Append(logger.Any("target_user_id", id))
		logger.Error("User attempted to read another user", lf)
		return *appctx.ResponseFromError(apperror.Forbidden("Cannot access another user"))
	}

	lf.Append(logger.Any("user_id", id))

	user, err := u.userRepo.GetUserByID(ctx, id)
	if err != nil {
		telemetry.SpanError(ctx, err)
		lf.Append(logger.Any("error", err.Error()))
		logger.Error("Failed to get user", lf)
		return *appctx.ResponseFromError(err)
	}

	resp := entity.GetUserByIDResponse{
		ID:        user.ID,
		Name:      user.Name,
		Email:     user.Email,
		Username:  user.Username,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}

	logger.Info("Successfully retrieved user", lf)
	return *appctx.NewResponse().WithCode(appctx.StatusOK).WithData(resp)
}
