# Contoh Penggunaan Publisher dalam HTTP Endpoint

## Scenario: User Registration dengan Event Publishing

Ketika user baru register via HTTP, kita ingin:
1. Simpan user ke database
2. Publish event `user.created` ke Pub/Sub
3. Consumer lain akan handle email notification, analytics, etc.

## Implementation

### 1. Update Create User UseCase dengan Publisher

File: `internal/usecase/create_user.go`

```go
package usecase

import (
	"github.com/gofiber/fiber/v2"
	"github.com/hanifkf12/hanif_skeleton/internal/appctx"
	"github.com/hanifkf12/hanif_skeleton/internal/entity"
	"github.com/hanifkf12/hanif_skeleton/internal/repository"
	"github.com/hanifkf12/hanif_skeleton/internal/usecase/contract"
	"github.com/hanifkf12/hanif_skeleton/pkg/logger"
	"github.com/hanifkf12/hanif_skeleton/pkg/pubsub"
	"github.com/hanifkf12/hanif_skeleton/pkg/telemetry"
)

type createUser struct {
	userRepo  repository.UserRepository
	publisher pubsub.Publisher // Publisher injection
}

// Constructor with optional publisher (nil-safe)
func NewCreateUser(userRepo repository.UserRepository, publisher pubsub.Publisher) contract.UseCase {
	return &createUser{
		userRepo:  userRepo,
		publisher: publisher,
	}
}

func (u *createUser) Serve(data appctx.Data) appctx.Response {
	ctx := data.FiberCtx.UserContext()
	ctx, span := telemetry.StartSpan(ctx, "createUser.Serve")
	defer span.End()

	lf := logger.NewFields("CreateUser").WithTrace(ctx)

	// Parse request body
	req := new(entity.CreateUserRequest)
	if err := data.FiberCtx.BodyParser(req); err != nil {
		telemetry.SpanError(ctx, err)
		lf.Append(logger.Any("error", err.Error()))
		logger.Error("Invalid create user request", lf)
		return *appctx.NewResponse().WithCode(fiber.StatusBadRequest).WithErrors(err.Error())
	}

	lf.Append(logger.Any("username", req.Username))
	lf.Append(logger.Any("email", req.Email))

	// Create user in database
	userID, err := u.userRepo.CreateUser(ctx, *req)
	if err != nil {
		telemetry.SpanError(ctx, err)
		lf.Append(logger.Any("error", err.Error()))
		logger.Error("Failed to create user", lf)
		return *appctx.NewResponse().WithCode(fiber.StatusInternalServerError).WithErrors(err.Error())
	}

	lf.Append(logger.Any("user_id", userID))

	// Publish event asynchronously (if publisher is configured)
	if u.publisher != nil {
		go func() {
			eventCtx := context.Background() // Use separate context for async operation
			eventData := map[string]interface{}{
				"user_id":  userID,
				"username": req.Username,
				"email":    req.Email,
				"action":   "user_created",
			}

			messageID, err := u.publisher.PublishWithAttributes(
				eventCtx,
				"user-events", // Topic ID
				eventData,
				map[string]string{
					"event_type": "user.created",
					"version":    "1.0",
				},
			)

			if err != nil {
				logger.Error("Failed to publish user created event",
					logger.NewFields("CreateUser").
						Append(logger.Any("user_id", userID)).
						Append(logger.Any("error", err.Error())))
			} else {
				logger.Info("User created event published",
					logger.NewFields("CreateUser").
						Append(logger.Any("user_id", userID)).
						Append(logger.Any("message_id", messageID)))
			}
		}()
	}

	// Prepare response
	resp := entity.CreateUserResponse{
		ID:       userID,
		Username: req.Username,
		Email:    req.Email,
	}

	logger.Info("User created successfully", lf)
	return *appctx.NewResponse().WithData(resp)
}
```

### 2. Update Router untuk Inject Publisher

File: `internal/router/router.go`

