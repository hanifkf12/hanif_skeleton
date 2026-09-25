package usecase

import (
	"github.com/hanifkf12/hanif_skeleton/internal/appctx"
	"github.com/hanifkf12/hanif_skeleton/internal/repository"
	"github.com/hanifkf12/hanif_skeleton/internal/usecase/contract"
	"github.com/hanifkf12/hanif_skeleton/pkg/logger"
	"github.com/hanifkf12/hanif_skeleton/pkg/telemetry"
)

type campaign struct {
	campaignRepo repository.CampaignRepository
}

func (c *campaign) Serve(data appctx.Data) appctx.Response {
	ctx := data.Request.Context()
	ctx, span := telemetry.StartSpan(ctx, "campaign.Serve")
	defer span.End()

	lf := logger.NewFields("GetAllCampaigns").WithTrace(ctx)

	campaigns, err := c.campaignRepo.GetAll(ctx)
	if err != nil {
		telemetry.SpanError(ctx, err)
		lf.Append(logger.Any("error", err.Error()))
		logger.Error("Failed to get all campaigns", lf)
		return *appctx.ResponseFromError(err)
	}

	lf.Append(logger.Any("count", len(campaigns)))
	logger.Info("Successfully retrieved all campaigns", lf)
	return *appctx.NewResponse().WithCode(appctx.StatusOK).WithData(campaigns)
}

func NewCampaign(campaignRepo repository.CampaignRepository) contract.UseCase {
	return &campaign{
		campaignRepo: campaignRepo,
	}
}
