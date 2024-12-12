package router

import (
	"github.com/gofiber/fiber/v2"
	"github.com/hanifkf12/hanif_skeleton/internal/appctx"
	"github.com/hanifkf12/hanif_skeleton/internal/usecase/contract"
	"github.com/hanifkf12/hanif_skeleton/pkg/config"
)

type httpHandlerFunc func(xCtx *fiber.Ctx, svc contract.UseCase, cfg *config.Config) appctx.Response

type Router interface {
	Route()
}
