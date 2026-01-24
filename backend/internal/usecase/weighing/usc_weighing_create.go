package usecase_weighing

// import (
// 	"context"
// 	"mini-sawit-chain/backend/internal/domain"
// )

// func (uc *WeighingUsecase) Create(ctx context.Context, input CreateWeighingInput) error {
// 	harvest, err := uc.harvestRepo.GetByID(ctx, input.HarvestID)
// 	if err != nil {
// 		return err
// 	}

// 	if harvest.Status != domain.HarvestStatusCreated {
// 		return ErrHarvestAlreadyWeighed
// 	}

// 	err = uc.tx.RunInTx(ctx, func(txCtx context.Context) error {
// 		err := uc.weighingRepo.Create(txCtx, weighing)
// 		if err != nil {
// 			return err
// 		}

// 		return uc.harvestRepo.UpdateStatus(
// 			txCtx,
// 			harvest.ID,
// 			domain.HarvestStatusWeighed,
// 		)
// 	})

// 	return err
// }
