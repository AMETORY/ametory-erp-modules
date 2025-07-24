package app

import (
	"log"

	"github.com/AMETORY/ametory-erp-modules/auth"
	"github.com/AMETORY/ametory-erp-modules/company"
	"github.com/AMETORY/ametory-erp-modules/contact"
	"github.com/AMETORY/ametory-erp-modules/content_management"
	"github.com/AMETORY/ametory-erp-modules/cooperative"
	"github.com/AMETORY/ametory-erp-modules/crowd_funding"
	"github.com/AMETORY/ametory-erp-modules/customer_relationship"
	"github.com/AMETORY/ametory-erp-modules/distribution"
	"github.com/AMETORY/ametory-erp-modules/distribution/logistic"
	"github.com/AMETORY/ametory-erp-modules/file"
	"github.com/AMETORY/ametory-erp-modules/finance"
	"github.com/AMETORY/ametory-erp-modules/hris"
	"github.com/AMETORY/ametory-erp-modules/inventory"
	"github.com/AMETORY/ametory-erp-modules/manufacture"
	"github.com/AMETORY/ametory-erp-modules/medical"
	"github.com/AMETORY/ametory-erp-modules/message"
	"github.com/AMETORY/ametory-erp-modules/notification"
	"github.com/AMETORY/ametory-erp-modules/order"
	"github.com/AMETORY/ametory-erp-modules/permit_hub"
	"github.com/AMETORY/ametory-erp-modules/planning_budget"
	"github.com/AMETORY/ametory-erp-modules/project_management/project"
	"github.com/AMETORY/ametory-erp-modules/shared/audit_trail"
	"github.com/AMETORY/ametory-erp-modules/shared/indonesia_regional"
	"github.com/AMETORY/ametory-erp-modules/tag"
	"github.com/AMETORY/ametory-erp-modules/user"
)

// WithAuth adds the AuthService to the AppContainer.
//
// It is a required option.
func WithAuth() AppContainerOption {
	return func(c *AppContainer) {
		c.AuthService = auth.NewAuthService(c.erpContext)
		log.Println("AuthService initialized")
	}
}

// WithAdminAuth adds the AdminAuthService to the AppContainer.
func WithAdminAuth() AppContainerOption {
	return func(c *AppContainer) {
		c.AdminAuthService = auth.NewAdminAuthService(c.erpContext)
		log.Println("AdminAuthService initialized")
	}
}

// WithRBAC adds the RBACService to the AppContainer.
//
// It is an optional option.
func WithRBAC() AppContainerOption {
	return func(c *AppContainer) {
		c.RBACService = auth.NewRBACService(c.erpContext)
		log.Println("RBACService initialized")
	}
}

// WithInventory adds the InventoryService to the AppContainer.
//
// It is an optional option.
func WithInventory() AppContainerOption {
	return func(c *AppContainer) {
		c.InventoryService = inventory.NewInventoryService(c.erpContext)
		log.Println("InventoryService initialized")
	}
}

// WithManufacture adds the ManufactureService to the AppContainer.
//
// It is an optional option. It requires WithInventory to be called before it.
func WithManufacture() AppContainerOption {
	return func(c *AppContainer) {
		c.ManufactureService = manufacture.NewManufactureService(c.erpContext, c.InventoryService)
		log.Println("ManufactureService initialized")
	}
}

// WithCompany adds the CompanyService to the AppContainer.
//
// It is an optional option.
func WithCompany() AppContainerOption {
	return func(c *AppContainer) {
		c.CompanyService = company.NewCompanyService(c.erpContext)
		log.Println("CompanyService initialized")
	}
}

// WithContact adds the ContactService to the AppContainer.
//
// It is an optional option. It requires WithCompany to be called before it.
func WithContact() AppContainerOption {
	return func(c *AppContainer) {
		c.ContactService = contact.NewContactService(c.erpContext, c.CompanyService)
		log.Println("ContactService initialized")
	}
}

// WithFinance adds the FinanceService to the AppContainer.
//
// It is an optional option.
func WithFinance() AppContainerOption {
	return func(c *AppContainer) {
		c.FinanceService = finance.NewFinanceService(c.erpContext)
		log.Println("FinanceService initialized")
	}
}

// WithCooperative adds the CooperativeService to the AppContainer.
//
// It is an optional option. It requires WithCompany and WithFinance to be called before it.
func WithCooperative() AppContainerOption {
	return func(c *AppContainer) {
		c.CooperativeService = cooperative.NewCooperativeService(c.erpContext, c.CompanyService, c.FinanceService)
		log.Println("CooperativeService initialized")
	}
}

// WithOrder adds the OrderService to the AppContainer.
//
// It is an optional option.

func WithOrder() AppContainerOption {
	return func(c *AppContainer) {
		c.OrderService = order.NewOrderService(c.erpContext)
		log.Println("OrderService initialized")
	}
}

// WithLogistic adds the LogisticService to the AppContainer.
//
// It is an optional option. It requires WithInventory to be called before it.
func WithLogistic() AppContainerOption {
	return func(c *AppContainer) {
		c.LogisticService = logistic.NewLogisticService(c.DB, c.erpContext, c.InventoryService)
		log.Println("LogisticService initialized")
	}
}