```go
package router

import (
	"context"
	"os"
	
	pubsubClient "cloud.google.com/go/pubsub"
	"github.com/gofiber/fiber/v2"
	"github.com/hanifkf12/hanif_skeleton/internal/appctx"
	"github.com/hanifkf12/hanif_skeleton/internal/bootstrap"
	"github.com/hanifkf12/hanif_skeleton/internal/handler"
	"github.com/hanifkf12/hanif_skeleton/internal/repository/campaign"
	"github.com/hanifkf12/hanif_skeleton/internal/repository/home"
	userRepo "github.com/hanifkf12/hanif_skeleton/internal/repository/user"
	"github.com/hanifkf12/hanif_skeleton/internal/usecase"
	"github.com/hanifkf12/hanif_skeleton/internal/usecase/contract"
	"github.com/hanifkf12/hanif_skeleton/pkg/config"
	"github.com/hanifkf12/hanif_skeleton/pkg/logger"
	"github.com/hanifkf12/hanif_skeleton/pkg/pubsub"
)

type router struct {
	cfg       *config.Config
	fiber     fiber.Router
	publisher pubsub.Publisher
}

func (rtr *router) handle(hfn httpHandlerFunc, svc contract.UseCase) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		resp := hfn(ctx, svc, rtr.cfg)
		return rtr.response(ctx, resp)
	}
}

func (rtr *router) response(ctx *fiber.Ctx, resp appctx.Response) error {
	ctx.Set("Content-Type", "application/json; charset=utf-8")
	return ctx.Status(200).Send(resp.Byte())
}

func (rtr *router) Route() {
	db := bootstrap.RegistryDatabase(rtr.cfg, false)
	homeRepo := home.NewHomeRepository(db)
	userRepository := userRepo.NewUserRepository(db)
	campaignRepository := campaign.NewCampaignRepository(db)

	healthUseCase := usecase.NewHealth(homeRepo)
	rtr.fiber.Get("/health", rtr.handle(
		handler.HttpRequest,
		healthUseCase,
	))

	// User routes with publisher
	userUseCase := usecase.NewUser(userRepository)
	rtr.fiber.Get("/users", rtr.handle(
		handler.HttpRequest,
		userUseCase,
	))

	// Create user with event publishing
	createUserUseCase := usecase.NewCreateUser(userRepository, rtr.publisher)
	rtr.fiber.Post("/users", rtr.handle(
		handler.HttpRequest,
		createUserUseCase,
	))

	updateUserUseCase := usecase.NewUpdateUser(userRepository)
	rtr.fiber.Put("/users/:id", rtr.handle(
		handler.HttpRequest,
		updateUserUseCase,
	))

	deleteUserUseCase := usecase.NewDeleteUser(userRepository)
	rtr.fiber.Delete("/users/:id", rtr.handle(
		handler.HttpRequest,
		deleteUserUseCase,
	))

	// Campaign routes
	campaignUseCase := usecase.NewCampaign(campaignRepository)
	rtr.fiber.Get("/campaigns", rtr.handle(
		handler.HttpRequest,
		campaignUseCase,
	))

	createCampaignUseCase := usecase.NewCreateCampaign(campaignRepository)
	rtr.fiber.Post("/campaigns", rtr.handle(
		handler.HttpRequest,
		createCampaignUseCase,
	))

	updateCampaignUseCase := usecase.NewUpdateCampaign(campaignRepository)
	rtr.fiber.Put("/campaigns", rtr.handle(
		handler.HttpRequest,
		updateCampaignUseCase,
	))

	deleteCampaignUseCase := usecase.NewDeleteCampaign(campaignRepository)
	rtr.fiber.Delete("/campaigns/:id", rtr.handle(
		handler.HttpRequest,
		deleteCampaignUseCase,
	))
}

func NewRouter(cfg *config.Config, fiber fiber.Router) Router {
	// Initialize Pub/Sub publisher (optional)
	var publisher pubsub.Publisher
	
	projectID := os.Getenv("GOOGLE_CLOUD_PROJECT")
	if projectID != "" {
		ctx := context.Background()
		client, err := pubsubClient.NewClient(ctx, projectID)
		if err != nil {
			lf := logger.NewFields("Router.NewRouter")
			lf.Append(logger.Any("error", err.Error()))
			logger.Error("Failed to create Pub/Sub client, publisher will be disabled", lf)
		} else {
			publisher = pubsub.NewPublisher(client)
			logger.Info("Pub/Sub publisher initialized", logger.NewFields("Router.NewRouter"))
		}
	} else {
		logger.Info("GOOGLE_CLOUD_PROJECT not set, publisher disabled", logger.NewFields("Router.NewRouter"))
	}

	return &router{
		cfg:       cfg,
		fiber:     fiber,
		publisher: publisher,
	}
}
```

### 3. Create Email Notification Consumer

File: `internal/usecase/user_email_notification_consumer.go`

```go
package usecase

import (
	"encoding/json"
	"github.com/hanifkf12/hanif_skeleton/internal/appctx"
	"github.com/hanifkf12/hanif_skeleton/internal/usecase/contract"
	"github.com/hanifkf12/hanif_skeleton/pkg/logger"
	"github.com/hanifkf12/hanif_skeleton/pkg/telemetry"
)

// Email notification consumer for user events
type userEmailNotificationConsumer struct {
	// Add email service dependency here
	// emailService email.Service
}

func NewUserEmailNotificationConsumer() contract.PubSubConsumer {
	return &userEmailNotificationConsumer{}
}

func (c *userEmailNotificationConsumer) Consume(data appctx.PubSubData) appctx.PubSubResponse {
	ctx, span := telemetry.StartSpan(data.Ctx, "userEmailNotificationConsumer.Consume")
	defer span.End()

	lf := logger.NewFields("UserEmailNotificationConsumer").WithTrace(ctx)
	lf.Append(logger.Any("message_id", data.Message.ID))

	// Get event type from attributes
	eventType := data.Message.Attributes["event_type"]
	lf.Append(logger.Any("event_type", eventType))

	logger.Info("Processing user email notification", lf)

	// Parse event data
	var eventData map[string]interface{}
	if err := json.Unmarshal(data.Message.Data, &eventData); err != nil {
		telemetry.SpanError(ctx, err)
		lf.Append(logger.Any("error", err.Error()))
		logger.Error("Failed to parse event data", lf)
		return *appctx.NewPubSubResponse().WithError(err)
	}

	// Extract user info
	email, _ := eventData["email"].(string)
	username, _ := eventData["username"].(string)

	lf.Append(logger.Any("email", email))
	lf.Append(logger.Any("username", username))

	// Send welcome email based on event type
	if eventType == "user.created" {
		// TODO: Implement actual email sending
		// err := c.emailService.SendWelcomeEmail(ctx, email, username)
		logger.Info("Welcome email sent (mock)", lf)
	}

	logger.Info("User email notification processed successfully", lf)
	return *appctx.NewPubSubResponse().WithMessage("Email notification sent")
}
```

