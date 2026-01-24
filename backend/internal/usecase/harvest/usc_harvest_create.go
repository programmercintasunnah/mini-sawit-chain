package usecase_harvest

import (
	"context"
	"mini-sawit-chain/backend/internal/domain"
	"time"

	"github.com/google/uuid"
)

func (u *harvestUsecase) Create(
	ctx context.Context,
	farmerID uuid.UUID,
	weightKg float64,
) (*domain.Harvest, error) {

	if weightKg <= 0 {
		return nil, domain.ErrInvalidWeight
	}

	harvest := &domain.Harvest{
		ID:        uuid.New(),
		FarmerID:  farmerID,
		WeightKg:  weightKg,
		Status:    domain.HarvestStatusCreated,
		CreatedAt: time.Now(),
	}

	if err := u.harvestRepo.Create(ctx, harvest); err != nil {
		return nil, err
	}

	return harvest, nil
}
