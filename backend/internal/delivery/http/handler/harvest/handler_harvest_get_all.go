package handler_harvest

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *Handler) GetAll(c *gin.Context) {
	harvests, err := h.harvestUC.GetAll(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var res []HarvestResponse
	for _, hrv := range harvests {
		res = append(res, toResponse(&hrv))
	}

	c.JSON(http.StatusOK, res)
}
