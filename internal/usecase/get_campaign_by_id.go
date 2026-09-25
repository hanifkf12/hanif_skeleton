package usecase

import (
	"github.com/hanifkf12/hanif_skeleton/internal/appctx"
	"github.com/hanifkf12/hanif_skeleton/internal/entity"
	"github.com/hanifkf12/hanif_skeleton/internal/repository"
	"github.com/hanifkf12/hanif_skeleton/internal/usecase/contract"
	"github.com/hanifkf12/hanif_skeleton/pkg/apperror"
	"github.com/hanifkf12/hanif_skeleton/pkg/logger"
	"github.com/hanifkf12/hanif_skeleton/pkg/telemetry"
)

type getCampaignByID struct {
	campaignRepo repository.CampaignRepository
}

func NewGetCampaignByID(campaignRepo repository.CampaignRepository) contract.UseCase {
	return &getCampaignByID{campaignRepo: campaignRepo}
}

func (c *getCampaignByID) Serve(data appctx.Data) appctx.Response {
	ctx := data.Request.Context()
	ctx, span := telemetry.StartSpan(ctx, "getCampaignByID.Serve")
	defer span.End()

	lf := logger.NewFields("GetCampaignByID").WithTrace(ctx)

	// Parse :id from URL param
	id := data.Request.Param("id")
	if id == "" {
		logger.Error("Campaign ID is empty", lf)
		return *appctx.ResponseFromError(apperror.Invalid("Campaign ID is required"))
	}

	lf.Append(logger.Any("campaign_id", id))

	campaign, err := c.campaignRepo.GetByID(ctx, id)
	if err != nil {
		telemetry.SpanError(ctx, err)
		lf.Append(logger.Any("error", err.Error()))
		logger.Error("Failed to get campaign", lf)
		return *appctx.ResponseFromError(apperror.NotFound("Campaign not found"))
	}

	resp := entity.GetCampaignByIDResponse{
		ID:             campaign.ID,
		Name:           campaign.Name,
		TargetDonation: campaign.TargetDonation,
		EndDate:        campaign.EndDate,
		CreatedAt:      campaign.CreatedAt,
		UpdatedAt:      campaign.UpdatedAt,
	}

	logger.Info("Successfully retrieved campaign", lf)
	return *appctx.NewResponse().WithCode(appctx.StatusOK).WithData(resp)
}
