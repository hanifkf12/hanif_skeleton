package pubsub

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"cloud.google.com/go/pubsub"
	"github.com/hanifkf12/hanif_skeleton/internal/bootstrap"
	userRepo "github.com/hanifkf12/hanif_skeleton/internal/repository/user"
	pubsubRouter "github.com/hanifkf12/hanif_skeleton/internal/router/pubsub"
	"github.com/hanifkf12/hanif_skeleton/internal/usecase"
	"github.com/hanifkf12/hanif_skeleton/pkg/config"
	"github.com/hanifkf12/hanif_skeleton/pkg/logger"
	"github.com/hanifkf12/hanif_skeleton/pkg/telemetry"
)

func Start() {
	logger.Setup()
	defer logger.Cleanup()

	cfg, err := config.LoadAllConfigs()
	if err != nil {
		logger.Fatal(err.Error())
	}

	// Initialize tracer
	cleanup, err := telemetry.InitTracer("hanif-skeleton-pubsub")
	if err != nil {
		logger.Fatal(err.Error())
	}
	defer cleanup()

	// Get Google Cloud project ID from config or environment
	projectID := os.Getenv("GOOGLE_CLOUD_PROJECT")
	if projectID == "" {
		logger.Fatal("GOOGLE_CLOUD_PROJECT environment variable is required")
	}

	ctx := context.Background()

	// Initialize Pub/Sub client
	client, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		logger.Fatal("Failed to create Pub/Sub client: " + err.Error())
	}
	defer client.Close()

	// Initialize database and repositories
	db := bootstrap.RegistryDatabase(cfg, false)
	userRepository := userRepo.NewUserRepository(db)

	// Create Pub/Sub router
	router := pubsubRouter.NewRouter(cfg, client)

	// Register subscriptions with their consumers
	// Example: Register user-created-subscription
	router.RegisterSubscription(pubsubRouter.SubscriptionConfig{
		SubscriptionID: "user-created-subscription", // Change to your actual subscription ID
		Consumer:       usecase.NewUserCreatedConsumer(userRepository),
		MaxConcurrent:  10,
	})

	// Add more subscriptions here as needed
	// router.RegisterSubscription(pubsubRouter.SubscriptionConfig{
	//     SubscriptionID: "another-subscription",
	//     Consumer:       usecase.NewAnotherConsumer(someRepo),
	//     MaxConcurrent:  5,
	// })

	logger.Info("Starting Pub/Sub worker")

	// Create context that can be cancelled
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		logger.Info("Received shutdown signal")
		cancel()
	}()

	// Start consuming messages
	if err := router.Start(ctx); err != nil && err != context.Canceled {
		logger.Fatal("Pub/Sub worker error: " + err.Error())
	}

	logger.Info("Pub/Sub worker stopped gracefully")
}
