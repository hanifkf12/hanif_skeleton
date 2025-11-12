package sqlbuilder

import (
	"context"
	"time"

	"github.com/hanifkf12/hanif_skeleton/pkg/databasex"
)

// Example: Campaign Repository using SQL Builder

type Campaign struct {
	ID             string    `json:"id" db:"id"`
	Name           string    `json:"name" db:"name"`
	TargetDonation float64   `json:"target_donation" db:"target_donation"`
	EndDate        time.Time `json:"end_date" db:"end_date"`
	Status         string    `json:"status" db:"status"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`
}

func (c Campaign) TableName() string {
	return "campaigns"
}

type CampaignFilter struct {
	Status      string
	MinDonation float64
	MaxDonation float64
	StartDate   time.Time
	EndDate     time.Time
	SearchName  string
	Page        int
	PerPage     int
}

// Example Repository Implementation

type ExampleCampaignRepository struct {
	db databasex.Database
}

// GetByID - Find campaign by ID
func (r *ExampleCampaignRepository) GetByID(ctx context.Context, id string) (*Campaign, error) {
	var campaign Campaign

	model := NewModel(r.db, &campaign)
	err := model.
		Table("campaigns").
		Where("id = ?", id).
		First(ctx, &campaign)

	if err != nil {
		return nil, err
	}

	return &campaign, nil
}

// GetAll - Get all campaigns with optional filtering
func (r *ExampleCampaignRepository) GetAll(ctx context.Context, filter CampaignFilter) ([]Campaign, error) {
	var campaigns []Campaign

	model := NewModel(r.db, &Campaign{})
	model.Table("campaigns")

	// Build dynamic WHERE conditions
	cb := NewConditionalBuilder()

	cb.AddIf(filter.Status != "", "status = ?", filter.Status)
	cb.AddIf(filter.MinDonation > 0, "target_donation >= ?", filter.MinDonation)
	cb.AddIf(filter.MaxDonation > 0, "target_donation <= ?", filter.MaxDonation)
	cb.AddIf(!filter.StartDate.IsZero(), "end_date >= ?", filter.StartDate)
	cb.AddIf(!filter.EndDate.IsZero(), "end_date <= ?", filter.EndDate)
	cb.AddIf(filter.SearchName != "", "name LIKE ?", "%"+filter.SearchName+"%")

	if !cb.IsEmpty() {
		condition, args := cb.Build()
		model.Where(condition, args...)
	}

	err := model.
		OrderBy("created_at", "DESC").
		GetAll(ctx, &campaigns)

	return campaigns, err
}

// GetWithPagination - Get campaigns with pagination
func (r *ExampleCampaignRepository) GetWithPagination(ctx context.Context, filter CampaignFilter) (*PaginationResult, error) {
	var campaigns []Campaign

	model := NewModel(r.db, &Campaign{})
	model.Table("campaigns")

	// Apply filters
	if filter.Status != "" {
		model.Where("status = ?", filter.Status)
	}

	if filter.SearchName != "" {
		model.Where("name LIKE ?", "%"+filter.SearchName+"%")
	}

	// Default pagination values
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PerPage < 1 {
		filter.PerPage = 10
	}

	return model.
		OrderBy("created_at", "DESC").
		GetWithPagination(ctx, &campaigns, filter.Page, filter.PerPage)
}

// Create - Create new campaign
func (r *ExampleCampaignRepository) Create(ctx context.Context, campaign *Campaign) error {
	model := NewModel(r.db, campaign)

	// Auto-generate ID if needed
	// campaign.ID = uuid.New().String()

	_, err := model.
		Table("campaigns").
		Insert(ctx, campaign)

	return err
}

// Update - Update existing campaign
func (r *ExampleCampaignRepository) Update(ctx context.Context, campaign *Campaign) error {
	model := NewModel(r.db, campaign)

	_, err := model.
		Table("campaigns").
		Where("id = ?", campaign.ID).
		Update(ctx, campaign)

	return err
}

// UpdatePartial - Update specific fields only
func (r *ExampleCampaignRepository) UpdatePartial(ctx context.Context, id string, updates map[string]interface{}) error {
	qb := NewQueryBuilder()
	query, args := qb.
		Table("campaigns").
		Update(updates).
		Where("id = ?", id).
		Build()

	_, err := r.db.Exec(ctx, query, args...)
	return err
}

// Delete - Delete campaign
func (r *ExampleCampaignRepository) Delete(ctx context.Context, id string) error {
	model := NewModel(r.db, nil)

	_, err := model.
		Table("campaigns").
		Where("id = ?", id).
		Delete(ctx)

	return err
}

// SoftDelete - Soft delete by updating deleted_at
func (r *ExampleCampaignRepository) SoftDelete(ctx context.Context, id string) error {
	qb := NewQueryBuilder()
	query, args := qb.
		Table("campaigns").
		Update(map[string]interface{}{
			"deleted_at": time.Now(),
		}).
		Where("id = ?", id).
		Build()

	_, err := r.db.Exec(ctx, query, args...)
	return err
}

