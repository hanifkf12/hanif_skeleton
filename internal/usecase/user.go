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

	lf.Append(logger.Any("user_id", 123))
	lf.Append(logger.Any("user_name", "John Doe"))
	lf.Append(logger.Any("user_email", "john.doe@example.com"))
	lf.Append(logger.Any("user_phone", "1234567890"))
	lf.Append(logger.Any("user_address", "123 Main St, Anytown, USA"))
	lf.Append(logger.Any("user_city", "Anytown"))
	lf.Append(logger.Any("user_state", "CA"))
	lf.Append(logger.Any("user_zip", "12345"))
	lf.Append(logger.Any("user_country", "USA"))

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
