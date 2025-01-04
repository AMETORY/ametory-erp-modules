package handlers

import (
	"net/http"

	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/gin-gonic/gin"
)

type AccountHandler struct {
	ctx *context.ERPContext
}

func NewAccountHandler(ctx *context.ERPContext) *AccountHandler {
	return &AccountHandler{ctx: ctx}
}

func (h *AccountHandler) CreateAccountHandler(c *gin.Context) {
	h.ctx.Request = c.Request
	// Implement logic to create an account
	c.JSON(http.StatusCreated, gin.H{"message": "Account created successfully"})
}

func (h *AccountHandler) GetAccountHandler(c *gin.Context) {
	h.ctx.Request = c.Request
	// Implement logic to get an account
	c.JSON(http.StatusOK, gin.H{"message": "Account retrieved successfully"})
}

func (h *AccountHandler) GetAccountByIdHandler(c *gin.Context) {
	h.ctx.Request = c.Request
	// Implement logic to get an account by ID
	id := c.Param("id")
	c.JSON(http.StatusOK, gin.H{"message": "Account retrieved successfully", "id": id})
}

func (h *AccountHandler) UpdateAccountHandler(c *gin.Context) {
	h.ctx.Request = c.Request
	// Implement logic to update an account
	c.JSON(http.StatusOK, gin.H{"message": "Account updated successfully"})
}

func (h *AccountHandler) DeleteAccountHandler(c *gin.Context) {
	h.ctx.Request = c.Request
	// Implement logic to delete an account
	c.JSON(http.StatusOK, gin.H{"message": "Account deleted successfully"})
}
