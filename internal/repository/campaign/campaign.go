package campaign

import (
	"context"

	"github.com/google/uuid"
	"github.com/hanifkf12/hanif_skeleton/internal/entity"
	"github.com/hanifkf12/hanif_skeleton/internal/repository"
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
		Insert(ctx, campaign)

	return err
}

func (c *campaignRepository) Update(ctx context.Context, campaign *entity.Campaign) error {
	ctx, span := telemetry.StartSpan(ctx, "CampaignRepository.Update")
	defer span.End()

	// Using SQL Builder for cleaner code
	model := sqlbuilder.NewModel(c.db, campaign)
	_, err := model.
		Table("campaigns").
		Where("id = ?", campaign.ID).
		Update(ctx, campaign)

	return err
}

func (c *campaignRepository) Delete(ctx context.Context, id string) error {
	ctx, span := telemetry.StartSpan(ctx, "CampaignRepository.Delete")
	defer span.End()

	// Using SQL Builder for cleaner code
	model := sqlbuilder.NewModel(c.db, nil)
	_, err := model.
		Table("campaigns").
		Where("id = ?", id).
		Delete(ctx)

	return err
}

func (c *campaignRepository) GetByID(ctx context.Context, id string) (*entity.Campaign, error) {
	ctx, span := telemetry.StartSpan(ctx, "CampaignRepository.GetByID")
	defer span.End()

	var campaign entity.Campaign

	// Using SQL Builder for cleaner code
	model := sqlbuilder.NewModel(c.db, &campaign)
	err := model.
		Table("campaigns").
		Where("id = ?", id).
		First(ctx, &campaign)

	if err != nil {
		return nil, err
	}

	return &campaign, nil
}

func (c *campaignRepository) GetAll(ctx context.Context) ([]entity.Campaign, error) {
	ctx, span := telemetry.StartSpan(ctx, "CampaignRepository.GetAll")
	defer span.End()

	var campaigns []entity.Campaign

	// Using SQL Builder for cleaner code
	model := sqlbuilder.NewModel(c.db, &entity.Campaign{})
	err := model.
		Table("campaigns").
		OrderBy("created_at", "DESC").
		GetAll(ctx, &campaigns)

	if err != nil {
		return nil, err
	}

	return campaigns, nil
}

func NewCampaignRepository(db databasex.Database) repository.CampaignRepository {
	return &campaignRepository{
		db: db,
	}
}
