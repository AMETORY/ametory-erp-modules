package context

import (
	"context"
	"net/http"

	"gorm.io/gorm"
)

// ERPContext adalah struct yang menyimpan semua dependencies
type ERPContext struct {
	DB      *gorm.DB         // Database connection
	Ctx     *context.Context // Context
	Request *http.Request    // HTTP request
	// Tambahkan service lainnya di sini
	InventoryService interface{} // Contoh: InventoryService
	AuthService      interface{} // Contoh: AuthService
	CompanyService   interface{} // Contoh: CompanyService
	FinanceService   interface{} // Contoh: FinanceService
	OrderService     interface{} // Contoh: OrderService
	// Tambahkan service lainnya di sini

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