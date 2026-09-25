package middleware

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"strconv"
	"time"

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
// Signatures older than five minutes (or too far in the future) are rejected
// to reduce replay risk. Returns 200 if valid, 401 if invalid.
func HMACAuth(secretKey string) Middleware {
	return func(ctx *fiber.Ctx, cfg *config.Config) appctx.Response {
		lf := logger.NewFields("Middleware.HMACAuth")

		// Get signature from header
		signature := ctx.Get("X-Signature")
		if signature == "" {
			lf.Append(logger.Any("error", "missing X-Signature header"))
			logger.Error("HMAC validation failed", lf)
			return *appctx.NewResponse().
				WithCode(appctx.StatusUnauthorized).
				WithErrors("Missing signature")
		}

		// Get timestamp
		timestamp := ctx.Get("X-Timestamp")
		if timestamp == "" {
			lf.Append(logger.Any("error", "missing X-Timestamp header"))
			logger.Error("HMAC validation failed", lf)
			return *appctx.NewResponse().
				WithCode(appctx.StatusUnauthorized).
				WithErrors("Missing timestamp")
		}
		unixTimestamp, err := strconv.ParseInt(timestamp, 10, 64)
		if err != nil || absDuration(time.Since(time.Unix(unixTimestamp, 0))) > 5*time.Minute {
			lf.Append(logger.Any("error", "invalid or stale timestamp"))
			logger.Error("HMAC validation failed", lf)
			return *appctx.NewResponse().
				WithCode(appctx.StatusUnauthorized).
				WithErrors("Invalid timestamp")
		}

		// Get request body
		body := ctx.Body()

		// Create message to sign: method + path + timestamp + body
		message := ctx.Method() + ctx.Path() + timestamp + string(body)

		// Calculate HMAC
		h := hmac.New(sha256.New, []byte(secretKey))
		h.Write([]byte(message))
		expectedSignature := h.Sum(nil)
		providedSignature, err := hex.DecodeString(signature)

		// Compare signatures
		if err != nil || !hmac.Equal(providedSignature, expectedSignature) {
			lf.Append(logger.Any("error", "invalid signature"))
			logger.Error("HMAC validation failed", lf)
			return *appctx.NewResponse().
				WithCode(appctx.StatusUnauthorized).
				WithErrors("Invalid signature")
		}

		lf.Append(logger.Any("method", ctx.Method()))
		lf.Append(logger.Any("path", ctx.Path()))
		logger.Info("HMAC validation successful", lf)

		return *appctx.NewResponse().WithCode(appctx.StatusOK)
	}
}

func absDuration(duration time.Duration) time.Duration {
	if duration < 0 {
		return -duration
	}
	return duration
}
