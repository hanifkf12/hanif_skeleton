package usecase

import (
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/hanifkf12/hanif_skeleton/internal/appctx"
	"github.com/hanifkf12/hanif_skeleton/internal/entity"
	"github.com/hanifkf12/hanif_skeleton/internal/repository"
	"github.com/hanifkf12/hanif_skeleton/internal/usecase/contract"
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
	ctx := data.FiberCtx.UserContext()
	ctx, span := telemetry.StartSpan(ctx, "createUser.Serve")
	defer span.End()

	lf := logger.NewFields("CreateUser").WithTrace(ctx)
	// Parse request body
	req := new(entity.CreateUserRequest)
	if err := data.FiberCtx.BodyParser(req); err != nil {
		telemetry.SpanError(ctx, err)
		lf.Append(logger.Any("error", err.Error()))
		logger.Error("Invalid create user request", lf)
		return *appctx.NewResponse().WithCode(fiber.StatusBadRequest).WithErrors("Invalid request body")
	}
	req.Username = strings.TrimSpace(req.Username)
	req.Email = strings.TrimSpace(req.Email)
	if err := u.validator.Struct(req); err != nil {
		telemetry.SpanError(ctx, err)
		return *appctx.NewResponse().WithCode(fiber.StatusBadRequest).WithErrors(err.Error())
	}

	passwordHash, err := u.hasher.HashPassword(req.Password)
	if err != nil {
		telemetry.SpanError(ctx, err)
		logger.Error("Failed to hash password", lf)
		return *appctx.NewResponse().WithCode(fiber.StatusInternalServerError).WithErrors("Failed to create user")
	}
	req.Password = passwordHash

	// Create user in database
	userID, err := u.userRepo.CreateUser(ctx, *req)
	if err != nil {
		telemetry.SpanError(ctx, err)
		lf.Append(logger.Any("error", err.Error()))
		logger.Error("Failed to create user", lf)
		return *appctx.NewResponse().WithCode(fiber.StatusInternalServerError).WithErrors("Failed to create user")
	}

	// Prepare response
	resp := entity.CreateUserResponse{
		ID:       userID,
		Username: req.Username,
		Email:    req.Email,
	}
	logger.Info("User created successfully", lf)

	return *appctx.NewResponse().WithCode(fiber.StatusCreated).WithData(resp)
}
