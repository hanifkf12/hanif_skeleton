package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/hanifkf12/hanif_skeleton/internal/appctx"
	"github.com/hanifkf12/hanif_skeleton/pkg/config"
)

// Middleware is a function that processes request before reaching the handler
// Returns appctx.Response with code 200 if middleware passes, otherwise returns error response
type Middleware func(ctx *fiber.Ctx, cfg *config.Config) appctx.Response
