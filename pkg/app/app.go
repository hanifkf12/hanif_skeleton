package app

import (
	"github.com/gofiber/fiber/v2"
	"github.com/hanifkf12/hanif_skeleton/internal/router"
	"github.com/hanifkf12/hanif_skeleton/pkg/config"
	"github.com/hanifkf12/hanif_skeleton/pkg/middleware"
)

type App struct {
	*fiber.App
	Cfg *config.Config
}

func InitializeApp(cfg *config.Config) *App {
	f := fiber.New(fiber.Config{})

	// Add global trace middleware to ensure all requests are traced
	f.Use(middleware.TraceMiddleware())

	rtr := router.NewRouter(cfg, f)

	rtr.Route()

	return &App{
		App: f,
		Cfg: cfg,
	}
}

func (app *App) Run() error {
	port := app.Cfg.App.Port
	if port == "" {
		port = "9000"
	}
	return app.Listen(":" + port)
}
