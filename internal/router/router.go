package router

import (
	"github.com/gofiber/fiber/v2"
	"github.com/hanifkf12/hanif_skeleton/internal/appctx"
	"github.com/hanifkf12/hanif_skeleton/internal/bootstrap"
	"github.com/hanifkf12/hanif_skeleton/internal/handler"
	"github.com/hanifkf12/hanif_skeleton/internal/middleware"
	"github.com/hanifkf12/hanif_skeleton/internal/repository/campaign"
	"github.com/hanifkf12/hanif_skeleton/internal/repository/home"
	userRepo "github.com/hanifkf12/hanif_skeleton/internal/repository/user"
	"github.com/hanifkf12/hanif_skeleton/internal/usecase"
	"github.com/hanifkf12/hanif_skeleton/internal/usecase/contract"
	"github.com/hanifkf12/hanif_skeleton/pkg/config"
	"github.com/hanifkf12/hanif_skeleton/pkg/logger"
)

type router struct {
	cfg   *config.Config
	fiber fiber.Router
}

// handle registers a handler without middleware
func (rtr *router) handle(hfn httpHandlerFunc, svc contract.UseCase) fiber.Handler {
	return rtr.handleWithMiddleware(hfn, svc)
}

// handleWithMiddleware registers a handler with optional middlewares
// Middlewares are executed in order, if any returns non-200 code, execution stops
func (rtr *router) handleWithMiddleware(hfn httpHandlerFunc, svc contract.UseCase, middlewares ...middleware.Middleware) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		// Execute middlewares in order
		for _, mw := range middlewares {
			resp := mw(ctx, rtr.cfg)

			// If middleware returns non-200, stop execution and return error response
			if resp.Code != fiber.StatusOK {
				lf := logger.NewFields("Router.Middleware")
				lf.Append(logger.Any("code", resp.Code))
				lf.Append(logger.Any("path", ctx.Path()))
				lf.Append(logger.Any("method", ctx.Method()))
				logger.Error("Middleware validation failed", lf)
				return rtr.response(ctx, resp)
			}
		}

		// All middlewares passed, execute handler
		resp := hfn(ctx, svc, rtr.cfg)
		return rtr.response(ctx, resp)
	}
}

func (rtr *router) response(ctx *fiber.Ctx, resp appctx.Response) error {
	ctx.Set("Content-Type", "application/json; charset=utf-8")

	// Use the response code from appctx.Response
	statusCode := resp.Code
	if statusCode == 0 {
		statusCode = 200
	}

	return ctx.Status(statusCode).Send(resp.Byte())
}

