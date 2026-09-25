package usecase

import (
	"encoding/json"

	"github.com/go-playground/validator/v10"
	"github.com/hanifkf12/hanif_skeleton/internal/appctx"
	"github.com/hanifkf12/hanif_skeleton/internal/entity"
	"github.com/hanifkf12/hanif_skeleton/internal/repository"
	"github.com/hanifkf12/hanif_skeleton/internal/usecase/contract"
	"github.com/hanifkf12/hanif_skeleton/pkg/crypto"
	"github.com/hanifkf12/hanif_skeleton/pkg/logger"
	"github.com/hanifkf12/hanif_skeleton/pkg/telemetry"
)

// userCreatedPayload adalah bentuk wire pesan Pub/Sub. Ia sengaja terpisah dari
// entity.UserInput: payload membawa password plaintext, sedangkan UserInput
// hanya bisa merepresentasikan hash. Kedua entry point (HTTP & Pub/Sub)
// bertemu di UserInput, sehingga invariant-nya sama persis.
type userCreatedPayload struct {
	Username string `json:"username" validate:"required"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6,max=72"`
}

// Example Pub/Sub consumer for creating users from Pub/Sub messages
type userCreatedConsumer struct {
	userRepo  repository.UserRepository
	hasher    *crypto.BcryptHasher
	validator *validator.Validate
}

func NewUserCreatedConsumer(userRepo repository.UserRepository, hasher *crypto.BcryptHasher) contract.PubSubConsumer {
	return &userCreatedConsumer{
		userRepo:  userRepo,
		hasher:    hasher,
		validator: validator.New(),
	}
}

func (c *userCreatedConsumer) Consume(data appctx.PubSubData) appctx.PubSubResponse {
	ctx, span := telemetry.StartSpan(data.Ctx, "userCreatedConsumer.Consume")
	defer span.End()

	lf := logger.NewFields("UserCreatedConsumer").WithTrace(ctx)
	lf.Append(logger.Any("message_id", data.Message.ID))

	logger.Info("Processing user created message", lf)

	// Parse message data
	var payload userCreatedPayload
	if err := json.Unmarshal(data.Message.Data, &payload); err != nil {
		telemetry.SpanError(ctx, err)
		lf.Append(logger.Any("error", err.Error()))
		logger.Error("Failed to parse message data", lf)
		return *appctx.NewPubSubResponse().WithError(err)
	}
	if err := c.validator.Struct(payload); err != nil {
		telemetry.SpanError(ctx, err)
		logger.Error("Invalid user created message", lf)
		return *appctx.NewPubSubResponse().WithError(err)
	}

	passwordHash, err := c.hasher.HashPassword(payload.Password)
	if err != nil {
		telemetry.SpanError(ctx, err)
		logger.Error("Failed to hash user password", lf)
		return *appctx.NewPubSubResponse().WithError(err)
	}

	input, err := entity.NewUserInput(payload.Username, payload.Email, passwordHash)
	if err != nil {
		telemetry.SpanError(ctx, err)
		logger.Error("Invalid user data in message", lf)
		return *appctx.NewPubSubResponse().WithError(err)
	}

	lf.Append(logger.Any("username", input.Username))
	lf.Append(logger.Any("email", input.Email))

	// Create user in database
	userID, err := c.userRepo.CreateUser(ctx, input)
	if err != nil {
		telemetry.SpanError(ctx, err)
		lf.Append(logger.Any("error", err.Error()))
		logger.Error("Failed to create user from Pub/Sub message", lf)
		return *appctx.NewPubSubResponse().WithError(err)
	}

	lf.Append(logger.Any("user_id", userID))
	logger.Info("User created successfully from Pub/Sub message", lf)

	return *appctx.NewPubSubResponse().WithMessage("User created successfully")
}
