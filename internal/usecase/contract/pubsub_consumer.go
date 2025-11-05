package contract

import "github.com/hanifkf12/hanif_skeleton/internal/appctx"

// PubSubConsumer is the contract for Pub/Sub message consumers
// Similar to UseCase interface for HTTP handlers
type PubSubConsumer interface {
	Consume(data appctx.PubSubData) appctx.PubSubResponse
}
