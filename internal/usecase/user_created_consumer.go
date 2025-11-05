package usecase

import (
	"encoding/json"

	"github.com/hanifkf12/hanif_skeleton/internal/appctx"
	"github.com/hanifkf12/hanif_skeleton/internal/entity"
	"github.com/hanifkf12/hanif_skeleton/internal/repository"
	"github.com/hanifkf12/hanif_skeleton/internal/usecase/contract"
	"github.com/hanifkf12/hanif_skeleton/pkg/logger"
	"github.com/hanifkf12/hanif_skeleton/pkg/telemetry"
)

// Example Pub/Sub consumer for creating users from Pub/Sub messages
type userCreatedConsumer struct {
	userRepo repository.UserRepository
}

func NewUserCreatedConsumer(userRepo repository.UserRepository) contract.PubSubConsumer {
	return &userCreatedConsumer{
		userRepo: userRepo,
	}
}

func (c *userCreatedConsumer) Consume(data appctx.PubSubData) appctx.PubSubResponse {
	ctx, span := telemetry.StartSpan(data.Ctx, "userCreatedConsumer.Consume")
	defer span.End()

	lf := logger.NewFields("UserCreatedConsumer").WithTrace(ctx)
	lf.Append(logger.Any("message_id", data.Message.ID))

	logger.Info("Processing user created message", lf)

	// Parse message data
	var req entity.CreateUserRequest
	if err := json.Unmarshal(data.Message.Data, &req); err != nil {
		telemetry.SpanError(ctx, err)
		lf.Append(logger.Any("error", err.Error()))
		logger.Error("Failed to parse message data", lf)
		return *appctx.NewPubSubResponse().WithError(err)
	}

	lf.Append(logger.Any("username", req.Username))
	lf.Append(logger.Any("email", req.Email))

	// Create user in database
	userID, err := c.userRepo.CreateUser(ctx, req)
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
