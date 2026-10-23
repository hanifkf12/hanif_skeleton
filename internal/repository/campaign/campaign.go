package campaign

import (
	"context"

	"github.com/google/uuid"
	"github.com/hanifkf12/hanif_skeleton/internal/entity"
	"github.com/hanifkf12/hanif_skeleton/internal/repository"
	"github.com/hanifkf12/hanif_skeleton/pkg/apperror"
	"github.com/hanifkf12/hanif_skeleton/pkg/databasex"
	"github.com/hanifkf12/hanif_skeleton/pkg/sqlbuilder"
	"github.com/hanifkf12/hanif_skeleton/pkg/telemetry"
)

type campaignRepository struct {
	db databasex.Database
}

func (c *campaignRepository) Create(ctx context.Context, campaign *entity.Campaign) error {
	ctx, span := telemetry.StartSpan(ctx, "CampaignRepository.Create")
	defer span.End()

	campaign.ID = uuid.New().String()

	// Using SQL Builder for cleaner code
	model := sqlbuilder.NewModel(c.db, campaign)
	_, err := model.
		Table("campaigns").
		InsertWithFields(ctx, campaign, "id", "name", "target_donation", "end_date")

	return repository.MapSQLError(err, "Failed to create campaign")
}

func (c *campaignRepository) Update(ctx context.Context, campaign *entity.Campaign) error {
	ctx, span := telemetry.StartSpan(ctx, "CampaignRepository.Update")
	defer span.End()

	// Using SQL Builder for cleaner code
	model := sqlbuilder.NewModel(c.db, campaign)
	result, err := model.
		Table("campaigns").
		Where("id = ?", campaign.ID).
		Update(ctx, campaign)
	if err != nil {
		return repository.MapSQLError(err, "Campaign not found")
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return apperror.Internal(err)
	}
	if affected == 0 {
		return apperror.NotFound("Campaign not found")
	}

	return nil
}

func (c *campaignRepository) Delete(ctx context.Context, id string) error {
	ctx, span := telemetry.StartSpan(ctx, "CampaignRepository.Delete")
	defer span.End()

	// Using SQL Builder for cleaner code
	model := sqlbuilder.NewModel(c.db, nil)
	result, err := model.
		Table("campaigns").
		Where("id = ?", id).
		Delete(ctx)
	if err != nil {
		return repository.MapSQLError(err, "Campaign not found")
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return apperror.Internal(err)
	}
	if affected == 0 {
		return apperror.NotFound("Campaign not found")
	}

	return nil
}

func (c *campaignRepository) GetByID(ctx context.Context, id string) (*entity.Campaign, error) {
	ctx, span := telemetry.StartSpan(ctx, "CampaignRepository.GetByID")
	defer span.End()

	var campaign entity.Campaign

	// Using SQL Builder for cleaner code
	model := sqlbuilder.NewModel(c.db, &campaign)
	err := model.
		Table("campaigns").
		Select("id", "name", "target_donation", "end_date", "created_at", "updated_at").
		Where("id = ?", id).
		First(ctx, &campaign)

	if err != nil {
		return nil, repository.MapSQLError(err, "Campaign not found")
	}

	return &campaign, nil
}

func (c *campaignRepository) GetAll(ctx context.Context) ([]entity.Campaign, error) {
	ctx, span := telemetry.StartSpan(ctx, "CampaignRepository.GetAll")
	defer span.End()

	var campaigns []entity.Campaign

	model := sqlbuilder.NewModel(c.db, &entity.Campaign{})
	err := model.
		Table("campaigns").
		Select("id", "name", "target_donation", "end_date", "created_at", "updated_at").
		OrderBy("created_at", "DESC").
		GetAll(ctx, &campaigns)

	if err != nil {
		return nil, repository.MapSQLError(err, "Campaigns not found")
	}

	return campaigns, nil
}

func NewCampaignRepository(db databasex.Database) repository.CampaignRepository {
	return &campaignRepository{
		db: db,
	}
}
