package usecase_harvest

import (
	"context"

	"mini-sawit-chain/backend/internal/domain"

	"github.com/google/uuid"
)

func (u *harvestUsecase) GetByID(
	ctx context.Context,
	id uuid.UUID,
) (*domain.Harvest, error) {
	return u.harvestRepo.GetByID(ctx, id)
}
