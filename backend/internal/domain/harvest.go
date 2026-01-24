package domain

import (
	"time"

	"github.com/google/uuid"
)

type HarvestStatus string

const (
	HarvestStatusCreated HarvestStatus = "CREATED"
	HarvestStatusWeighed HarvestStatus = "WEIGHED"
	HarvestStatusPaid    HarvestStatus = "PAID"
)

type Harvest struct {
	ID        uuid.UUID
	FarmerID  uuid.UUID
	WeightKg  float64
	Status    HarvestStatus
	CreatedAt time.Time
	UpdatedAt time.Time
}
