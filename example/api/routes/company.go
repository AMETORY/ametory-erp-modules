package routes

import (
	"ametory-erp/api/handlers"
	mid "ametory-erp/api/middlewares"

	"github.com/AMETORY/ametory-erp-modules/company"
	"github.com/gin-gonic/gin"
)

func SetupCompanyRoutes(r *gin.Engine, companyService *company.CompanyService) {
	companyHandler := handlers.NewCompanyHandler(companyService)

	companyGroup := r.Group("/company")
	companyGroup.Use(mid.AuthMiddleware())
	{
		companyGroup.POST("", companyHandler.CreateCompanyHandler)
	}
}
