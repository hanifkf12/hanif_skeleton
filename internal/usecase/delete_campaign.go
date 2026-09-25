package usecase

import (
	"github.com/hanifkf12/hanif_skeleton/internal/appctx"
	"github.com/hanifkf12/hanif_skeleton/internal/repository"
	"github.com/hanifkf12/hanif_skeleton/internal/usecase/contract"
	"github.com/hanifkf12/hanif_skeleton/pkg/apperror"
	"github.com/hanifkf12/hanif_skeleton/pkg/logger"
	"github.com/hanifkf12/hanif_skeleton/pkg/telemetry"
)

type deleteCampaign struct {
	campaignRepo repository.CampaignRepository
}

func (d *deleteCampaign) Serve(data appctx.Data) appctx.Response {
	ctx := data.Request.Context()
	ctx, span := telemetry.StartSpan(ctx, "deleteCampaign.Serve")
	defer span.End()

	lf := logger.NewFields("DeleteCampaign").WithTrace(ctx)

	id := data.Request.Param("id")
	if id == "" {
		lf.Append(logger.Any("error", "Campaign ID is required"))
		logger.Error("Missing campaign ID in request", lf)
		return *appctx.ResponseFromError(apperror.Invalid("Campaign ID is required"))
	}

	lf.Append(logger.Any("campaign_id", id))

	// Check if campaign exists
	_, err := d.campaignRepo.GetByID(ctx, id)
	if err != nil {
		telemetry.SpanError(ctx, err)
		lf.Append(logger.Any("error", err.Error()))
		logger.Error("Campaign not found", lf)
		return *appctx.ResponseFromError(apperror.NotFound("Campaign not found"))
	}

	if err := d.campaignRepo.Delete(ctx, id); err != nil {
		telemetry.SpanError(ctx, err)
		lf.Append(logger.Any("error", err.Error()))
		logger.Error("Failed to delete campaign", lf)
		return *appctx.ResponseFromError(err)
	}

	logger.Info("Campaign deleted successfully", lf)
	return *appctx.NewResponse().WithCode(appctx.StatusOK).WithMessage("Campaign deleted successfully")
}

func NewDeleteCampaign(campaignRepo repository.CampaignRepository) contract.UseCase {
	return &deleteCampaign{
		campaignRepo: campaignRepo,
	}
}
