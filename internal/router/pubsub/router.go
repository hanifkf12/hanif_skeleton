package pubsub

import (
	"context"

	"cloud.google.com/go/pubsub"
	"github.com/hanifkf12/hanif_skeleton/internal/handler"
	"github.com/hanifkf12/hanif_skeleton/internal/usecase/contract"
	"github.com/hanifkf12/hanif_skeleton/pkg/config"
	"github.com/hanifkf12/hanif_skeleton/pkg/logger"
)

// ConsumerHandlerFunc wraps the consumer with handler
type ConsumerHandlerFunc func(ctx context.Context, msg *pubsub.Message, consumer contract.PubSubConsumer, conf *config.Config)

// SubscriptionConfig holds configuration for a subscription
type SubscriptionConfig struct {
	SubscriptionID string
	Consumer       contract.PubSubConsumer
	MaxConcurrent  int // max concurrent messages to process, default 10
}

// Router manages Pub/Sub subscriptions
type Router interface {
	RegisterSubscription(config SubscriptionConfig)
	Start(ctx context.Context) error
	Stop() error
}

type router struct {
	cfg           *config.Config
	client        *pubsub.Client
	subscriptions []SubscriptionConfig
}

// RegisterSubscription registers a new subscription handler
func (r *router) RegisterSubscription(config SubscriptionConfig) {
	if config.MaxConcurrent == 0 {
		config.MaxConcurrent = 10
	}
	r.subscriptions = append(r.subscriptions, config)
	logger.Info("Registered Pub/Sub subscription", logger.NewFields(config.SubscriptionID))
}

// Start begins consuming messages from all registered subscriptions
func (r *router) Start(ctx context.Context) error {
	lf := logger.NewFields("PubSubRouter.Start")

	if len(r.subscriptions) == 0 {
		logger.Info("No subscriptions registered", lf)
		return nil
	}

	lf.Append(logger.Any("subscriptions", len(r.subscriptions)))
	logger.Info("Starting Pub/Sub consumer", lf)

	errChan := make(chan error, len(r.subscriptions))

	for _, subConfig := range r.subscriptions {
		go func(sc SubscriptionConfig) {
			sub := r.client.Subscription(sc.SubscriptionID)
			sub.ReceiveSettings.MaxOutstandingMessages = sc.MaxConcurrent

			subLogger := logger.NewFields(sc.SubscriptionID)
			logger.Info("Starting subscription consumer", subLogger)

			err := sub.Receive(ctx, func(ctx context.Context, msg *pubsub.Message) {
				msgLogger := logger.NewFields(sc.SubscriptionID)
				msgLogger.Append(logger.Any("message_id", msg.ID))
				msgLogger.Append(logger.Any("publish_time", msg.PublishTime))

				logger.Info("Received message", msgLogger)

				// Call the handler (similar to HTTP handler pattern)
				resp := handler.PubSubHandler(ctx, msg, sc.Consumer, r.cfg)

				if resp.Success {
					msg.Ack()
					logger.Info("Message processed successfully", msgLogger)
				} else {
					msg.Nack()
					msgLogger.Append(logger.Any("error", resp.Error))
					logger.Error("Message processing failed", msgLogger)
				}
			})

			if err != nil {
				subLogger.Append(logger.Any("error", err))
				logger.Error("Subscription receive error", subLogger)
				errChan <- err
			}
		}(subConfig)
	}

	// Wait for context cancellation or error
	select {
	case <-ctx.Done():
		logger.Info("Pub/Sub consumer context cancelled", lf)
		return ctx.Err()
	case err := <-errChan:
		lf.Append(logger.Any("error", err))
		logger.Error("Pub/Sub consumer error", lf)
		return err
	}
}

// Stop gracefully stops the Pub/Sub consumer
func (r *router) Stop() error {
	lf := logger.NewFields("PubSubRouter.Stop")
	logger.Info("Stopping Pub/Sub consumer", lf)

	if r.client != nil {
		return r.client.Close()
	}
	return nil
}

// NewRouter creates a new Pub/Sub router
func NewRouter(cfg *config.Config, client *pubsub.Client) Router {
	return &router{
		cfg:           cfg,
		client:        client,
		subscriptions: make([]SubscriptionConfig, 0),
	}
}
