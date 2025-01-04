package handlers

import (
	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/gin-gonic/gin"
)

type CompanyHandler struct {
	ctx *context.ERPContext
}

func NewCompanyHandler(ctx *context.ERPContext) *CompanyHandler {
	return &CompanyHandler{ctx: ctx}
}

func (h *CompanyHandler) CreateCompanyHandler(c *gin.Context) {
	h.ctx.Request = c.Request
	// CreateCompanyHandler is a handler function to create a new company.
	c.JSON(200, gin.H{"message": "Company created successfully"})
}
