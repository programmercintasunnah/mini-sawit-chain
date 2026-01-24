package usecase_harvest

import (
	"context"

	"mini-sawit-chain/backend/internal/domain"
)

func (u *harvestUsecase) GetAll(
	ctx context.Context,
) ([]domain.Harvest, error) {
	return u.harvestRepo.GetAll(ctx)
}
