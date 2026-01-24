package usecase_harvest

import (
	"context"
	"mini-sawit-chain/backend/internal/domain"
	repository_harvest "mini-sawit-chain/backend/internal/repository/harvest"

	"github.com/google/uuid"
)

type HarvestUsecase interface {
	Create(ctx context.Context, farmerID uuid.UUID, weightKg float64) (*domain.Harvest, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Harvest, error)
	GetAll(ctx context.Context) ([]domain.Harvest, error)
	// Update(ctx context.Context, id uuid.UUID, weightKg float64) error
	// Delete(ctx context.Context, id uuid.UUID) error
	Weigh(ctx context.Context, harvestID uuid.UUID, weightKg float64) error
	Pay(ctx context.Context, harvestID uuid.UUID) error
}

type harvestUsecase struct {
	harvestRepo repository_harvest.HarvestRepository
}

func NewUsecase(
	harvestRepo repository_harvest.HarvestRepository,
) HarvestUsecase {
	return &harvestUsecase{
		harvestRepo: harvestRepo,
	}
}
