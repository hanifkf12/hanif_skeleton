package campaign

import (
	"context"
	"github.com/google/uuid"
	"github.com/hanifkf12/hanif_skeleton/internal/entity"
	"github.com/hanifkf12/hanif_skeleton/internal/repository"
	"github.com/hanifkf12/hanif_skeleton/pkg/databasex"
	"github.com/hanifkf12/hanif_skeleton/pkg/telemetry"
)

type campaignRepository struct {
	db databasex.Database
}

func (c *campaignRepository) Create(ctx context.Context, campaign *entity.Campaign) error {
	ctx, span := telemetry.StartSpan(ctx, "CampaignRepository.Create")
	defer span.End()

	campaign.ID = uuid.New().String()
	query := `INSERT INTO campaigns (id, name, target_donation, end_date) VALUES (?, ?, ?, ?)`
	_, err := c.db.Exec(ctx, query, campaign.ID, campaign.Name, campaign.TargetDonation, campaign.EndDate)
	return err
}

func (c *campaignRepository) Update(ctx context.Context, campaign *entity.Campaign) error {
	ctx, span := telemetry.StartSpan(ctx, "CampaignRepository.Update")
	defer span.End()

	query := `UPDATE campaigns SET name = ?, target_donation = ?, end_date = ? WHERE id = ?`
	_, err := c.db.Exec(ctx, query, campaign.Name, campaign.TargetDonation, campaign.EndDate, campaign.ID)
	return err
}

func (c *campaignRepository) Delete(ctx context.Context, id string) error {
	ctx, span := telemetry.StartSpan(ctx, "CampaignRepository.Delete")
	defer span.End()

	query := `DELETE FROM campaigns WHERE id = ?`
	_, err := c.db.Exec(ctx, query, id)
	return err
}

func (c *campaignRepository) GetByID(ctx context.Context, id string) (*entity.Campaign, error) {
	ctx, span := telemetry.StartSpan(ctx, "CampaignRepository.GetByID")
	defer span.End()

	query := `SELECT * FROM campaigns WHERE id = ?`
	var campaign entity.Campaign
	err := c.db.Get(ctx, &campaign, query, id)
	if err != nil {
		return nil, err
	}
	return &campaign, nil
}

func (c *campaignRepository) GetAll(ctx context.Context) ([]entity.Campaign, error) {
	ctx, span := telemetry.StartSpan(ctx, "CampaignRepository.GetAll")
	defer span.End()

	query := `SELECT * FROM campaigns`
	var campaigns []entity.Campaign
	err := c.db.Select(ctx, &campaigns, query)
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