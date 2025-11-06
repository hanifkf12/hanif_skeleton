package usecase

import (
	"github.com/gofiber/fiber/v2"
	"github.com/hanifkf12/hanif_skeleton/internal/appctx"
	"github.com/hanifkf12/hanif_skeleton/internal/repository"
	"github.com/hanifkf12/hanif_skeleton/internal/usecase/contract"
	"github.com/hanifkf12/hanif_skeleton/pkg/crypto"
	"github.com/hanifkf12/hanif_skeleton/pkg/jwt"
	"github.com/hanifkf12/hanif_skeleton/pkg/logger"
	"github.com/hanifkf12/hanif_skeleton/pkg/telemetry"
)

// Login usecase for user authentication
type login struct {
	userRepo repository.UserRepository
	hasher   *crypto.BcryptHasher
	jwt      jwt.JWT
}

// LoginRequest represents login request
type LoginRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

// LoginResponse represents login response
type LoginResponse struct {
	Token     string `json:"token"`
	UserID    int64  `json:"user_id"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	Role      string `json:"role"`
	ExpiresIn string `json:"expires_in"`
}

func NewLogin(userRepo repository.UserRepository, hasher *crypto.BcryptHasher, jwtInstance jwt.JWT) contract.UseCase {
	return &login{
		userRepo: userRepo,
		hasher:   hasher,
		jwt:      jwtInstance,
	}
}

func (u *login) Serve(data appctx.Data) appctx.Response {
	ctx := data.FiberCtx.UserContext()
	ctx, span := telemetry.StartSpan(ctx, "login.Serve")
	defer span.End()

	lf := logger.NewFields("Login").WithTrace(ctx)

	// Parse request
	var req LoginRequest
	if err := data.FiberCtx.BodyParser(&req); err != nil {
		telemetry.SpanError(ctx, err)
		lf.Append(logger.Any("error", err.Error()))
		logger.Error("Invalid login request", lf)
		return *appctx.NewResponse().
			WithCode(fiber.StatusBadRequest).
			WithErrors("Invalid request body")
	}

	lf.Append(logger.Any("username", req.Username))

	// TODO: Implement user lookup by username from database
	// For now, this is a placeholder - you need to implement GetUserByUsername in repository
	// user, err := u.userRepo.GetUserByUsername(ctx, req.Username)
	// if err != nil {
	//     logger.Error("User not found", lf)
	//     return *appctx.NewResponse().
	//         WithCode(fiber.StatusUnauthorized).
	//         WithErrors("Invalid credentials")
	// }

	// TODO: Verify password
	// if !u.hasher.ComparePassword(req.Password, user.HashedPassword) {
	//     logger.Error("Invalid password", lf)
	//     return *appctx.NewResponse().
	//         WithCode(fiber.StatusUnauthorized).
	//         WithErrors("Invalid credentials")
	// }

	// For demo purposes, using hardcoded user data
	// Replace this with actual database lookup
	userID := int64(1)
	username := req.Username
	email := "user@example.com"
	role := "user"

	// Generate JWT token
	claims := jwt.Claims{
		UserID:   userID,
		Username: username,
		Email:    email,
		Role:     role,
	}

	token, err := u.jwt.Generate(claims)
	if err != nil {
		telemetry.SpanError(ctx, err)
		lf.Append(logger.Any("error", err.Error()))
		logger.Error("Failed to generate token", lf)
		return *appctx.NewResponse().
			WithCode(fiber.StatusInternalServerError).
			WithErrors("Failed to generate token")
	}

	response := LoginResponse{
		Token:     token,
		UserID:    userID,
		Username:  username,
		Email:     email,
		Role:      role,
		ExpiresIn: "24h",
	}

	logger.Info("Login successful", lf)
	return *appctx.NewResponse().WithCode(fiber.StatusOK).WithData(response)
}

// RefreshToken usecase for refreshing JWT token
type refreshToken struct {
	jwt jwt.JWT
}

// RefreshTokenRequest represents refresh token request
type RefreshTokenRequest struct {
	Token string `json:"token" validate:"required"`
}

// RefreshTokenResponse represents refresh token response
type RefreshTokenResponse struct {
	Token     string `json:"token"`
	ExpiresIn string `json:"expires_in"`
}

func NewRefreshToken(jwtInstance jwt.JWT) contract.UseCase {
	return &refreshToken{
		jwt: jwtInstance,
	}
}

func (u *refreshToken) Serve(data appctx.Data) appctx.Response {
	ctx := data.FiberCtx.UserContext()
	ctx, span := telemetry.StartSpan(ctx, "refreshToken.Serve")
	defer span.End()

	lf := logger.NewFields("RefreshToken").WithTrace(ctx)

	// Parse request
	var req RefreshTokenRequest
	if err := data.FiberCtx.BodyParser(&req); err != nil {
		telemetry.SpanError(ctx, err)
		lf.Append(logger.Any("error", err.Error()))
		logger.Error("Invalid refresh token request", lf)
		return *appctx.NewResponse().
			WithCode(fiber.StatusBadRequest).
			WithErrors("Invalid request body")
	}

	// Refresh token
	newToken, err := u.jwt.Refresh(req.Token)
	if err != nil {
		telemetry.SpanError(ctx, err)
		lf.Append(logger.Any("error", err.Error()))
		logger.Error("Failed to refresh token", lf)
		return *appctx.NewResponse().
			WithCode(fiber.StatusUnauthorized).
			WithErrors("Invalid or expired token")
	}

	response := RefreshTokenResponse{
		Token:     newToken,
		ExpiresIn: "24h",
	}

	logger.Info("Token refreshed successfully", lf)
	return *appctx.NewResponse().WithCode(fiber.StatusOK).WithData(response)
}
