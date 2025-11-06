package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/hanifkf12/hanif_skeleton/internal/appctx"
	"github.com/hanifkf12/hanif_skeleton/pkg/config"
	"github.com/hanifkf12/hanif_skeleton/pkg/jwt"
	"github.com/hanifkf12/hanif_skeleton/pkg/logger"
)

// JWTAuth validates JWT token from Authorization header
// Returns 200 if valid, 401 if invalid
func JWTAuth(jwtInstance jwt.JWT) Middleware {
	return func(ctx *fiber.Ctx, cfg *config.Config) appctx.Response {
		lf := logger.NewFields("Middleware.JWTAuth")

		// Get Authorization header
		authHeader := ctx.Get("Authorization")
		if authHeader == "" {
			lf.Append(logger.Any("error", "missing Authorization header"))
			logger.Error("JWT auth validation failed", lf)
			return *appctx.NewResponse().
				WithCode(fiber.StatusUnauthorized).
				WithErrors("Missing authorization header")
		}

		// Check Bearer prefix
		if !strings.HasPrefix(authHeader, "Bearer ") {
			lf.Append(logger.Any("error", "invalid Authorization format"))
			logger.Error("JWT auth validation failed", lf)
			return *appctx.NewResponse().
				WithCode(fiber.StatusUnauthorized).
				WithErrors("Invalid authorization format")
		}

		// Extract token
		token := strings.TrimPrefix(authHeader, "Bearer ")
		if token == "" {
			lf.Append(logger.Any("error", "empty token"))
			logger.Error("JWT auth validation failed", lf)
			return *appctx.NewResponse().
				WithCode(fiber.StatusUnauthorized).
				WithErrors("Empty token")
		}

		// Parse and validate JWT token
		claims, err := jwtInstance.Parse(token)
		if err != nil {
			lf.Append(logger.Any("error", err.Error()))
			logger.Error("JWT auth validation failed", lf)

			// Determine error message based on error type
			errorMsg := "Invalid token"
			if err == jwt.ErrTokenExpired {
				errorMsg = "Token expired"
			}

			return *appctx.NewResponse().
				WithCode(fiber.StatusUnauthorized).
				WithErrors(errorMsg)
		}

		// Store claims in context for later use in handlers
		ctx.Locals("user_id", claims.UserID)
		ctx.Locals("username", claims.Username)
		ctx.Locals("email", claims.Email)
		ctx.Locals("role", claims.Role)
		ctx.Locals("claims", claims)

		lf.Append(logger.Any("user_id", claims.UserID))
		lf.Append(logger.Any("username", claims.Username))
		lf.Append(logger.Any("role", claims.Role))
		lf.Append(logger.Any("path", ctx.Path()))
		logger.Info("JWT auth validation successful", lf)

		return *appctx.NewResponse().WithCode(fiber.StatusOK)
	}
}

// RequireRole validates user role from JWT claims
// Must be used after JWTAuth middleware
func RequireRole(allowedRoles []string) Middleware {
	return func(ctx *fiber.Ctx, cfg *config.Config) appctx.Response {
		lf := logger.NewFields("Middleware.RequireRole")

		// Get role from context (set by JWTAuth middleware)
		role, ok := ctx.Locals("role").(string)
		if !ok {
			lf.Append(logger.Any("error", "role not found in context"))
			logger.Error("Role validation failed", lf)
			return *appctx.NewResponse().
				WithCode(fiber.StatusForbidden).
				WithErrors("Role not found")
		}

		// Check if role is allowed
		roleAllowed := false
		for _, allowedRole := range allowedRoles {
			if role == allowedRole {
				roleAllowed = true
				break
			}
		}

		if !roleAllowed {
			lf.Append(logger.Any("user_role", role))
			lf.Append(logger.Any("allowed_roles", allowedRoles))
			logger.Error("Role validation failed - insufficient permissions", lf)
			return *appctx.NewResponse().
				WithCode(fiber.StatusForbidden).
				WithErrors("Insufficient permissions")
		}

		lf.Append(logger.Any("role", role))
		logger.Info("Role validation successful", lf)

		return *appctx.NewResponse().WithCode(fiber.StatusOK)
	}
}

// BearerAuth validates Bearer token from Authorization header (Simple version without JWT)
// Returns 200 if valid, 401 if invalid
// Note: Use JWTAuth for JWT-based authentication
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
