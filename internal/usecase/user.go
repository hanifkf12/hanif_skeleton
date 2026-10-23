package usecase

import (
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
	ctx := data.Request.Context()
	ctx, span := telemetry.StartSpan(ctx, "user.Serve")
	defer span.End()
	var (
		lf = logger.NewFields("GetUsers").WithTrace(ctx)
	)
	users, err := u.userRepo.GetUsers(ctx)
	if err != nil {
		telemetry.SpanError(ctx, err)
		lf.Append(logger.Any("error", err.Error()))
		logger.Error("Failed to get users", lf)
		return *appctx.ResponseFromError(err)
	}

	logger.Info("Successfully retrieved users", lf)
	return *appctx.NewResponse().WithCode(appctx.StatusOK).WithData(users)
}

func NewUser(userRepo repository.UserRepository) contract.UseCase {
	return &user{
		userRepo: userRepo,
	}
}
