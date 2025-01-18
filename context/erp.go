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
	InventoryService    interface{} // Contoh: InventoryService
	AuthService         interface{} // Contoh: AuthService
	AdminAuthService    interface{} // Contoh: AdminAuthService
	RBACService         interface{} // Contoh: RBACService
	CompanyService      interface{} // Contoh: CompanyService
	FinanceService      interface{} // Contoh: FinanceService
	OrderService        interface{} // Contoh: OrderService
	DistributionService interface{} // Contoh: DistributionService
	FileService         interface{} // Contoh: FileService
	Firestore           interface{} // Contoh: Firestore
	IndonesiaRegService interface{} // Contoh: IndonesiaRegService
	UserService         interface{} // Contoh: IndonesiaRegService
	AppService          interface{}
	// Add additional services here
	EmailSender  *thirdparty.SMTPSender
	WatzapClient *thirdparty.WatzapClient

	SkipMigration bool // SkipMigration adalah flag untuk menentukan apakah migrasi dijalankan atau tidak
}

// NewERPContext membuat instance baru dari ERPContext
func NewERPContext(db *gorm.DB, req *http.Request, ctx *context.Context, skipMigrate bool) *ERPContext {
	return &ERPContext{
		DB:            db,
		Request:       req,
		Ctx:           ctx,
		SkipMigration: skipMigrate,
	}
}
