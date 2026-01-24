package handler_harvest

import "mini-sawit-chain/backend/internal/domain"

func toResponse(h *domain.Harvest) HarvestResponse {
	return HarvestResponse{
		ID:        h.ID,
		FarmerID:  h.FarmerID,
		WeightKg:  h.WeightKg,
		Status:    string(h.Status),
		CreatedAt: h.CreatedAt,
	}
}
