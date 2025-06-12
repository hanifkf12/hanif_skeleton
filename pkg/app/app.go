package app

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/monitor"
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

	// Initialize default config (Assign the middleware to /metrics)
	f.Get("/metrics", monitor.New())
	return &App{
		App: f,
		Cfg: cfg,
	}
}

func (app *App) Run() error {
	return app.Listen(fmt.Sprintf("%s:%s", "localhost", app.Cfg.App.Port))
}
