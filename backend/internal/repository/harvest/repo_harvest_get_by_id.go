package repository_harvest

import (
	"context"
	"database/sql"
	"mini-sawit-chain/backend/internal/domain"

	"github.com/google/uuid"
)

func (r *harvestRepository) GetByID(
	ctx context.Context,
	id uuid.UUID,
) (*domain.Harvest, error) {

	query := `
		SELECT
			id,
			farmer_id,
			weight_kg,
			status,
			created_at
		FROM harvests
		WHERE id = $1
		  AND deleted_at IS NULL
	`

	row := r.db.QueryRowContext(ctx, query, id)

	var harvest domain.Harvest
	if err := row.Scan(
		&harvest.ID,
		&harvest.FarmerID,
		&harvest.WeightKg,
		&harvest.Status,
		&harvest.CreatedAt,
	); err != nil {

		if err == sql.ErrNoRows {
			return nil, domain.ErrHarvestNotFound
		}
		return nil, err
	}

	return &harvest, nil
}
