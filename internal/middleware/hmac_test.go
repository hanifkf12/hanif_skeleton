package middleware

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"
)

func TestHMACAuthRejectsStaleRequest(t *testing.T) {
	const secret = "test-secret"
	timestamp := strconv.FormatInt(time.Now().Add(-10*time.Minute).Unix(), 10)
	signature := testHMACSignature(fiber.MethodPost, "/webhook", timestamp, "{}", secret)

	app := fiber.New()
	validate := HMACAuth(secret)
	app.Post("/webhook", func(ctx *fiber.Ctx) error {
		resp := validate(ctx, nil)
		return ctx.SendStatus(resp.Code)
	})

	req := httptest.NewRequest(fiber.MethodPost, "/webhook", strings.NewReader("{}"))
	req.Header.Set("X-Timestamp", timestamp)
	req.Header.Set("X-Signature", signature)
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestHMACAuthAcceptsCurrentValidRequest(t *testing.T) {
	const secret = "test-secret"
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	signature := testHMACSignature(fiber.MethodPost, "/webhook", timestamp, "{}", secret)

	app := fiber.New()
	validate := HMACAuth(secret)
	app.Post("/webhook", func(ctx *fiber.Ctx) error {
		resp := validate(ctx, nil)
		return ctx.SendStatus(resp.Code)
	})

	req := httptest.NewRequest(fiber.MethodPost, "/webhook", strings.NewReader("{}"))
	req.Header.Set("X-Timestamp", timestamp)
	req.Header.Set("X-Signature", signature)
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusOK, resp.StatusCode)
}

func testHMACSignature(method, path, timestamp, body, secret string) string {
	hash := hmac.New(sha256.New, []byte(secret))
	hash.Write([]byte(method + path + timestamp + body))
	return hex.EncodeToString(hash.Sum(nil))
}
