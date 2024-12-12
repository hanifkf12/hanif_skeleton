package router

import (
	"github.com/gofiber/fiber/v2"
	"github.com/hanifkf12/hanif_skeleton/internal/appctx"
	"github.com/hanifkf12/hanif_skeleton/internal/bootstrap"
	"github.com/hanifkf12/hanif_skeleton/internal/handler"
	"github.com/hanifkf12/hanif_skeleton/internal/repository/home"
	"github.com/hanifkf12/hanif_skeleton/internal/usecase"
	"github.com/hanifkf12/hanif_skeleton/internal/usecase/contract"
	"github.com/hanifkf12/hanif_skeleton/pkg/config"
)

type router struct {
	cfg   *config.Config
	fiber fiber.Router
}

func (rtr *router) handle(hfn httpHandlerFunc, svc contract.UseCase) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		resp := hfn(ctx, svc, rtr.cfg)
		return rtr.response(ctx, resp)
	}
}

func (rtr *router) response(ctx *fiber.Ctx, resp appctx.Response) error {
	ctx.Set("Content-Type", "application/json; charset=utf-8")
	return ctx.Status(200).Send(resp.Byte())
}

func (rtr *router) Route() {
	db := bootstrap.RegistryDatabase(rtr.cfg)
	homeRepo := home.NewHomeRepository(db)

	healthUseCase := usecase.NewHealth(homeRepo)
	rtr.fiber.Get("/health", rtr.handle(
		handler.HttpRequest,
		healthUseCase,
	))

}

func NewRouter(cfg *config.Config, fiber fiber.Router) Router {
	return &router{
		cfg:   cfg,
		fiber: fiber,
	}
}
