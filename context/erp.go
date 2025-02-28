package context

import (
	"context"
	"net/http"

	"github.com/AMETORY/ametory-erp-modules/thirdparty"
	"gorm.io/gorm"
)

// ERPContext adalah struct yang menyimpan semua dependencies
type ERPContext struct {
	DB      *gorm.DB         // Database connection
	Ctx     *context.Context // Context
	Request *http.Request    // HTTP request
	// Tambahkan service lainnya di sini
	InventoryService            interface{} // Contoh: InventoryService
	AuthService                 interface{} // Contoh: AuthService
	AdminAuthService            interface{} // Contoh: AdminAuthService
	RBACService                 interface{} // Contoh: RBACService
	CompanyService              interface{} // Contoh: CompanyService
	ContactService              interface{} // Contoh: ContactService
	FinanceService              interface{} // Contoh: FinanceService
	OrderService                interface{} // Contoh: OrderService
	DistributionService         interface{} // Contoh: DistributionService
	CustomerRelationshipService interface{} // Contoh: CustomerRelationshipService
	FileService                 interface{} // Contoh: FileService
	MedicalService              interface{} // Contoh: MedicalService
	Firestore                   interface{} // Contoh: Firestore
	FCMService                  interface{} // Contoh: Firestore
	IndonesiaRegService         interface{} // Contoh: IndonesiaRegService
	UserService                 interface{} // Contoh: UserService
	ContentManagementService    interface{}
	ProjectManagementService    interface{}
	AppService                  interface{}
	CrowdFundingService         interface{}
	NotificationService         interface{}
	InternalService             interface{}
	HRISService                 interface{}
	TempData                    interface{}
	Config                      ctxConfig

	ThirdPartyServices map[string]interface{}
	// Add additional services here
	EmailSender  *thirdparty.SMTPSender
	WatzapClient *thirdparty.WatzapClient

	SkipMigration bool // SkipMigration adalah flag untuk menentukan apakah migrasi dijalankan atau tidak
}

// NewERPContext membuat instance baru dari ERPContext
func NewERPContext(db *gorm.DB, req *http.Request, ctx *context.Context, skipMigrate bool) *ERPContext {
	return &ERPContext{
		DB:                 db,
		Request:            req,
		Ctx:                ctx,
		SkipMigration:      skipMigrate,
		ThirdPartyServices: make(map[string]interface{}, 0),
	}
}

func (erp *ERPContext) SetConfig(wkhtmltopdfPath, pdfFooter string) {
	erp.Config.WkhtmltopdfPath = wkhtmltopdfPath
	erp.Config.PdfFooter = pdfFooter
}
func (erp *ERPContext) AddThirdPartyService(name string, service interface{}) {
	erp.ThirdPartyServices[name] = service
}

type ctxConfig struct {
	WkhtmltopdfPath string
	PdfFooter       string
}

func (erp *ERPContext) AlterColumn(dst interface{}, field string) error {
	if !erp.SkipMigration {
		return erp.DB.Migrator().AlterColumn(dst, field)
	}
	return nil
}
func (erp *ERPContext) Migrate(models ...interface{}) error {
	if !erp.SkipMigration {

		return erp.DB.AutoMigrate(models)
	}
	return nil
}
