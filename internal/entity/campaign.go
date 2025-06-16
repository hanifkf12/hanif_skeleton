package entity

import (
	"time"
)

type Campaign struct {
	ID             string    `json:"id" db:"id"`
	Name           string    `json:"name" db:"name"`
	TargetDonation float64   `json:"target_donation" db:"target_donation"`
	EndDate        time.Time `json:"end_date" db:"end_date"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`
}

type CreateCampaignRequest struct {
	Name           string    `json:"name" validate:"required"`
	TargetDonation float64   `json:"target_donation" validate:"required,gt=0"`
	EndDate        time.Time `json:"end_date" validate:"required,gt=now"`
}

type UpdateCampaignRequest struct {
	ID             string    `json:"id" validate:"required,uuid"`
	Name           string    `json:"name" validate:"required"`
	TargetDonation float64   `json:"target_donation" validate:"required,gt=0"`
	EndDate        time.Time `json:"end_date" validate:"required,gt=now"`
}