package usecase

import (
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/hanifkf12/hanif_skeleton/internal/appctx"
	"github.com/stretchr/testify/require"
)

func TestHealthDoesNotDependOnApplicationTables(t *testing.T) {
	service := NewHealth()
	app := fiber.New()
	app.Get("/health", func(ctx *fiber.Ctx) error {
		resp := service.Serve(appctx.Data{FiberCtx: ctx})
		return ctx.Status(resp.Code).JSON(resp)
	})

	resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/health", nil))
	require.NoError(t, err)
	require.Equal(t, fiber.StatusOK, resp.StatusCode)
}
