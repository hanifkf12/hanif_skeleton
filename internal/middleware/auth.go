package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/hanifkf12/hanif_skeleton/internal/appctx"
	"github.com/hanifkf12/hanif_skeleton/pkg/config"
	"github.com/hanifkf12/hanif_skeleton/pkg/logger"
)

// BearerAuth validates Bearer token from Authorization header
// Returns 200 if valid, 401 if invalid
func BearerAuth(validTokens []string) Middleware {
	return func(ctx *fiber.Ctx, cfg *config.Config) appctx.Response {
		lf := logger.NewFields("Middleware.BearerAuth")

		// Get Authorization header
		authHeader := ctx.Get("Authorization")
		if authHeader == "" {
			lf.Append(logger.Any("error", "missing Authorization header"))
			logger.Error("Auth validation failed", lf)
			return *appctx.NewResponse().
				WithCode(fiber.StatusUnauthorized).
				WithErrors("Missing authorization header")
		}

		// Check Bearer prefix
		if !strings.HasPrefix(authHeader, "Bearer ") {
			lf.Append(logger.Any("error", "invalid Authorization format"))
			logger.Error("Auth validation failed", lf)
			return *appctx.NewResponse().
				WithCode(fiber.StatusUnauthorized).
				WithErrors("Invalid authorization format")
		}

		// Extract token
		token := strings.TrimPrefix(authHeader, "Bearer ")
		if token == "" {
			lf.Append(logger.Any("error", "empty token"))
			logger.Error("Auth validation failed", lf)
			return *appctx.NewResponse().
				WithCode(fiber.StatusUnauthorized).
				WithErrors("Empty token")
		}

		// Validate token
		valid := false
		for _, validToken := range validTokens {
			if token == validToken {
				valid = true
				break
			}
		}

		if !valid {
			lf.Append(logger.Any("error", "invalid token"))
			logger.Error("Auth validation failed", lf)
			return *appctx.NewResponse().
				WithCode(fiber.StatusUnauthorized).
				WithErrors("Invalid token")
		}

		lf.Append(logger.Any("path", ctx.Path()))
		logger.Info("Auth validation successful", lf)

		// Store token in context for later use
		ctx.Locals("token", token)

		return *appctx.NewResponse().WithCode(fiber.StatusOK)
	}
}

// APIKeyAuth validates API key from header
// Returns 200 if valid, 401 if invalid
func APIKeyAuth(headerName string, validKeys []string) Middleware {
	return func(ctx *fiber.Ctx, cfg *config.Config) appctx.Response {
		lf := logger.NewFields("Middleware.APIKeyAuth")

		// Get API key from header
		apiKey := ctx.Get(headerName)
		if apiKey == "" {
			lf.Append(logger.Any("error", "missing API key header"))
			lf.Append(logger.Any("header", headerName))
			logger.Error("API key validation failed", lf)
			return *appctx.NewResponse().
				WithCode(fiber.StatusUnauthorized).
				WithErrors("Missing API key")
		}

		// Validate API key
		valid := false
		for _, validKey := range validKeys {
			if apiKey == validKey {
				valid = true
				break
			}
		}

		if !valid {
			lf.Append(logger.Any("error", "invalid API key"))
			logger.Error("API key validation failed", lf)
			return *appctx.NewResponse().
				WithCode(fiber.StatusUnauthorized).
				WithErrors("Invalid API key")
		}

		lf.Append(logger.Any("path", ctx.Path()))
		logger.Info("API key validation successful", lf)

		// Store API key in context for later use
		ctx.Locals("api_key", apiKey)

		return *appctx.NewResponse().WithCode(fiber.StatusOK)
	}
}
