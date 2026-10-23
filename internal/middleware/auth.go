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
				WithCode(appctx.StatusUnauthorized).
				WithErrors("Missing authorization header")
		}

		// Check Bearer prefix
		if !strings.HasPrefix(authHeader, "Bearer ") {
			lf.Append(logger.Any("error", "invalid Authorization format"))
			logger.Error("JWT auth validation failed", lf)
			return *appctx.NewResponse().
				WithCode(appctx.StatusUnauthorized).
				WithErrors("Invalid authorization format")
		}

		// Extract token
		token := strings.TrimPrefix(authHeader, "Bearer ")
		if token == "" {
			lf.Append(logger.Any("error", "empty token"))
			logger.Error("JWT auth validation failed", lf)
			return *appctx.NewResponse().
				WithCode(appctx.StatusUnauthorized).
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
				WithCode(appctx.StatusUnauthorized).
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

		return *appctx.NewResponse().WithCode(appctx.StatusOK)
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
				WithCode(appctx.StatusForbidden).
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
				WithCode(appctx.StatusForbidden).
				WithErrors("Insufficient permissions")
		}

		lf.Append(logger.Any("role", role))
		logger.Info("Role validation successful", lf)

		return *appctx.NewResponse().WithCode(appctx.StatusOK)
	}
}