// WithAuditTrail adds the AuditTrailService to the AppContainer.
//
// It is an optional option.
func WithAuditTrail() AppContainerOption {
	return func(c *AppContainer) {
		c.AuditTrailService = audit_trail.NewAuditTrailService(c.erpContext)
		log.Println("AuditTrailService initialized")
	}
}

// WithDistribution adds the DistributionService to the AppContainer.
//
// It is an optional option. It requires WithAuditTrail, WithInventory, and WithOrder to be called before it.

func WithDistribution() AppContainerOption {
	return func(c *AppContainer) {
		c.DistributionService = distribution.NewDistributionService(c.erpContext, c.AuditTrailService, c.InventoryService, c.OrderService)
		log.Println("DistributionService initialized")
	}
}

// WithCustomerRelationship adds the CustomerRelationshipService to the AppContainer.
//
// It is an optional option.
func WithCustomerRelationship() AppContainerOption {
	return func(c *AppContainer) {
		c.CustomerRelationshipService = customer_relationship.NewCustomerRelationshipService(c.erpContext)
		log.Println("CustomerRelationshipService initialized")
	}
}

// WithFile adds the FileService to the AppContainer.
//
// It is an optional option.

func WithFile() AppContainerOption {
	return func(c *AppContainer) {
		c.FileService = file.NewFileService(c.erpContext, c.baseURL)
		c.erpContext.Firestore = c.Firestore
		log.Println("FileService initialized")
	}
}

// WithMedical adds the MedicalService to the AppContainer.
//
// It is an optional option.
func WithMedical() AppContainerOption {
	return func(c *AppContainer) {
		c.MedicalService = medical.NewMedicalService(c.DB, c.erpContext)
		log.Println("MedicalService initialized")
	}
}

// WithIndonesiaReg adds the IndonesiaRegService to the AppContainer.
//
// It is an optional option.
func WithIndonesiaReg() AppContainerOption {
	return func(c *AppContainer) {
		c.IndonesiaRegService = indonesia_regional.NewIndonesiaRegService(c.erpContext)
		log.Println("IndonesiaRegService initialized")
	}
}

// WithUser adds the UserService to the AppContainer.
//
// It is an optional option.
func WithUser() AppContainerOption {
	return func(c *AppContainer) {
		c.UserService = user.NewUserService(c.erpContext)
		log.Println("UserService initialized")
	}
}

// WithContentManagement adds the ContentManagementService to the AppContainer.
//
// It is an optional option.
func WithContentManagement() AppContainerOption {
	return func(c *AppContainer) {
		c.ContentManagementService = content_management.NewContentManagementService(c.erpContext)
		log.Println("ContentManagementService initialized")
	}
}

// WithTag adds the TagService to the AppContainer.
//
// It is an optional option.
func WithTag() AppContainerOption {
	return func(c *AppContainer) {
		c.TagService = tag.NewTagService(c.erpContext)
		log.Println("TagService initialized")
	}
}

// WithMessage adds the MessageService to the AppContainer.
//
// It is an optional option.
func WithMessage() AppContainerOption {
	return func(c *AppContainer) {
		c.MessageService = message.NewMessageService(c.erpContext)
		log.Println("MessageService initialized")
	}
}

// WithProjectManagement adds the ProjectManagementService to the AppContainer.
//
// It is an optional option.
func WithProjectManagement() AppContainerOption {
	return func(c *AppContainer) {
		c.ProjectManagementService = project.NewProjectService(c.erpContext)
		log.Println("ProjectManagementService initialized")
	}
}

// WithCrowdFunding adds the CrowdFundingService to the AppContainer.
//
// It is an optional option.
func WithCrowdFunding() AppContainerOption {
	return func(c *AppContainer) {
		c.CrowdFundingService = crowd_funding.NewCrowdFundingService(c.erpContext)
		log.Println("CrowdFundingService initialized")
	}
}

// WithNotification adds the NotificationService to the AppContainer.
//
// It is an optional option.
func WithNotification() AppContainerOption {
	return func(c *AppContainer) {
		c.NotificationService = notification.NewNotificationService(c.erpContext)
		log.Println("NotificationService initialized")
	}
}

// WithHRIS adds the HRISService to the AppContainer.
//
// It is an optional option.
func WithHRIS() AppContainerOption {
	return func(c *AppContainer) {
		c.HRISService = hris.NewHRISservice(c.erpContext)
		log.Println("HRISService initialized")
	}
}

// WithPermitHub adds the PermitHubService to the AppContainer.
//
// It is an optional option.
func WithPermitHub() AppContainerOption {
	return func(c *AppContainer) {
		c.PermitHubService = permit_hub.NewPermitHubService(c.erpContext)
		log.Println("PermitHubService initialized")
	}
}

// WithPlanningBudget adds the PlanningBudgetService to the AppContainer.
//
// It is an optional option.
func WithPlanningBudget() AppContainerOption {
	return func(c *AppContainer) {
		c.PlanningBudgetService = planning_budget.NewPlanningBudgetService(c.erpContext)
		log.Println("PlanningBudgetService initialized")
	}
}
