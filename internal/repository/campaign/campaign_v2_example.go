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

// Example: Campaign Repository REFACTORED using SQL Builder
// This is the improved version of your campaign repository

type campaignRepositoryV2 struct {
	db databasex.Database
}

func (c *campaignRepositoryV2) Create(ctx context.Context, campaign *entity.Campaign) error {
	ctx, span := telemetry.StartSpan(ctx, "CampaignRepository.Create")
	defer span.End()

	campaign.ID = uuid.New().String()

	// Using SQL Builder
	model := sqlbuilder.NewModel(c.db, campaign)
	_, err := model.
		Table("campaigns").
		Insert(ctx, campaign)

	return err
}

func (c *campaignRepositoryV2) Update(ctx context.Context, campaign *entity.Campaign) error {
	ctx, span := telemetry.StartSpan(ctx, "CampaignRepository.Update")
	defer span.End()

	// Using SQL Builder
	model := sqlbuilder.NewModel(c.db, campaign)
	_, err := model.
		Table("campaigns").
		Where("id = ?", campaign.ID).
		Update(ctx, campaign)

	return err
}

func (c *campaignRepositoryV2) Delete(ctx context.Context, id string) error {
	ctx, span := telemetry.StartSpan(ctx, "CampaignRepository.Delete")
	defer span.End()

	// Using SQL Builder
	model := sqlbuilder.NewModel(c.db, nil)
	_, err := model.
		Table("campaigns").
		Where("id = ?", id).
		Delete(ctx)

	return err
}

func (c *campaignRepositoryV2) GetByID(ctx context.Context, id string) (*entity.Campaign, error) {
	ctx, span := telemetry.StartSpan(ctx, "CampaignRepository.GetByID")
	defer span.End()

	var campaign entity.Campaign

	// Using SQL Builder
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

func (c *campaignRepositoryV2) GetAll(ctx context.Context) ([]entity.Campaign, error) {
	ctx, span := telemetry.StartSpan(ctx, "CampaignRepository.GetAll")
	defer span.End()

	var campaigns []entity.Campaign

	// Using SQL Builder
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

// Advanced examples using SQL Builder features

// GetActiveCampaigns - Get all active campaigns
func (c *campaignRepositoryV2) GetActiveCampaigns(ctx context.Context) ([]entity.Campaign, error) {
	ctx, span := telemetry.StartSpan(ctx, "CampaignRepository.GetActiveCampaigns")
	defer span.End()

	var campaigns []entity.Campaign

	model := sqlbuilder.NewModel(c.db, &entity.Campaign{})
	err := model.
		Table("campaigns").
		Where("end_date > ?", "NOW()"). // Still active
		OrderBy("target_donation", "DESC").
		GetAll(ctx, &campaigns)

	return campaigns, err
}

// SearchCampaigns - Search campaigns with dynamic filters
func (c *campaignRepositoryV2) SearchCampaigns(ctx context.Context, name string, minDonation, maxDonation float64) ([]entity.Campaign, error) {
	ctx, span := telemetry.StartSpan(ctx, "CampaignRepository.SearchCampaigns")
	defer span.End()

	var campaigns []entity.Campaign

	// Using Conditional Builder for dynamic WHERE
	cb := sqlbuilder.NewConditionalBuilder()
	cb.AddIf(name != "", "name LIKE ?", "%"+name+"%")
	cb.AddIf(minDonation > 0, "target_donation >= ?", minDonation)
	cb.AddIf(maxDonation > 0, "target_donation <= ?", maxDonation)

	model := sqlbuilder.NewModel(c.db, &entity.Campaign{})
	model.Table("campaigns")

	if !cb.IsEmpty() {
		condition, args := cb.Build()
		model.Where(condition, args...)
	}

	err := model.
		OrderBy("created_at", "DESC").
		GetAll(ctx, &campaigns)

	return campaigns, err
}

// GetCampaignsPaginated - Get campaigns with pagination
func (c *campaignRepositoryV2) GetCampaignsPaginated(ctx context.Context, page, perPage int) (*sqlbuilder.PaginationResult, error) {
	ctx, span := telemetry.StartSpan(ctx, "CampaignRepository.GetCampaignsPaginated")
	defer span.End()

	var campaigns []entity.Campaign

	model := sqlbuilder.NewModel(c.db, &entity.Campaign{})
	result, err := model.
		Table("campaigns").
		OrderBy("created_at", "DESC").
		GetWithPagination(ctx, &campaigns, page, perPage)

	return result, err
}

// GetCampaignsByIDs - Get multiple campaigns by IDs
func (c *campaignRepositoryV2) GetCampaignsByIDs(ctx context.Context, ids []string) ([]entity.Campaign, error) {
	ctx, span := telemetry.StartSpan(ctx, "CampaignRepository.GetCampaignsByIDs")
	defer span.End()

	var campaigns []entity.Campaign

	// Convert to []interface{} for WhereIn
	idsInterface := make([]interface{}, len(ids))
	for i, id := range ids {
		idsInterface[i] = id
	}

	model := sqlbuilder.NewModel(c.db, &entity.Campaign{})
	err := model.
		Table("campaigns").
		WhereIn("id", idsInterface).
		GetAll(ctx, &campaigns)

	return campaigns, err
}

// CountActiveCampaigns - Count active campaigns
func (c *campaignRepositoryV2) CountActiveCampaigns(ctx context.Context) (int64, error) {
	ctx, span := telemetry.StartSpan(ctx, "CampaignRepository.CountActiveCampaigns")
	defer span.End()

	model := sqlbuilder.NewModel(c.db, nil)
	count, err := model.
		Table("campaigns").
		Where("end_date > NOW()").
		Count(ctx)

	return count, err
}

// UpdatePartial - Update specific fields only
func (c *campaignRepositoryV2) UpdatePartial(ctx context.Context, id string, name string, targetDonation float64) error {
	ctx, span := telemetry.StartSpan(ctx, "CampaignRepository.UpdatePartial")
	defer span.End()

	// Create partial update
	campaign := entity.Campaign{
		ID:             id,
		Name:           name,
		TargetDonation: targetDonation,
	}

	model := sqlbuilder.NewModel(c.db, &campaign)
	_, err := model.
		Table("campaigns").
		Where("id = ?", id).
		UpdateWithFields(ctx, &campaign, "name", "target_donation")

	return err
}

// BulkCreateCampaigns - Create multiple campaigns at once
func (c *campaignRepositoryV2) BulkCreateCampaigns(ctx context.Context, campaigns []entity.Campaign) error {
	ctx, span := telemetry.StartSpan(ctx, "CampaignRepository.BulkCreateCampaigns")
	defer span.End()

	bulkInsert := sqlbuilder.NewBulkInsertBuilder("campaigns")

	for _, campaign := range campaigns {
		campaign.ID = uuid.New().String()
		bulkInsert.AddFromStruct(&campaign)
	}

	query, args := bulkInsert.Build()
	_, err := c.db.Exec(ctx, query, args...)

	return err
}

func NewCampaignRepositoryV2(db databasex.Database) repository.CampaignRepository {
	return &campaignRepositoryV2{
		db: db,
	}
}
