package usecase

import (
	"github.com/gofiber/fiber/v2"
	"github.com/hanifkf12/hanif_skeleton/internal/appctx"
	"github.com/hanifkf12/hanif_skeleton/internal/repository"
	"github.com/hanifkf12/hanif_skeleton/internal/usecase/contract"
	"github.com/hanifkf12/hanif_skeleton/pkg/logger"
	"github.com/hanifkf12/hanif_skeleton/pkg/telemetry"
)

type user struct {
	userRepo repository.UserRepository
}

func (u *user) Serve(data appctx.Data) appctx.Response {
	ctx := data.FiberCtx.UserContext()
	ctx, span := telemetry.StartSpan(ctx, "user.Serve")
	defer span.End()
	var (
		lf = logger.NewFields("GetUsers").WithTrace(ctx)
	)

	users, err := u.userRepo.GetUsers(ctx)
	if err != nil {
		logger.Error("Failed to get users", lf)
		return *appctx.NewResponse().WithCode(fiber.StatusInternalServerError).WithErrors(err.Error())
	}

	logger.Info("Successfully retrieved users", lf)
	return *appctx.NewResponse().WithCode(fiber.StatusOK).WithData(users)
}

func NewUser(userRepo repository.UserRepository) contract.UseCase {
	return &user{
		userRepo: userRepo,
	}
}
