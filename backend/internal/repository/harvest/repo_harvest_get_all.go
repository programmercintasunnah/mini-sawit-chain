package repository_harvest

import (
	"context"
	"mini-sawit-chain/backend/internal/domain"
)

func (r *harvestRepository) GetAll(
	ctx context.Context,
) ([]domain.Harvest, error) {

	query := `
		SELECT
			id,
			farmer_id,
			weight_kg,
			status,
			created_at
		FROM harvests
		WHERE deleted_at IS NULL
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var harvests []domain.Harvest

	for rows.Next() {
		var h domain.Harvest
		if err := rows.Scan(
			&h.ID,
			&h.FarmerID,
			&h.WeightKg,
			&h.Status,
			&h.CreatedAt,
		); err != nil {
			return nil, err
		}
		harvests = append(harvests, h)
	}

	return harvests, nil
}
