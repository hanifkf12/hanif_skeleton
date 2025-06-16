package usecase

import (
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/hanifkf12/hanif_skeleton/internal/appctx"
	"github.com/hanifkf12/hanif_skeleton/internal/entity"
	"github.com/hanifkf12/hanif_skeleton/internal/repository"
	"github.com/hanifkf12/hanif_skeleton/internal/usecase/contract"
	"github.com/hanifkf12/hanif_skeleton/pkg/logger"
	"github.com/hanifkf12/hanif_skeleton/pkg/telemetry"
)

type createCampaign struct {
	campaignRepo repository.CampaignRepository
	validator    *validator.Validate
}

func (c *createCampaign) Serve(data appctx.Data) appctx.Response {
	ctx := data.FiberCtx.UserContext()
	ctx, span := telemetry.StartSpan(ctx, "createCampaign.Serve")
	defer span.End()

	lf := logger.NewFields("CreateCampaign").WithTrace(ctx)

	req := new(entity.CreateCampaignRequest)
	if err := data.FiberCtx.BodyParser(req); err != nil {
		lf.Append(logger.Any("error", err.Error()))
		logger.Error("Failed to parse create campaign request", lf)
		return *appctx.NewResponse().WithCode(fiber.StatusBadRequest).WithErrors(err.Error())
	}

	if err := c.validator.Struct(req); err != nil {
		lf.Append(logger.Any("error", err.Error()))
		lf.Append(logger.Any("request", req))
		logger.Error("Invalid create campaign request", lf)
		return *appctx.NewResponse().WithCode(fiber.StatusBadRequest).WithErrors(err.Error())
	}

	campaign := &entity.Campaign{
		Name:           req.Name,
		TargetDonation: req.TargetDonation,
		EndDate:        req.EndDate,
	}

	lf.Append(logger.Any("campaign", campaign))

	if err := c.campaignRepo.Create(ctx, campaign); err != nil {
		lf.Append(logger.Any("error", err.Error()))
		logger.Error("Failed to create campaign", lf)
		return *appctx.NewResponse().WithCode(fiber.StatusInternalServerError).WithErrors(err.Error())
	}

	logger.Info("Campaign created successfully", lf)
	return *appctx.NewResponse().WithCode(fiber.StatusCreated).WithData(campaign)
}

func NewCreateCampaign(campaignRepo repository.CampaignRepository) contract.UseCase {
	return &createCampaign{
		campaignRepo: campaignRepo,
		validator:    validator.New(),
	}
}