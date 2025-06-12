package router

import (
	"github.com/gofiber/fiber/v2"
	"github.com/hanifkf12/hanif_skeleton/internal/appctx"
	"github.com/hanifkf12/hanif_skeleton/internal/bootstrap"
	"github.com/hanifkf12/hanif_skeleton/internal/handler"
	"github.com/hanifkf12/hanif_skeleton/internal/repository/home"
	userRepo "github.com/hanifkf12/hanif_skeleton/internal/repository/user"
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
	userRepository := userRepo.NewUserRepository(db)

	healthUseCase := usecase.NewHealth(homeRepo)
	rtr.fiber.Get("/health", rtr.handle(
		handler.HttpRequest,
		healthUseCase,
	))

	userUseCase := usecase.NewUser(userRepository)
	rtr.fiber.Get("/users", rtr.handle(
		handler.HttpRequest,
		userUseCase,
	))

	createUserUseCase := usecase.NewCreateUser(userRepository)
	rtr.fiber.Post("/users", rtr.handle(
		handler.HttpRequest,
		createUserUseCase,
	))

	updateUserUseCase := usecase.NewUpdateUser(userRepository)
	rtr.fiber.Put("/users/:id", rtr.handle(
		handler.HttpRequest,
		updateUserUseCase,
	))

	deleteUserUseCase := usecase.NewDeleteUser(userRepository)
	rtr.fiber.Delete("/users/:id", rtr.handle(
		handler.HttpRequest,
		deleteUserUseCase,
	))
}

func NewRouter(cfg *config.Config, fiber fiber.Router) Router {
	return &router{
		cfg:   cfg,
		fiber: fiber,
	}
}
