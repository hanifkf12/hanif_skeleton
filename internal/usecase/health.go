package usecase

import (
	"github.com/hanifkf12/hanif_skeleton/internal/appctx"
	"github.com/hanifkf12/hanif_skeleton/internal/usecase/contract"
)

type health struct{}

func (h *health) Serve(data appctx.Data) appctx.Response {
	return *appctx.NewResponse().
		WithCode(appctx.StatusOK).
		WithStatus(true).
		WithData(map[string]string{"status": "healthy"})
}

func NewHealth() contract.UseCase {
	return &health{}
}
