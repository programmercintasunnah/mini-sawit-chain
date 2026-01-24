package repository_farmer

import (
	"context"
	"mini-sawit-chain/backend/internal/domain"

	"github.com/google/uuid"
)

type FarmerRepository interface {
	Create(ctx context.Context, farmer *domain.Farmer) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Farmer, error)
	GetAll(ctx context.Context) ([]domain.Farmer, error)
	Update(ctx context.Context, farmer *domain.Farmer) error
	SoftDelete(ctx context.Context, id uuid.UUID) error
}
