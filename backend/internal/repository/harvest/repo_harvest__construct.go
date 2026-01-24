package repository_harvest

import (
	"context"
	"database/sql"
	"mini-sawit-chain/backend/internal/domain"

	"github.com/google/uuid"
)

type HarvestRepository interface {
	Create(ctx context.Context, harvest *domain.Harvest) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Harvest, error)
	GetAll(ctx context.Context) ([]domain.Harvest, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status domain.HarvestStatus) error
	SoftDelete(ctx context.Context, id uuid.UUID) error
}

type harvestRepository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *harvestRepository {
	return &harvestRepository{
		db: db,
	}
}
