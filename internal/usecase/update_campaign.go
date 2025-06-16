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

type updateCampaign struct {
	campaignRepo repository.CampaignRepository
	validator    *validator.Validate
}

func (u *updateCampaign) Serve(data appctx.Data) appctx.Response {
	ctx := data.FiberCtx.UserContext()
	ctx, span := telemetry.StartSpan(ctx, "updateCampaign.Serve")
	defer span.End()

	lf := logger.NewFields("UpdateCampaign").WithTrace(ctx)

	req := new(entity.UpdateCampaignRequest)
	if err := data.FiberCtx.BodyParser(req); err != nil {
		lf.Append(logger.Any("error", err.Error()))
		logger.Error("Failed to parse update campaign request", lf)
		return *appctx.NewResponse().WithCode(fiber.StatusBadRequest).WithErrors(err.Error())
	}

	if err := u.validator.Struct(req); err != nil {
		lf.Append(logger.Any("error", err.Error()))
		lf.Append(logger.Any("request", req))
		logger.Error("Invalid update campaign request", lf)
		return *appctx.NewResponse().WithCode(fiber.StatusBadRequest).WithErrors(err.Error())
	}

	lf.Append(logger.Any("campaign_id", req.ID))

	// Check if campaign exists
	existing, err := u.campaignRepo.GetByID(ctx, req.ID)
	if err != nil {
		lf.Append(logger.Any("error", err.Error()))
		logger.Error("Campaign not found", lf)
		return *appctx.NewResponse().WithCode(fiber.StatusNotFound).WithErrors("Campaign not found")
	}

	// Update campaign fields
	existing.Name = req.Name
	existing.TargetDonation = req.TargetDonation
	existing.EndDate = req.EndDate

	lf.Append(logger.Any("updated_campaign", existing))

	if err := u.campaignRepo.Update(ctx, existing); err != nil {
		lf.Append(logger.Any("error", err.Error()))
		logger.Error("Failed to update campaign", lf)
		return *appctx.NewResponse().WithCode(fiber.StatusInternalServerError).WithErrors(err.Error())
	}

	logger.Info("Campaign updated successfully", lf)
	return *appctx.NewResponse().WithCode(fiber.StatusOK).WithData(existing)
}

func NewUpdateCampaign(campaignRepo repository.CampaignRepository) contract.UseCase {
	return &updateCampaign{
		campaignRepo: campaignRepo,
		validator:    validator.New(),
	}
}