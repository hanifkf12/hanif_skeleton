package pubsub

import (
	"context"
	"encoding/json"
	"time"

	"cloud.google.com/go/pubsub"
	"github.com/hanifkf12/hanif_skeleton/pkg/logger"
	"github.com/hanifkf12/hanif_skeleton/pkg/telemetry"
)

// Publisher wraps Google Cloud Pub/Sub client for publishing messages
type Publisher interface {
	Publish(ctx context.Context, topicID string, data interface{}) (string, error)
	PublishWithAttributes(ctx context.Context, topicID string, data interface{}, attributes map[string]string) (string, error)
	Close() error
}

type publisher struct {
	client *pubsub.Client
}

// NewPublisher creates a new Publisher instance
func NewPublisher(client *pubsub.Client) Publisher {
	return &publisher{
		client: client,
	}
}

// Publish publishes a message to a topic
func (p *publisher) Publish(ctx context.Context, topicID string, data interface{}) (string, error) {
	return p.PublishWithAttributes(ctx, topicID, data, nil)
}

// PublishWithAttributes publishes a message with custom attributes to a topic
func (p *publisher) PublishWithAttributes(ctx context.Context, topicID string, data interface{}, attributes map[string]string) (string, error) {
	ctx, span := telemetry.StartSpan(ctx, "publisher.Publish")
	defer span.End()

	lf := logger.NewFields("PubSubPublisher").WithTrace(ctx)
	lf.Append(logger.Any("topic_id", topicID))

	// Marshal data to JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		telemetry.SpanError(ctx, err)
		lf.Append(logger.Any("error", err.Error()))
		logger.Error("Failed to marshal message data", lf)
		return "", err
	}

	topic := p.client.Topic(topicID)
	defer topic.Stop()

	// Create message
	msg := &pubsub.Message{
		Data:       jsonData,
		Attributes: attributes,
	}

	// Add timestamp if not provided
	if msg.Attributes == nil {
		msg.Attributes = make(map[string]string)
	}
	if _, ok := msg.Attributes["timestamp"]; !ok {
		msg.Attributes["timestamp"] = time.Now().Format(time.RFC3339)
	}

	// Publish message
	result := topic.Publish(ctx, msg)

	// Block until the result is returned and a server-generated ID is returned for the published message
	messageID, err := result.Get(ctx)
	if err != nil {
		telemetry.SpanError(ctx, err)
		lf.Append(logger.Any("error", err.Error()))
		logger.Error("Failed to publish message", lf)
		return "", err
	}

	lf.Append(logger.Any("message_id", messageID))
	logger.Info("Message published successfully", lf)

	return messageID, nil
}

// Close closes the underlying Pub/Sub client
func (p *publisher) Close() error {
	if p.client != nil {
		return p.client.Close()
	}
	return nil
}
