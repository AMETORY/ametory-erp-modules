package main

import (
	"ametory-erp/api/routes"
	"ametory-erp/config"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/AMETORY/ametory-erp-modules/company"
	"github.com/AMETORY/ametory-erp-modules/modules"
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

	log.Println("Log file created:", logPath)
}

func main() {
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
	registerService(r, db)

	// Setup routes for finance service

	// Start the server
	fmt.Println("ERP system is running on :8080...")
	r.Run(":" + cfg.Server.Port)
}

func registerService(r *gin.Engine, db *gorm.DB) error {
	_, err := modules.RegisterService("auth", db, false)
	if err != nil {
		log.Fatalf("Gagal menjalankan migrasi service auth: %v", err)
		return err

	}
	companyService, err := modules.RegisterService("company", db, false)
	if err != nil {
		log.Fatalf("Gagal menjalankan migrasi service company: %v", err)
		return err

	}
	routes.SetupCompanyRoutes(r, companyService.(*company.CompanyService))
	return nil
}
