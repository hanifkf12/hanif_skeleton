package usecase

import (
	"strconv"

	"github.com/go-playground/validator/v10"
	"github.com/hanifkf12/hanif_skeleton/internal/appctx"
	"github.com/hanifkf12/hanif_skeleton/internal/entity"
	"github.com/hanifkf12/hanif_skeleton/internal/repository"
	"github.com/hanifkf12/hanif_skeleton/internal/usecase/contract"
	"github.com/hanifkf12/hanif_skeleton/pkg/apperror"
	"github.com/hanifkf12/hanif_skeleton/pkg/crypto"
	"github.com/hanifkf12/hanif_skeleton/pkg/logger"
	"github.com/hanifkf12/hanif_skeleton/pkg/telemetry"
)

type updateUser struct {
	userRepo  repository.UserRepository
	hasher    *crypto.BcryptHasher
	validator *validator.Validate
}

func NewUpdateUser(userRepo repository.UserRepository, hasher *crypto.BcryptHasher) contract.UseCase {
	return &updateUser{userRepo: userRepo, hasher: hasher, validator: validator.New()}
}

func (u *updateUser) Serve(data appctx.Data) appctx.Response {
	ctx := data.Request.Context()
	ctx, span := telemetry.StartSpan(ctx, "updateUser.Serve")
	defer span.End()

	lf := logger.NewFields("UpdateUser").WithTrace(ctx)

	// Parse user ID from path parameter
	userID := data.Request.Param("id")
	if userID == "" {
		return *appctx.ResponseFromError(apperror.Invalid("User ID is required"))
	}

	id, err := strconv.ParseInt(userID, 10, 64)
	if err != nil {
		return *appctx.ResponseFromError(apperror.Invalid("Invalid user ID format"))
	}

	// Otorisasi dijalankan SEBELUM body di-parse dan sebelum apa pun ditulis,
	// supaya request yang ditolak tidak pernah menyentuh database.
	actor, ok := data.Request.Principal()
	if !ok {
		return *appctx.ResponseFromError(apperror.Unauthorized("Authentication required"))
	}
	if !actor.CanModifyUser(id) {
		lf.Append(logger.Any("actor_user_id", actor.UserID))
		lf.Append(logger.Any("target_user_id", id))
		logger.Error("User attempted to modify another user", lf)
		return *appctx.ResponseFromError(apperror.Forbidden("Cannot modify another user"))
	}

	// Parse request body
	req := new(entity.UpdateUserRequest)
	if err := data.Request.Body(req); err != nil {
		return *appctx.ResponseFromError(apperror.Invalid("Invalid request body"))
	}
	if err := u.validator.Struct(req); err != nil {
		return *appctx.ResponseFromError(apperror.Wrap(err, apperror.KindInvalid, "Invalid request body"))
	}

	// UpdateUserRequest tetap menjadi bentuk wire (body HTTP) dan dipetakan ke
	// entity.UserUpdate di sini — repository tidak pernah melihatnya.
	update := entity.UserUpdate{}
	if req.Username != "" {
		update.Username = &req.Username
	}
	if req.Email != "" {
		update.Email = &req.Email
	}
	if req.Password != "" {
		passwordHash, err := u.hasher.HashPassword(req.Password)
		if err != nil {
			telemetry.SpanError(ctx, err)
			logger.Error("Failed to hash password", lf)
			return *appctx.ResponseFromError(err)
		}
		update.PasswordHash = &passwordHash
	}

	if update.IsEmpty() {
		return *appctx.ResponseFromError(apperror.Invalid("No fields to update"))
	}

	if err := u.userRepo.UpdateUser(ctx, id, update); err != nil {
		telemetry.SpanError(ctx, err)
		lf.Append(logger.Any("error", err.Error()))
		logger.Error("Failed to update user", lf)
		return *appctx.ResponseFromError(err)
	}

	resp := entity.UpdateUserResponse{ID: id}
	if update.Username != nil {
		resp.Username = *update.Username
	}
	if update.Email != nil {
		resp.Email = *update.Email
	}

	logger.Info("User updated successfully", lf)
	return *appctx.NewResponse().WithCode(appctx.StatusOK).WithData(resp)
}
