package app

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/hanifkf12/hanif_skeleton/internal/router"
	"github.com/hanifkf12/hanif_skeleton/pkg/config"
)

type App struct {
	*fiber.App
	Cfg *config.Config
}

func InitializeApp(cfg *config.Config) *App {
	f := fiber.New(fiber.Config{})

	rtr := router.NewRouter(cfg, f)

	rtr.Route()

	return &App{
		App: f,
		Cfg: cfg,
	}
}

func (app *App) Run() error {
	return app.Listen(fmt.Sprintf("%s:%s", "localhost", app.Cfg.App.Port))
}
