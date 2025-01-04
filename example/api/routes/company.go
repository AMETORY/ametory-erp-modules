package routes

import (
	"ametory-erp/api/handlers"
	mid "ametory-erp/api/middlewares"

	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/gin-gonic/gin"
)

func SetupCompanyRoutes(r *gin.RouterGroup, ctx *context.ERPContext) {
	companyHandler := handlers.NewCompanyHandler(ctx)

	companyGroup := r.Group("/company")
	companyGroup.Use(mid.AuthMiddleware())
	{
		companyGroup.POST("", companyHandler.CreateCompanyHandler)
	}
}
