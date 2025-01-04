package routes

import (
	"ametory-erp/api/handlers"
	mid "ametory-erp/api/middlewares"

	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/gin-gonic/gin"
)

func SetupFinanceRoutes(r *gin.RouterGroup, ctx *context.ERPContext) {
	accountHandler := handlers.NewAccountHandler(ctx)

	accountGroup := r.Group("/account")
	accountGroup.Use(mid.AuthMiddleware())
	{
		accountGroup.POST("", accountHandler.CreateAccountHandler)
		accountGroup.GET("", accountHandler.GetAccountHandler)
		accountGroup.GET("/:id", accountHandler.GetAccountByIdHandler)
		accountGroup.PUT("/:id", accountHandler.UpdateAccountHandler)
		accountGroup.DELETE("/:id", accountHandler.DeleteAccountHandler)
	}

	transactionHandler := handlers.NewTransactionHandler(ctx)

	transactionGroup := r.Group("/transaction")
	transactionGroup.Use(mid.AuthMiddleware())
	{
		transactionGroup.POST("", transactionHandler.CreateTransactionHandler)
		transactionGroup.GET("", transactionHandler.GetTransactionHandler)
		transactionGroup.GET("/:id", transactionHandler.GetTransactionByIdHandler)
		transactionGroup.PUT("/:id", transactionHandler.UpdateTransactionHandler)
		transactionGroup.DELETE("/:id", transactionHandler.DeleteTransactionHandler)
	}
}
