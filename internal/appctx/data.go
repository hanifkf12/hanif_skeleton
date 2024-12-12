package appctx

import (
	"github.com/gofiber/fiber/v2"
	"github.com/hanifkf12/hanif_skeleton/pkg/config"
)

type Data struct {
	FiberCtx *fiber.Ctx
	Cfg      *config.Config
}
