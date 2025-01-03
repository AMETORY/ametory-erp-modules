package handlers

import (
	"github.com/AMETORY/ametory-erp-modules/company"
	"github.com/gin-gonic/gin"
)

type CompanyHandler struct {
	companyService *company.CompanyService
}

func NewCompanyHandler(companyService *company.CompanyService) *CompanyHandler {
	return &CompanyHandler{companyService: companyService}
}

func (h *CompanyHandler) CreateCompanyHandler(c *gin.Context) {
	// CreateCompanyHandler is a handler function to create a new company.
	c.JSON(200, gin.H{"message": "Company created successfully"})
}
