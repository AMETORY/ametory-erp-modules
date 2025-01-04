package routes

import (
	"ametory-erp/api/handlers"

	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/gin-gonic/gin"
)

func SetupAuthRoutes(r *gin.RouterGroup, ctx *context.ERPContext) {
	authHandler := handlers.NewAuthHandler(ctx)

	authGroup := r.Group("/auth")
	authGroup.Use()
	{
		authGroup.POST("/register", authHandler.RegisterHandler)
		authGroup.POST("/login", authHandler.LoginHandler)
	}
}
