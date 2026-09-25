package usecase

import (
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

type createUser struct {
	userRepo  repository.UserRepository
	hasher    *crypto.BcryptHasher
	validator *validator.Validate
}

func NewCreateUser(userRepo repository.UserRepository, hasher *crypto.BcryptHasher) contract.UseCase {
	return &createUser{userRepo: userRepo, hasher: hasher, validator: validator.New()}
}

func (u *createUser) Serve(data appctx.Data) appctx.Response {
	ctx := data.Request.Context()
	ctx, span := telemetry.StartSpan(ctx, "createUser.Serve")
	defer span.End()

	lf := logger.NewFields("CreateUser").WithTrace(ctx)

	// Parse request body
	req := new(entity.CreateUserRequest)
	if err := data.Request.Body(req); err != nil {
		telemetry.SpanError(ctx, err)
		lf.Append(logger.Any("error", err.Error()))
		logger.Error("Invalid create user request", lf)
		return *appctx.ResponseFromError(apperror.Invalid("Invalid request body"))
	}

	if err := u.validator.Struct(req); err != nil {
		telemetry.SpanError(ctx, err)
		return *appctx.ResponseFromError(apperror.Wrap(err, apperror.KindInvalid, "Invalid request body"))
	}

	passwordHash, err := u.hasher.HashPassword(req.Password)
	if err != nil {
		telemetry.SpanError(ctx, err)
		logger.Error("Failed to hash password", lf)
		return *appctx.ResponseFromError(err)
	}

	// Input domain dibangun setelah hashing. CreateUserRequest tidak pernah
	// dipakai sebagai input repository, sehingga password plaintext tidak punya
	// jalur menuju persistence.
	input, err := entity.NewUserInput(req.Username, req.Email, passwordHash)
	if err != nil {
		telemetry.SpanError(ctx, err)
		return *appctx.ResponseFromError(apperror.Wrap(err, apperror.KindInvalid, "Invalid user data"))
	}

	userID, err := u.userRepo.CreateUser(ctx, input)
	if err != nil {
		telemetry.SpanError(ctx, err)
		lf.Append(logger.Any("error", err.Error()))
		logger.Error("Failed to create user", lf)
		return *appctx.ResponseFromError(err)
	}

	resp := entity.CreateUserResponse{
		ID:       userID,
		Username: input.Username,
		Email:    input.Email,
	}
	logger.Info("User created successfully", lf)

	return *appctx.NewResponse().WithCode(appctx.StatusCreated).WithData(resp)
}
