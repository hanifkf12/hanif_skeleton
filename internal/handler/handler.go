package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/hanifkf12/hanif_skeleton/internal/appctx"
	"github.com/hanifkf12/hanif_skeleton/internal/usecase/contract"
	"github.com/hanifkf12/hanif_skeleton/pkg/config"
)

func HttpRequest(xCtx *fiber.Ctx, svc contract.UseCase, conf *config.Config) appctx.Response {
	data := appctx.Data{
		FiberCtx: xCtx,
		Cfg:      conf,
	}

	return svc.Serve(data)
}
