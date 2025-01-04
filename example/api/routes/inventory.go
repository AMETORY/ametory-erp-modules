package routes

import (
	"ametory-erp/api/handlers"
	mid "ametory-erp/api/middlewares"

	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/gin-gonic/gin"
)

func SetupInventoryRoutes(r *gin.RouterGroup, ctx *context.ERPContext) {
	productHandler := handlers.NewProductHandler(ctx)

	productGroup := r.Group("/product")
	productGroup.Use(mid.AuthMiddleware())
	{
		productGroup.POST("", productHandler.CreateProductHandler)
		productGroup.GET("", productHandler.GetProductHandler)
		productGroup.GET("/:id", productHandler.GetProductByIdHandler)
		productGroup.PUT("/:id", productHandler.UpdateProductHandler)
		productGroup.DELETE("/:id", productHandler.DeleteProductHandler)
	}

	stockMovementHandler := handlers.NewStockMovementHandler(ctx)

	stockMovementGroup := r.Group("/stock-movement")
	stockMovementGroup.Use(mid.AuthMiddleware())
	{
		stockMovementGroup.POST("", stockMovementHandler.CreateStockMovementHandler)
		stockMovementGroup.GET("", stockMovementHandler.GetStockMovementHandler)
		stockMovementGroup.GET("/:id", stockMovementHandler.GetStockMovementByIdHandler)
		stockMovementGroup.PUT("/:id", stockMovementHandler.UpdateStockMovementHandler)
		stockMovementGroup.DELETE("/:id", stockMovementHandler.DeleteStockMovementHandler)
	}

	warehouseHandler := handlers.NewWarehouseHandler(ctx)

	warehouseGroup := r.Group("/warehouse")
	warehouseGroup.Use(mid.AuthMiddleware())
	{
		warehouseGroup.POST("", warehouseHandler.CreateWarehouseHandler)
		warehouseGroup.GET("", warehouseHandler.GetWarehouseHandler)
		warehouseGroup.GET("/:id", warehouseHandler.GetWarehouseByIdHandler)
		warehouseGroup.PUT("/:id", warehouseHandler.UpdateWarehouseHandler)
		warehouseGroup.DELETE("/:id", warehouseHandler.DeleteWarehouseHandler)
	}

}
