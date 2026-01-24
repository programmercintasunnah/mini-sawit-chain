package usecase_harvest

import (
	"context"
	"errors"
	"mini-sawit-chain/backend/internal/domain"

	"github.com/google/uuid"
)

func (u *harvestUsecase) Pay(
	ctx context.Context,
	harvestID uuid.UUID,
) error {

	harvest, err := u.harvestRepo.GetByID(ctx, harvestID)
	if err != nil {
		return err
	}

	if harvest.Status != domain.HarvestStatusWeighed {
		return errors.New("harvest must be weighed before payment")
	}

	return u.harvestRepo.UpdateStatus(
		ctx,
		harvestID,
		domain.HarvestStatusPaid,
	)
}
