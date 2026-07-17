package middleware

import (
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"
)

func TestContentTypeValidatorAcceptsParameters(t *testing.T) {
	app := fiber.New()
	validate := ContentTypeValidator([]string{"application/json"})
	app.Post("/", func(ctx *fiber.Ctx) error {
		resp := validate(ctx, nil)
		return ctx.SendStatus(resp.Code)
	})

	req := httptest.NewRequest(fiber.MethodPost, "/", nil)
	req.Header.Set(fiber.HeaderContentType, "application/json; charset=utf-8")
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusOK, resp.StatusCode)
}

func TestRateLimitResetsAfterWindow(t *testing.T) {
	app := fiber.New()
	limit := RateLimit(RateLimitConfig{MaxRequests: 1, WindowSize: 1})
	app.Get("/", func(ctx *fiber.Ctx) error {
		resp := limit(ctx, nil)
		return ctx.SendStatus(resp.Code)
	})

	request := func() int {
		resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/", nil))
		require.NoError(t, err)
		return resp.StatusCode
	}

	require.Equal(t, fiber.StatusOK, request())
	require.Equal(t, fiber.StatusTooManyRequests, request())
	time.Sleep(1100 * time.Millisecond)
	require.Equal(t, fiber.StatusOK, request())
}
