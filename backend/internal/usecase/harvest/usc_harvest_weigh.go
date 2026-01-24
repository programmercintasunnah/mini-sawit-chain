package usecase_harvest

import (
	"context"
	"errors"
	"mini-sawit-chain/backend/internal/domain"

	"github.com/google/uuid"
)

func (u *harvestUsecase) Weigh(
	ctx context.Context,
	harvestID uuid.UUID,
	weightKg float64,
) error {

	if weightKg <= 0 {
		return domain.ErrInvalidWeight
	}

	harvest, err := u.harvestRepo.GetByID(ctx, harvestID)
	if err != nil {
		return err
	}

	if harvest.Status != domain.HarvestStatusCreated {
		return errors.New("harvest cannot be weighed")
	}

	return u.harvestRepo.UpdateStatus(
		ctx,
		harvestID,
		domain.HarvestStatusWeighed,
	)
}
