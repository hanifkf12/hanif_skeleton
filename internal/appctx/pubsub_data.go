package appctx

import (
	"cloud.google.com/go/pubsub"
	"context"
	"github.com/hanifkf12/hanif_skeleton/pkg/config"
)

// PubSubData contains context and data for Pub/Sub message processing
// Similar to Data struct for HTTP requests
type PubSubData struct {
	Ctx     context.Context
	Message *pubsub.Message
	Cfg     *config.Config
}
