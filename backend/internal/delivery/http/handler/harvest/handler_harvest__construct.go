package handler_harvest

import (
	usecase_harvest "mini-sawit-chain/backend/internal/usecase/harvest"
	"time"

	"github.com/google/uuid"
)

type Handler struct {
	harvestUC usecase_harvest.HarvestUsecase
}

func NewHandler(uc usecase_harvest.HarvestUsecase) *Handler {
	return &Handler{
		harvestUC: uc,
	}
}

type CreateHarvestRequest struct {
	FarmerID uuid.UUID `json:"farmer_id" validate:"required"`
	WeightKg float64   `json:"weight_kg" validate:"required,gt=0"`
}

type HarvestResponse struct {
	ID        uuid.UUID `json:"id"`
	FarmerID  uuid.UUID `json:"farmer_id"`
	WeightKg  float64   `json:"weight_kg"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}
