package http

import (
	"database/sql"
	handler_harvest "mini-sawit-chain/backend/internal/delivery/http/handler/harvest"
	repository_harvest "mini-sawit-chain/backend/internal/repository/harvest"
	usecase_harvest "mini-sawit-chain/backend/internal/usecase/harvest"

	"github.com/gin-gonic/gin"
)

func NewRouter(db *sql.DB) *gin.Engine {
	r := gin.Default()

	repo := repository_harvest.NewRepository(db)
	uc := usecase_harvest.NewUsecase(repo)
	handler := handler_harvest.NewHandler(uc)

	v1 := r.Group("/api/v1")
	{
		harvest := v1.Group("/harvests")
		{
			harvest.POST("", handler.Create)
			harvest.GET("", handler.GetAll)
			harvest.POST("/:id/weigh", handler.Weigh)
			harvest.POST("/:id/pay", handler.Pay)
		}
	}

	return r
}
