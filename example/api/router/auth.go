package router

import (
	"ametory-erp/api/handler"

	"github.com/AMETORY/ametory-erp-modules/app"
	"github.com/gin-gonic/gin"
)

func SetUpAuthRoutes(router *gin.RouterGroup, appContainer *app.AppContainer) {
	authHandler := handler.NewAuthHandler(appContainer)
	authGroup := router.Group("/auth")
	{
		authGroup.POST("/login", authHandler.Login)
	}
}
