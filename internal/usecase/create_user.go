package usecase

import (
	"github.com/gofiber/fiber/v2"
	"github.com/hanifkf12/hanif_skeleton/internal/appctx"
	"github.com/hanifkf12/hanif_skeleton/internal/entity"
	"github.com/hanifkf12/hanif_skeleton/internal/repository"
	"github.com/hanifkf12/hanif_skeleton/internal/usecase/contract"
	"github.com/hanifkf12/hanif_skeleton/pkg/logger"
	"github.com/hanifkf12/hanif_skeleton/pkg/telemetry"
)

type createUser struct {
	userRepo repository.UserRepository
}

func NewCreateUser(userRepo repository.UserRepository) contract.UseCase {
	return &createUser{userRepo: userRepo}
}

func (u *createUser) Serve(data appctx.Data) appctx.Response {
	ctx := data.FiberCtx.UserContext()
	ctx, span := telemetry.StartSpan(ctx, "createCampaign.Serve")
	defer span.End()

	lf := logger.NewFields("CreateCampaign").WithTrace(ctx)
	// Parse request body
	req := new(entity.CreateUserRequest)
	if err := data.FiberCtx.BodyParser(req); err != nil {
		telemetry.SpanError(ctx, err)
		lf.Append(logger.Any("error", err.Error()))
		lf.Append(logger.Any("request", req))
		logger.Error("Invalid create campaign request", lf)
		return *appctx.NewResponse().WithCode(fiber.StatusBadRequest).WithErrors(err.Error())
	}

	// Create user in database
	userID, err := u.userRepo.CreateUser(data.FiberCtx.Context(), *req)
	if err != nil {
		telemetry.SpanError(ctx, err)
		lf.Append(logger.Any("error", err.Error()))
		logger.Error("Failed to create user", lf)
		return *appctx.NewResponse().WithCode(fiber.StatusInternalServerError).WithErrors(err.Error())
	}

	// Prepare response
	resp := entity.CreateUserResponse{
		ID:       userID,
		Username: req.Username,
		Email:    req.Email,
	}
	logger.Info("User created successfully", lf)

	return *appctx.NewResponse().WithData(resp)
}