// GetActiveCount - Count active campaigns
func (r *ExampleCampaignRepository) GetActiveCount(ctx context.Context) (int64, error) {
	model := NewModel(r.db, nil)
	return model.
		Table("campaigns").
		Where("status = ?", "active").
		WhereNull("deleted_at").
		Count(ctx)
}

// GetByIDs - Get campaigns by multiple IDs
func (r *ExampleCampaignRepository) GetByIDs(ctx context.Context, ids []string) ([]Campaign, error) {
	var campaigns []Campaign

	// Convert to []interface{} for WhereIn
	idsInterface := make([]interface{}, len(ids))
	for i, id := range ids {
		idsInterface[i] = id
	}

	model := NewModel(r.db, &Campaign{})
	err := model.
		Table("campaigns").
		WhereIn("id", idsInterface).
		GetAll(ctx, &campaigns)

	return campaigns, err
}

// CampaignStats represents campaign statistics
type CampaignStats struct {
	Status string  `db:"status"`
	Count  int     `db:"count"`
	Total  float64 `db:"total"`
}

// GetStats - Get campaign statistics
func (r *ExampleCampaignRepository) GetStats(ctx context.Context) ([]CampaignStats, error) {
	var stats []CampaignStats

	model := NewModel(r.db, nil)
	err := model.
		Table("campaigns").
		Select("status", "COUNT(*) as count", "SUM(target_donation) as total").
		GroupBy("status").
		GetAll(ctx, &stats)

	return stats, err
}

// GetExpiringSoon - Get campaigns expiring soon
func (r *ExampleCampaignRepository) GetExpiringSoon(ctx context.Context, days int) ([]Campaign, error) {
	var campaigns []Campaign

	endDate := time.Now().AddDate(0, 0, days)

	model := NewModel(r.db, &Campaign{})
	err := model.
		Table("campaigns").
		Where("status = ?", "active").
		WhereBetween("end_date", time.Now(), endDate).
		OrderBy("end_date", "ASC").
		GetAll(ctx, &campaigns)

	return campaigns, err
}

// BulkCreate - Create multiple campaigns at once
func (r *ExampleCampaignRepository) BulkCreate(ctx context.Context, campaigns []Campaign) error {
	bulkInsert := NewBulkInsertBuilder("campaigns")

	for _, campaign := range campaigns {
		bulkInsert.AddFromStruct(&campaign)
	}

	query, args := bulkInsert.Build()
	_, err := r.db.Exec(ctx, query, args...)

	return err
}

// CampaignWithCreator represents campaign with creator information
type CampaignWithCreator struct {
	Campaign
	CreatorName string `db:"creator_name"`
}

// GetWithCreator - Example with JOIN (assuming there's a users table)
func (r *ExampleCampaignRepository) GetWithCreator(ctx context.Context) ([]CampaignWithCreator, error) {
	var results []CampaignWithCreator

	model := NewModel(r.db, nil)
	err := model.
		Table("campaigns").
		Select("campaigns.*, users.name as creator_name").
		Join("users", "campaigns.creator_id = users.id").
		Where("campaigns.status = ?", "active").
		GetAll(ctx, &results)

	return results, err
}

// SearchCampaigns - Advanced search with multiple conditions
func (r *ExampleCampaignRepository) SearchCampaigns(ctx context.Context, searchTerm string, minDonation float64, statuses []string) ([]Campaign, error) {
	var campaigns []Campaign

	model := NewModel(r.db, &Campaign{})
	model.Table("campaigns")

	// Search in name
	if searchTerm != "" {
		model.Where("name LIKE ?", "%"+searchTerm+"%")
	}

	// Minimum donation filter
	if minDonation > 0 {
		model.Where("target_donation >= ?", minDonation)
	}

	// Status filter
	if len(statuses) > 0 {
		statusesInterface := make([]interface{}, len(statuses))
		for i, s := range statuses {
			statusesInterface[i] = s
		}
		model.WhereIn("status", statusesInterface)
	}

	err := model.
		OrderBy("created_at", "DESC").
		GetAll(ctx, &campaigns)

	return campaigns, err
}

// UpdateStatus - Update campaign status with conditions
func (r *ExampleCampaignRepository) UpdateStatus(ctx context.Context, status string, condition string, args ...interface{}) error {
	qb := NewQueryBuilder()
	query, queryArgs := qb.
		Table("campaigns").
		Update(map[string]interface{}{
			"status":     status,
			"updated_at": time.Now(),
		}).
		Where(condition, args...).
		Build()

	_, err := r.db.Exec(ctx, query, queryArgs...)
	return err
}

// UpsertCampaign - Insert or update campaign (MySQL)
func (r *ExampleCampaignRepository) UpsertCampaign(ctx context.Context, campaign *Campaign) error {
	upsert := NewUpsertBuilder("campaigns").
		Insert(StructToMap(campaign, false)).
		Update(StructToMapExclude(campaign, "id", "created_at"))

	query, args := upsert.Build()
	_, err := r.db.Exec(ctx, query, args...)

	return err
}