### 4. Register Email Notification Consumer

File: `cmd/pubsub/pubsub.go` (update)

```go
// Register user email notification consumer
router.RegisterSubscription(pubsubRouter.SubscriptionConfig{
	SubscriptionID: "user-events-email-subscription",
	Consumer:       usecase.NewUserEmailNotificationConsumer(),
	MaxConcurrent:  5,
})
```

## Testing

### 1. Test HTTP Endpoint dengan Publisher

```bash
# Create user via HTTP API
curl -X POST http://localhost:8080/users \
  -H "Content-Type: application/json" \
  -d '{
    "username": "john_doe",
    "email": "john@example.com",
    "password": "secret123"
  }'

# Response:
# {
#   "code": 200,
#   "status": true,
#   "data": {
#     "id": 123,
#     "username": "john_doe",
#     "email": "john@example.com"
#   }
# }

# Event akan dipublish ke topic "user-events"
# Consumer akan menerima dan process event tersebut
```

### 2. Monitor Logs

```bash
# Terminal 1: HTTP Server
go run main.go http

# Terminal 2: Pub/Sub Worker
export GOOGLE_CLOUD_PROJECT=your-project-id
go run main.go pubsub

# You should see logs:
# HTTP Server: "User created successfully"
# HTTP Server: "User created event published" with message_id
# Pub/Sub Worker: "Received message" with message_id
# Pub/Sub Worker: "Processing user email notification"
# Pub/Sub Worker: "Welcome email sent"
# Pub/Sub Worker: "Message processed successfully"
```

## Flow Diagram

```
┌──────────┐
│  Client  │
└────┬─────┘
     │ POST /users
     ▼
┌─────────────────┐
│  HTTP Handler   │
└────┬────────────┘
     │
     ▼
┌─────────────────┐
│ Create User UC  │──┐ Save to DB
└────┬────────────┘  │
     │                ▼
     │         ┌──────────────┐
     │         │  Repository  │
     │         └──────────────┘
     │
     │ Publish Event (async)
     ▼
┌─────────────────┐
│  Pub/Sub Topic  │
│  "user-events"  │
└────┬────────────┘
     │
     ├─────────────────────┐
     │                     │
     ▼                     ▼
┌─────────────┐    ┌──────────────┐
│ Email Sub   │    │ Analytics Sub│
└─────┬───────┘    └──────┬───────┘
      │                   │
      ▼                   ▼
┌─────────────┐    ┌──────────────┐
│ Email       │    │ Analytics    │
│ Consumer    │    │ Consumer     │
└─────────────┘    └──────────────┘
```

## Benefits

1. **Decoupling**: HTTP handler tidak perlu tahu tentang email/analytics
2. **Scalability**: Consumer dapat di-scale independent
3. **Reliability**: Message akan di-retry jika gagal
4. **Async**: HTTP response cepat, heavy processing di background
5. **Multiple Consumers**: Satu event bisa ditangani banyak consumer

## Environment Variables

```bash
# Required for publisher
GOOGLE_CLOUD_PROJECT=your-gcp-project-id

# Optional: Use service account key
GOOGLE_APPLICATION_CREDENTIALS=/path/to/service-account-key.json

# If not set, publisher will be disabled (nil-safe)
```

## Graceful Degradation

Publisher di-inject sebagai optional dependency:
- Jika Pub/Sub not configured → Publisher = nil
- UseCase tetap berfungsi normal
- Event publishing di-skip
- No errors, graceful degradation

```go
// In usecase
if u.publisher != nil {
    // Publish event
} else {
    // Skip publishing, continue normal flow
}
```

## Best Practices

1. **Async Publishing**: Publish dalam goroutine agar tidak block response
2. **Separate Context**: Gunakan context terpisah untuk async operation
3. **Error Logging**: Log error publishing tapi jangan fail request
4. **Idempotency**: Consumer harus idempoten (safe untuk retry)
5. **Event Versioning**: Tambahkan version di attributes untuk backward compatibility

## Next Steps

1. Implement actual email service integration
2. Add analytics consumer
3. Setup monitoring/alerting for failed messages
4. Configure dead letter queue
5. Add integration tests

