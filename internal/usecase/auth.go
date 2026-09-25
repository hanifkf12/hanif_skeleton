package usecase

import (
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/hanifkf12/hanif_skeleton/internal/appctx"
	"github.com/hanifkf12/hanif_skeleton/internal/repository"
	"github.com/hanifkf12/hanif_skeleton/internal/usecase/contract"
	"github.com/hanifkf12/hanif_skeleton/pkg/apperror"
	"github.com/hanifkf12/hanif_skeleton/pkg/crypto"
	"github.com/hanifkf12/hanif_skeleton/pkg/jwt"
	"github.com/hanifkf12/hanif_skeleton/pkg/logger"
	"github.com/hanifkf12/hanif_skeleton/pkg/telemetry"
)

// Login usecase for user authentication
type login struct {
	userRepo  repository.UserRepository
	hasher    *crypto.BcryptHasher
	jwt       jwt.JWT
	validator *validator.Validate
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
		userRepo:  userRepo,
		hasher:    hasher,
		jwt:       jwtInstance,
		validator: validator.New(),
	}
}

func (u *login) Serve(data appctx.Data) appctx.Response {
	ctx := data.Request.Context()
	ctx, span := telemetry.StartSpan(ctx, "login.Serve")
	defer span.End()

	lf := logger.NewFields("Login").WithTrace(ctx)

	// Parse request
	var req LoginRequest
	if err := data.Request.Body(&req); err != nil {
		telemetry.SpanError(ctx, err)
		lf.Append(logger.Any("error", err.Error()))
		logger.Error("Invalid login request", lf)
		return *appctx.ResponseFromError(apperror.Invalid("Invalid request body"))
	}
	req.Username = strings.TrimSpace(req.Username)
	if err := u.validator.Struct(req); err != nil {
		telemetry.SpanError(ctx, err)
		return *appctx.ResponseFromError(apperror.Invalid("Username and password are required"))
	}

	lf.Append(logger.Any("username", req.Username))

	user, err := u.userRepo.GetUserByUsername(ctx, req.Username)
	if err != nil {
		// NotFound sengaja tidak di-log: ia adalah hasil normal dari username
		// yang tidak dikenal, dan permintaan ini tetap dijawab 401 agar klien
		// tidak bisa membedakan "user tidak ada" dari "password salah".
		if apperror.KindOf(err) != apperror.KindNotFound {
			telemetry.SpanError(ctx, err)
			lf.Append(logger.Any("error", err.Error()))
			logger.Error("Failed to look up user", lf)
		}
		return *appctx.ResponseFromError(apperror.Unauthorized("Invalid credentials"))
	}

	if user.PasswordHash == "" || !u.hasher.ComparePassword(req.Password, user.PasswordHash) {
		logger.Error("Login rejected", lf)
		return *appctx.ResponseFromError(apperror.Unauthorized("Invalid credentials"))
	}

	// Generate JWT token
	claims := jwt.Claims{
		UserID:   user.ID,
		Username: user.Username,
		Email:    user.Email,
		Role:     user.Role,
	}

	token, err := u.jwt.Generate(claims)
	if err != nil {
		telemetry.SpanError(ctx, err)
		lf.Append(logger.Any("error", err.Error()))
		logger.Error("Failed to generate token", lf)
		return *appctx.ResponseFromError(apperror.Internal(err))
	}

	response := LoginResponse{
		Token:     token,
		UserID:    user.ID,
		Username:  user.Username,
		Email:     user.Email,
		Role:      user.Role,
		ExpiresIn: "24h",
	}

	logger.Info("Login successful", lf)
	return *appctx.NewResponse().WithCode(appctx.StatusOK).WithData(response)
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
	ctx := data.Request.Context()
	ctx, span := telemetry.StartSpan(ctx, "refreshToken.Serve")
	defer span.End()

	lf := logger.NewFields("RefreshToken").WithTrace(ctx)

	// Parse request
	var req RefreshTokenRequest
	if err := data.Request.Body(&req); err != nil {
		telemetry.SpanError(ctx, err)
		lf.Append(logger.Any("error", err.Error()))
		logger.Error("Invalid refresh token request", lf)
		return *appctx.ResponseFromError(apperror.Invalid("Invalid request body"))
	}

	// Refresh token
	newToken, err := u.jwt.Refresh(req.Token)
	if err != nil {
		telemetry.SpanError(ctx, err)
		lf.Append(logger.Any("error", err.Error()))
		logger.Error("Failed to refresh token", lf)
		return *appctx.ResponseFromError(apperror.Unauthorized("Invalid or expired token"))
	}

	response := RefreshTokenResponse{
		Token:     newToken,
		ExpiresIn: "24h",
	}

	logger.Info("Token refreshed successfully", lf)
	return *appctx.NewResponse().WithCode(appctx.StatusOK).WithData(response)
}
