package app

import (
	ctx "context"
	"log"
	"net/http"

	"github.com/AMETORY/ametory-erp-modules/app/flow_engine"
	"github.com/AMETORY/ametory-erp-modules/auth"
	"github.com/AMETORY/ametory-erp-modules/company"
	"github.com/AMETORY/ametory-erp-modules/contact"
	"github.com/AMETORY/ametory-erp-modules/content_management"
	"github.com/AMETORY/ametory-erp-modules/context"
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
	"github.com/AMETORY/ametory-erp-modules/thirdparty"
	"github.com/AMETORY/ametory-erp-modules/thirdparty/ai_generator"
	"github.com/AMETORY/ametory-erp-modules/thirdparty/email_api"
	"github.com/AMETORY/ametory-erp-modules/thirdparty/google"
	"github.com/AMETORY/ametory-erp-modules/thirdparty/kafka"
	"github.com/AMETORY/ametory-erp-modules/thirdparty/redis"
	"github.com/AMETORY/ametory-erp-modules/thirdparty/websocket"
	"github.com/AMETORY/ametory-erp-modules/thirdparty/whatsmeow_client"
	"github.com/AMETORY/ametory-erp-modules/user"
	"gorm.io/gorm"
)

type AppContainer struct {
	erpContext                  *context.ERPContext // Context for the ERP application
	DB                          *gorm.DB            // Database connection
	Ctx                         ctx.Context         // Context
	Request                     *http.Request       // HTTP request
	InventoryService            *inventory.InventoryService
	ManufactureService          *manufacture.ManufactureService
	AuthService                 *auth.AuthService
	AdminAuthService            *auth.AdminAuthService
	RBACService                 *auth.RBACService
	CompanyService              *company.CompanyService
	ContactService              *contact.ContactService
	FinanceService              *finance.FinanceService
	CooperativeService          *cooperative.CooperativeService
	OrderService                *order.OrderService
	LogisticService             *logistic.LogisticService
	DistributionService         *distribution.DistributionService
	CustomerRelationshipService *customer_relationship.CustomerRelationshipService
	FileService                 *file.FileService
	MedicalService              *medical.MedicalService
	IndonesiaRegService         *indonesia_regional.IndonesiaRegService
	UserService                 *user.UserService
	ContentManagementService    *content_management.ContentManagementService
	TagService                  *tag.TagService
	MessageService              *message.MessageService
	ProjectManagementService    *project.ProjectService
	PlanningBudgetService       *planning_budget.PlanningBudgetService
	CrowdFundingService         *crowd_funding.CrowdFundingService
	NotificationService         *notification.NotificationService
	HRISService                 *hris.HRISservice
	AuditTrailService           *audit_trail.AuditTrailService
	PermitHubService            *permit_hub.PermitHubService
	AiGeneratorService          ai_generator.AiGenerator

	ThirdPartyServices map[string]any
	// Add additional services here
	EmailSender      *thirdparty.SMTPSender
	EmailAPIService  *email_api.EmailApiService
	WatzapClient     *thirdparty.WatzapClient
	GoogleAPIService *google.GoogleAPIService
	GeminiService    *google.GeminiService
	SkipMigration    bool
	Firestore        *thirdparty.Firestore
	WhatsmeowService *whatsmeow_client.WhatsmeowService
	FCMService       *google.FCMService
	RedisService     *redis.RedisService
	KafkaService     *kafka.KafkaService
	WebsocketService *websocket.WebsocketService
	AppService       any    // This can be a specific service or a generic interface
	baseURL          string // Base URL for the application

	// FLOW ENGINE
	FlowEngine *flow_engine.FlowEngine
}

type AppContainerOption func(*AppContainer)

func NewAppContainer(db *gorm.DB, req *http.Request, golangContext *ctx.Context, skipMigrate bool, baseURL string, opts ...AppContainerOption) *AppContainer {
	if skipMigrate {
		log.Println("Skipping migration ...")
	}
	container := &AppContainer{
		erpContext:         context.NewERPContext(db, req, golangContext, skipMigrate),
		DB:                 db,
		Ctx:                *golangContext,
		Request:            req,
		SkipMigration:      skipMigrate,
		baseURL:            baseURL,
		ThirdPartyServices: make(map[string]any),
	}

	for _, opt := range opts {
		opt(container)
	}

	return container
}