func (rtr *router) Route() {
	db := bootstrap.RegistryDatabase(rtr.cfg, false)
	homeRepo := home.NewHomeRepository(db)
	userRepository := userRepo.NewUserRepository(db)
	campaignRepository := campaign.NewCampaignRepository(db)

	// Initialize JWT
	jwtInstance := bootstrap.RegistryJWT(rtr.cfg)
	hasher := bootstrap.RegistryBcryptHasher(rtr.cfg)

	// Public routes - no middleware
	healthUseCase := usecase.NewHealth(homeRepo)
	rtr.fiber.Get("/health", rtr.handle(
		handler.HttpRequest,
		healthUseCase,
	))

	// Auth routes - public
	loginUseCase := usecase.NewLogin(userRepository, hasher, jwtInstance)
	rtr.fiber.Post("/auth/login", rtr.handle(
		handler.HttpRequest,
		loginUseCase,
	))

	refreshTokenUseCase := usecase.NewRefreshToken(jwtInstance)
	rtr.fiber.Post("/auth/refresh", rtr.handle(
		handler.HttpRequest,
		refreshTokenUseCase,
	))

	// Protected routes with JWT Auth
	userUseCase := usecase.NewUser(userRepository)
	rtr.fiber.Get("/users", rtr.handleWithMiddleware(
		handler.HttpRequest,
		userUseCase,
		middleware.JWTAuth(jwtInstance),
	))

	// Protected route with API Key (alternative auth method)
	campaignUseCase := usecase.NewCampaign(campaignRepository)
	rtr.fiber.Get("/campaigns", rtr.handleWithMiddleware(
		handler.HttpRequest,
		campaignUseCase,
		middleware.APIKeyAuth("X-API-Key", []string{"api-key-123", "api-key-456"}),
	))

	// Protected route with JWT + Content Type validation
	createCampaignUseCase := usecase.NewCreateCampaign(campaignRepository)
	rtr.fiber.Post("/campaigns", rtr.handleWithMiddleware(
		handler.HttpRequest,
		createCampaignUseCase,
		middleware.JWTAuth(jwtInstance),
		middleware.ContentTypeValidator([]string{"application/json"}),
	))

	// Protected route with JWT
	updateCampaignUseCase := usecase.NewUpdateCampaign(campaignRepository)
	rtr.fiber.Put("/campaigns", rtr.handleWithMiddleware(
		handler.HttpRequest,
		updateCampaignUseCase,
		middleware.JWTAuth(jwtInstance),
		middleware.ContentTypeValidator([]string{"application/json"}),
	))

	deleteCampaignUseCase := usecase.NewDeleteCampaign(campaignRepository)
	rtr.fiber.Delete("/campaigns/:id", rtr.handleWithMiddleware(
		handler.HttpRequest,
		deleteCampaignUseCase,
		middleware.JWTAuth(jwtInstance),
	))

	// User routes with JWT + Role-based access control
	createUserUseCase := usecase.NewCreateUser(userRepository)
	rtr.fiber.Post("/users", rtr.handleWithMiddleware(
		handler.HttpRequest,
		createUserUseCase,
		middleware.JWTAuth(jwtInstance),
		middleware.RequireRole([]string{"admin"}), // Only admin can create users
		middleware.ContentTypeValidator([]string{"application/json"}),
	))

	updateUserUseCase := usecase.NewUpdateUser(userRepository)
	rtr.fiber.Put("/users/:id", rtr.handleWithMiddleware(
		handler.HttpRequest,
		updateUserUseCase,
		middleware.JWTAuth(jwtInstance),
		middleware.ContentTypeValidator([]string{"application/json"}),
	))

	deleteUserUseCase := usecase.NewDeleteUser(userRepository)
	rtr.fiber.Delete("/users/:id", rtr.handleWithMiddleware(
		handler.HttpRequest,
		deleteUserUseCase,
		middleware.JWTAuth(jwtInstance),
		middleware.RequireRole([]string{"admin"}), // Only admin can delete
	))

	// Example: HMAC protected endpoint (for webhooks, external APIs, etc.)
	// rtr.fiber.Post("/webhooks/payment", rtr.handleWithMiddleware(
	// 	handler.HttpRequest,
	// 	paymentWebhookUseCase,
	// 	middleware.HMACAuth("your-hmac-secret-key"),
	// 	middleware.ContentTypeValidator([]string{"application/json"}),
	// ))

	// Example: IP whitelist for admin endpoints
	// rtr.fiber.Get("/admin/stats", rtr.handleWithMiddleware(
	// 	handler.HttpRequest,
	// 	statsUseCase,
	// 	middleware.IPWhitelist([]string{"127.0.0.1", "10.0.0.1"}),
	// 	middleware.JWTAuth(jwtInstance),
	// 	middleware.RequireRole([]string{"admin"}),
	// ))

	// Example: Rate limited public endpoint
	// rtr.fiber.Post("/public/contact", rtr.handleWithMiddleware(
	// 	handler.HttpRequest,
	// 	contactUseCase,
	// 	middleware.RateLimit(middleware.RateLimitConfig{
	// 		MaxRequests: 10,
	// 		WindowSize:  60, // 10 requests per 60 seconds
	// 	}),
	// ))
}

func NewRouter(cfg *config.Config, fiber fiber.Router) Router {
	return &router{
		cfg:   cfg,
		fiber: fiber,
	}
}
