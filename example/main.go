package main

import (
	"ametory-erp/api/routes"
	"ametory-erp/config"
	ctx "context"
	"fmt"
	"log"
	"net/mail"
	"os"
	"path/filepath"
	"time"

	"github.com/AMETORY/ametory-erp-modules/auth"
	"github.com/AMETORY/ametory-erp-modules/company"
	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/finance"
	"github.com/AMETORY/ametory-erp-modules/inventory"
	"github.com/AMETORY/ametory-erp-modules/order"
	"github.com/AMETORY/ametory-erp-modules/thirdparty"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func initLog() {
	t := time.Now()
	filename := t.Format("2006-01-02")
	logDir := "log"
	logPath := filepath.Join(logDir, filename+".log")

	if err := os.MkdirAll(logDir, os.ModePerm); err != nil {
		log.Fatalf("error creating directory: %v", err)
	}
	f, err := os.OpenFile(logPath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer f.Close()
	log.SetOutput(f)

	log.Println("log started at:", time.Now().Format("2006-01-02 15:04:05"))
}

func main() {
	ctx := ctx.Background()
	initLog()
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Gagal memuat konfigurasi: %v", err)
	}

	// Initialize database
	db, err := config.InitDB(cfg)
	if err != nil {
		log.Fatalf("Gagal menghubungkan ke database: %v", err)
	}

	fmt.Println("Berhasil terhubung ke database", cfg.Database.Name)
	// Initialize Gin
	r := gin.Default()
	registerService(r, db, &ctx, cfg.Email)

	// Setup routes for finance service

	// Start the server
	fmt.Println("ERP system is running on :8080...")
	r.Run(":" + cfg.Server.Port)
}

func registerService(r *gin.Engine, db *gorm.DB, ctx *ctx.Context, emailConfig config.EmailConfig) error {
	erpContext := context.NewERPContext(db, nil, ctx, false)
	authSrv := auth.NewAuthService(db)
	companySrv := company.NewCompanyService(erpContext)
	financeSrv := finance.NewFinanceService(erpContext)
	invSrv := inventory.NewInventoryService(erpContext)
	orderSrv := order.NewOrderService(erpContext)

	erpContext.AuthService = authSrv
	erpContext.CompanyService = companySrv
	erpContext.FinanceService = financeSrv
	erpContext.InventoryService = invSrv
	erpContext.OrderService = orderSrv

	// THIRD PARTY SERVICE
	emailSender := thirdparty.NewSMTPSender(emailConfig.Server, emailConfig.Port, emailConfig.Username, emailConfig.Password, mail.Address{Name: emailConfig.From, Address: emailConfig.From})
	emailSender.SetTemplate("templates/email/layout.html", "templates/email/body.html")

	erpContext.EmailSender = emailSender

	v1 := r.Group("/v1")

	routes.SetupAuthRoutes(v1, erpContext)
	routes.SetupCompanyRoutes(v1, erpContext)
	routes.SetupFinanceRoutes(v1, erpContext)
	routes.SetupInventoryRoutes(v1, erpContext)
	routes.SetupOrderRoutes(v1, erpContext)
	return nil
}
