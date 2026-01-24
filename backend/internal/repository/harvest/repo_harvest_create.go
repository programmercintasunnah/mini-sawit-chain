package repository_harvest

import (
	"context"
	"mini-sawit-chain/backend/internal/domain"
)

func (r *harvestRepository) Create(
	ctx context.Context,
	harvest *domain.Harvest,
) error {

	query := `
		INSERT INTO harvests (
			id,
			farmer_id,
			weight_kg,
			status,
			created_at
		)
		VALUES ($1, $2, $3, $4, $5)
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		harvest.ID,
		harvest.FarmerID,
		harvest.WeightKg,
		harvest.Status,
		harvest.CreatedAt,
	)

	return err
}
