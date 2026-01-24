package handler_harvest

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type WeighRequest struct {
	WeightKg float64 `json:"weight_kg" validate:"required,gt=0"`
}

func (h *Handler) Weigh(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var req WeighRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.harvestUC.Weigh(
		c.Request.Context(),
		id,
		req.WeightKg,
	); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}
