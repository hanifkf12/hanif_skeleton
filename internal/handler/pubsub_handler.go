package handler

import (
	"context"

	"cloud.google.com/go/pubsub"
	"github.com/hanifkf12/hanif_skeleton/internal/appctx"
	"github.com/hanifkf12/hanif_skeleton/internal/usecase/contract"
	"github.com/hanifkf12/hanif_skeleton/pkg/config"
)

// PubSubHandler is a general handler for Pub/Sub messages
// Similar to HttpRequest handler for HTTP
func PubSubHandler(ctx context.Context, msg *pubsub.Message, consumer contract.PubSubConsumer, conf *config.Config) appctx.PubSubResponse {
	data := appctx.PubSubData{
		Ctx:     ctx,
		Message: msg,
		Cfg:     conf,
	}

	return consumer.Consume(data)
}
