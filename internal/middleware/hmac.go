package middleware

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"

	"github.com/gofiber/fiber/v2"
	"github.com/hanifkf12/hanif_skeleton/internal/appctx"
	"github.com/hanifkf12/hanif_skeleton/pkg/config"
	"github.com/hanifkf12/hanif_skeleton/pkg/logger"
)

// HMACAuth validates HMAC signature from request headers
// Expects headers:
//   - X-Signature: HMAC signature
//   - X-Timestamp: Request timestamp
//
// Returns 200 if valid, 401 if invalid
func HMACAuth(secretKey string) Middleware {
	return func(ctx *fiber.Ctx, cfg *config.Config) appctx.Response {
		lf := logger.NewFields("Middleware.HMACAuth")

		// Get signature from header
		signature := ctx.Get("X-Signature")
		if signature == "" {
			lf.Append(logger.Any("error", "missing X-Signature header"))
			logger.Error("HMAC validation failed", lf)
			return *appctx.NewResponse().
				WithCode(fiber.StatusUnauthorized).
				WithErrors("Missing signature")
		}

		// Get timestamp
		timestamp := ctx.Get("X-Timestamp")
		if timestamp == "" {
			lf.Append(logger.Any("error", "missing X-Timestamp header"))
			logger.Error("HMAC validation failed", lf)
			return *appctx.NewResponse().
				WithCode(fiber.StatusUnauthorized).
				WithErrors("Missing timestamp")
		}

		// Get request body
		body := ctx.Body()

		// Create message to sign: method + path + timestamp + body
		message := ctx.Method() + ctx.Path() + timestamp + string(body)

		// Calculate HMAC
		h := hmac.New(sha256.New, []byte(secretKey))
		h.Write([]byte(message))
		expectedSignature := hex.EncodeToString(h.Sum(nil))

		// Compare signatures
		if !hmac.Equal([]byte(signature), []byte(expectedSignature)) {
			lf.Append(logger.Any("error", "invalid signature"))
			lf.Append(logger.Any("expected", expectedSignature))
			lf.Append(logger.Any("received", signature))
			logger.Error("HMAC validation failed", lf)
			return *appctx.NewResponse().
				WithCode(fiber.StatusUnauthorized).
				WithErrors("Invalid signature")
		}

		lf.Append(logger.Any("method", ctx.Method()))
		lf.Append(logger.Any("path", ctx.Path()))
		logger.Info("HMAC validation successful", lf)

		return *appctx.NewResponse().WithCode(fiber.StatusOK)
	}
}
