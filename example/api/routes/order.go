package routes

import (
	"ametory-erp/api/handlers"
	mid "ametory-erp/api/middlewares"

	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/gin-gonic/gin"
)

func SetupOrderRoutes(r *gin.RouterGroup, ctx *context.ERPContext) {
	salesHandler := handlers.NewSalesHandler(ctx)

	salesGroup := r.Group("/sales")
	salesGroup.Use(mid.AuthMiddleware())
	{
		salesGroup.POST("", salesHandler.CreateSalesHandler)
		salesGroup.GET("", salesHandler.GetSalesHandler)
		salesGroup.GET("/:id", salesHandler.GetSalesByIdHandler)
		salesGroup.PUT("/:id", salesHandler.UpdateSalesHandler)
		salesGroup.DELETE("/:id", salesHandler.DeleteSalesHandler)
	}
}
