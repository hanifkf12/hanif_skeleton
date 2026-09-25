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

type createCampaign struct {
	campaignRepo repository.CampaignRepository
	validator    *validator.Validate
}

func (c *createCampaign) Serve(data appctx.Data) appctx.Response {
	ctx := data.Request.Context()
	ctx, span := telemetry.StartSpan(ctx, "createCampaign.Serve")
	defer span.End()

	lf := logger.NewFields("CreateCampaign").WithTrace(ctx)

	req := new(entity.CreateCampaignRequest)
	if err := data.Request.Body(req); err != nil {
		telemetry.SpanError(ctx, err)
		lf.Append(logger.Any("error", err.Error()))
		logger.Error("Failed to parse create campaign request", lf)
		return *appctx.ResponseFromError(apperror.Invalid("Invalid request body"))
	}

	if err := c.validator.Struct(req); err != nil {
		telemetry.SpanError(ctx, err)
		lf.Append(logger.Any("error", err.Error()))
		logger.Error("Invalid create campaign request", lf)
		return *appctx.ResponseFromError(apperror.Wrap(err, apperror.KindInvalid, "Invalid request body"))
	}

	campaign := &entity.Campaign{
		Name:           req.Name,
		TargetDonation: req.TargetDonation,
		EndDate:        req.EndDate,
	}

	lf.Append(logger.Any("campaign", campaign))

	if err := c.campaignRepo.Create(ctx, campaign); err != nil {
		telemetry.SpanError(ctx, err)
		lf.Append(logger.Any("error", err.Error()))
		logger.Error("Failed to create campaign", lf)
		return *appctx.ResponseFromError(err)
	}

	logger.Info("Campaign created successfully", lf)
	return *appctx.NewResponse().WithCode(appctx.StatusCreated).WithData(campaign)
}

func NewCreateCampaign(campaignRepo repository.CampaignRepository) contract.UseCase {
	return &createCampaign{
		campaignRepo: campaignRepo,
		validator:    validator.New(),
	}
}
