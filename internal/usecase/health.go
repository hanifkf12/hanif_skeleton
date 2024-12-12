package usecase

import (
	"github.com/gofiber/fiber/v2"
	"github.com/hanifkf12/hanif_skeleton/internal/appctx"
	"github.com/hanifkf12/hanif_skeleton/internal/repository"
	"github.com/hanifkf12/hanif_skeleton/internal/usecase/contract"
	"github.com/hanifkf12/hanif_skeleton/pkg/logger"
)

type health struct {
	homeRepo repository.HomeRepository
}

func (h *health) Serve(data appctx.Data) appctx.Response {
	var (
		lf = logger.NewFields(logger.EventName("Testt"))
	)
	lf.Append(logger.Any("data", "datalllll"))
	list, err := h.homeRepo.GetAdmin(data.FiberCtx.UserContext(), "aaaaa")
	if err != nil {
		return *appctx.NewResponse().WithCode(fiber.StatusInternalServerError).WithErrors(err.Error())
	}
	logger.Info("OKKKK", lf...)
	return *appctx.NewResponse().WithCode(fiber.StatusOK).WithData("Pong").WithData(list)
}

func NewHealth(homeRepo repository.HomeRepository) contract.UseCase {
	return &health{
		homeRepo: homeRepo,
	}
}
