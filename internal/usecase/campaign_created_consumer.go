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

// Example Pub/Sub consumer for creating campaigns from Pub/Sub messages
type campaignCreatedConsumer struct {
	campaignRepo repository.CampaignRepository
}

func NewCampaignCreatedConsumer(campaignRepo repository.CampaignRepository) contract.PubSubConsumer {
	return &campaignCreatedConsumer{
		campaignRepo: campaignRepo,
	}
}

func (c *campaignCreatedConsumer) Consume(data appctx.PubSubData) appctx.PubSubResponse {
	ctx, span := telemetry.StartSpan(data.Ctx, "campaignCreatedConsumer.Consume")
	defer span.End()

	lf := logger.NewFields("CampaignCreatedConsumer").WithTrace(ctx)
	lf.Append(logger.Any("message_id", data.Message.ID))

	logger.Info("Processing campaign created message", lf)

	// Parse message data
	var campaign entity.Campaign
	if err := json.Unmarshal(data.Message.Data, &campaign); err != nil {
		telemetry.SpanError(ctx, err)
		lf.Append(logger.Any("error", err.Error()))
		logger.Error("Failed to parse campaign message data", lf)
		return *appctx.NewPubSubResponse().WithError(err)
	}

	lf.Append(logger.Any("campaign_name", campaign.Name))

	// Create campaign in database
	if err := c.campaignRepo.Create(ctx, &campaign); err != nil {
		telemetry.SpanError(ctx, err)
		lf.Append(logger.Any("error", err.Error()))
		logger.Error("Failed to create campaign from Pub/Sub message", lf)
		return *appctx.NewPubSubResponse().WithError(err)
	}

	lf.Append(logger.Any("campaign_id", campaign.ID))
	logger.Info("Campaign created successfully from Pub/Sub message", lf)

	return *appctx.NewPubSubResponse().WithMessage("Campaign created successfully")
}
