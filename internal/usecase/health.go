package usecase

import (
	"github.com/gofiber/fiber/v2"
	"github.com/hanifkf12/hanif_skeleton/internal/appctx"
	"github.com/hanifkf12/hanif_skeleton/internal/usecase/contract"
)

type health struct{}

func (h *health) Serve(data appctx.Data) appctx.Response {
	return *appctx.NewResponse().
		WithCode(fiber.StatusOK).
		WithStatus(true).
		WithData(map[string]string{"status": "healthy"})
}

func NewHealth() contract.UseCase {
	return &health{}
}
