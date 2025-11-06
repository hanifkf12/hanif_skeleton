package jobs

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/hanifkf12/hanif_skeleton/pkg/cache"
	"github.com/hanifkf12/hanif_skeleton/pkg/httpclient"
	"github.com/hanifkf12/hanif_skeleton/pkg/logger"
	"github.com/hanifkf12/hanif_skeleton/pkg/queue"
)

// SyncDataJob handles data synchronization with external API
type SyncDataJob struct {
	httpClient httpclient.HTTPClient
	cache      cache.Cache
}

// SyncDataPayload is the payload for sync data job
type SyncDataPayload struct {
	EntityType string `json:"entity_type"`
	EntityID   string `json:"entity_id"`
	Action     string `json:"action"` // create, update, delete
}

// NewSyncDataJob creates a new sync data job handler
func NewSyncDataJob(
	httpClient httpclient.HTTPClient,
	cache cache.Cache,
) queue.JobHandler {
	job := &SyncDataJob{
		httpClient: httpClient,
		cache:      cache,
	}
	return job.Handle
}

// Handle processes the sync data job
func (j *SyncDataJob) Handle(ctx context.Context, payload []byte) error {
	lf := logger.NewFields("SyncDataJob")

	// Unmarshal payload
	var data SyncDataPayload
	if err := json.Unmarshal(payload, &data); err != nil {
		lf.Append(logger.Any("error", err.Error()))
		logger.Error("Failed to unmarshal payload", lf)
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	lf.Append(logger.Any("entity_type", data.EntityType))
	lf.Append(logger.Any("entity_id", data.EntityID))
	lf.Append(logger.Any("action", data.Action))

	logger.Info("Starting data sync", lf)

	// Get data from cache if exists
	cacheKey := cache.NewCacheKey("entity").Build(data.EntityType, data.EntityID)
	cachedData, _ := j.cache.Get(ctx, cacheKey)

	// Prepare sync payload
	syncPayload := map[string]interface{}{
		"entity_type": data.EntityType,
		"entity_id":   data.EntityID,
		"action":      data.Action,
		"timestamp":   time.Now(),
		"data":        cachedData,
	}

	// Call external API to sync data
	headers := map[string]string{
		"X-API-Key": "your-sync-api-key",
	}

	resp, err := j.httpClient.Post(
		ctx,
		"https://api.external-service.com/sync",
		syncPayload,
		headers,
	)

	if err != nil {
		lf.Append(logger.Any("error", err.Error()))
		logger.Error("Failed to sync data", lf)
		return fmt.Errorf("failed to sync data: %w", err)
	}

	if !resp.IsSuccess() {
		lf.Append(logger.Any("status_code", resp.StatusCode))
		lf.Append(logger.Any("response", resp.String()))
		logger.Error("Sync API returned error", lf)
		return fmt.Errorf("sync API error: status %d", resp.StatusCode)
	}

	// Update sync status in cache
	syncStatusKey := cache.NewCacheKey("sync_status").Build(data.EntityType, data.EntityID)
	j.cache.Set(ctx, syncStatusKey, "synced", 1*time.Hour)

	logger.Info("Data synced successfully", lf)
	return nil
}
