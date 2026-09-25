package usecase

import (
	"github.com/go-playground/validator/v10"
	"github.com/hanifkf12/hanif_skeleton/internal/appctx"
	"github.com/hanifkf12/hanif_skeleton/internal/entity"
	"github.com/hanifkf12/hanif_skeleton/internal/repository"
	"github.com/hanifkf12/hanif_skeleton/internal/usecase/contract"
	"github.com/hanifkf12/hanif_skeleton/pkg/apperror"
	"github.com/hanifkf12/hanif_skeleton/pkg/logger"
	"github.com/hanifkf12/hanif_skeleton/pkg/telemetry"
)

type updateCampaign struct {
	campaignRepo repository.CampaignRepository
	validator    *validator.Validate
}

func (u *updateCampaign) Serve(data appctx.Data) appctx.Response {
	ctx := data.Request.Context()
	ctx, span := telemetry.StartSpan(ctx, "updateCampaign.Serve")
	defer span.End()

	lf := logger.NewFields("UpdateCampaign").WithTrace(ctx)

	req := new(entity.UpdateCampaignRequest)
	if err := data.Request.Body(req); err != nil {
		telemetry.SpanError(ctx, err)
		lf.Append(logger.Any("error", err.Error()))
		logger.Error("Failed to parse update campaign request", lf)
		return *appctx.ResponseFromError(apperror.Invalid("Invalid request body"))
	}

	if err := u.validator.Struct(req); err != nil {
		telemetry.SpanError(ctx, err)
		lf.Append(logger.Any("error", err.Error()))
		logger.Error("Invalid update campaign request", lf)
		return *appctx.ResponseFromError(apperror.Wrap(err, apperror.KindInvalid, "Invalid request body"))
	}

	lf.Append(logger.Any("campaign_id", req.ID))

	// Check if campaign exists
	existing, err := u.campaignRepo.GetByID(ctx, req.ID)
	if err != nil {
		telemetry.SpanError(ctx, err)
		lf.Append(logger.Any("error", err.Error()))
		logger.Error("Campaign not found", lf)
		return *appctx.ResponseFromError(apperror.NotFound("Campaign not found"))
	}

	// Update campaign fields
	existing.Name = req.Name
	existing.TargetDonation = req.TargetDonation
	existing.EndDate = req.EndDate

	lf.Append(logger.Any("updated_campaign", existing))

	if err := u.campaignRepo.Update(ctx, existing); err != nil {
		telemetry.SpanError(ctx, err)
		lf.Append(logger.Any("error", err.Error()))
		logger.Error("Failed to update campaign", lf)
		return *appctx.ResponseFromError(err)
	}

	logger.Info("Campaign updated successfully", lf)
	return *appctx.NewResponse().WithCode(appctx.StatusOK).WithData(existing)
}

func NewUpdateCampaign(campaignRepo repository.CampaignRepository) contract.UseCase {
	return &updateCampaign{
		campaignRepo: campaignRepo,
		validator:    validator.New(),
	}
}
