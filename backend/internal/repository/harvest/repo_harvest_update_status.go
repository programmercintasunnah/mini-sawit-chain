package repository_harvest

import (
	"context"
	"mini-sawit-chain/backend/internal/domain"

	"github.com/google/uuid"
)

func (r *harvestRepository) UpdateStatus(
	ctx context.Context,
	id uuid.UUID,
	status domain.HarvestStatus,
) error {

	query := `
		UPDATE harvests
		SET status = $1
		WHERE id = $2
		  AND deleted_at IS NULL
	`

	_, err := r.db.ExecContext(ctx, query, status, id)
	return err
}
