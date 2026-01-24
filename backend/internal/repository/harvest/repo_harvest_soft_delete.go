package repository_harvest

import (
	"context"
	"time"

	"github.com/google/uuid"
)

func (r *harvestRepository) SoftDelete(
	ctx context.Context,
	id uuid.UUID,
) error {

	query := `
		UPDATE harvests
		SET deleted_at = $1
		WHERE id = $2
	`

	_, err := r.db.ExecContext(ctx, query, time.Now(), id)
	return err
}
