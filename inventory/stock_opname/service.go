package stock_opname

import (
	"errors"
	"time"

	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/inventory/product"
	stockmovement "github.com/AMETORY/ametory-erp-modules/inventory/stock_movement"
	"github.com/AMETORY/ametory-erp-modules/shared/models"
	"gorm.io/gorm"
)

type StockOpnameService struct {
	db                   *gorm.DB
	ctx                  *context.ERPContext
	productService       *product.ProductService
	stockMovementService *stockmovement.StockMovementService
}

func NewStockOpnameService(db *gorm.DB, ctx *context.ERPContext, productService *product.ProductService, stockMovementService *stockmovement.StockMovementService) *StockOpnameService {
	return &StockOpnameService{db: db, ctx: ctx, productService: productService, stockMovementService: stockMovementService}
}

func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(&models.StockOpnameHeader{}, &models.StockOpnameDetail{})
}

func (s *StockOpnameService) CreateStockOpname(warehouseID string, products []models.ProductModel, notes string) (*models.StockOpnameHeader, error) {
	if s.productService == nil {
		return nil, errors.New("product service is not initialized")
	}
	// Buat stock opname header
	stockOpnameHeader := models.StockOpnameHeader{
		WarehouseID: warehouseID,
		Status:      models.StatusDraft,
		OpnameDate:  time.Now(),
		Notes:       notes,
	}

	// Simpan stock opname header ke database
	if err := s.db.Create(&stockOpnameHeader).Error; err != nil {
		return nil, err
	}

	// Buat stock opname detail untuk setiap produk
	for _, product := range products {
		// Dapatkan stok sistem dari inventory service
		systemQty, err := s.productService.GetStock(product.ID, nil, &warehouseID)
		if err != nil {
			return nil, err
		}

		// Hitung selisih stok
		difference := product.TotalStock - systemQty

		// Buat stock opname detail
		stockOpnameDetail := models.StockOpnameDetail{
			StockOpnameID: stockOpnameHeader.ID,
			ProductID:     product.ID,
			Quantity:      product.TotalStock,
			SystemQty:     systemQty,
			Difference:    difference,
		}
		if err := s.db.Create(&stockOpnameDetail).Error; err != nil {
			return nil, err
		}
	}

	return &stockOpnameHeader, nil
}

func (s *StockOpnameService) CompleteStockOpname(stockOpnameID string) error {
	// Dapatkan stock opname header
	var stockOpnameHeader models.StockOpnameHeader
	if err := s.db.Preload("Details").First(&stockOpnameHeader, stockOpnameID).Error; err != nil {
		return err
	}

	// Update stok di sistem untuk setiap produk
	for _, detail := range stockOpnameHeader.Details {
		if detail.Difference != 0 {
			if _, err := s.stockMovementService.CreateAdjustment(
				time.Now(),
				detail.ProductID,
				stockOpnameHeader.WarehouseID,
				detail.VariantID,
				nil,
				nil,
				nil,
				detail.Difference,
				stockOpnameHeader.ID,
				stockOpnameHeader.Notes,
			); err != nil {
				return err
			}
		}
	}

	// Update status stock opname menjadi "COMPLETED"
	return s.db.Model(&stockOpnameHeader).Update("status", models.StatusCompleted).Error
}

type StockDiscrepancyReport struct {
	ProductID   string `json:"product_id"`
	ProductName string `json:"product_name"`
	PhysicalQty int    `json:"physical_qty"`
	SystemQty   int    `json:"system_qty"`
	Difference  int    `json:"difference"`
	Notes       string `json:"notes"`
}

func (s *StockOpnameService) GenerateDiscrepancyReport(stockOpnameID string) ([]StockDiscrepancyReport, error) {
	var report []StockDiscrepancyReport
	err := s.db.Table("stock_opname_details").
		Joins("JOIN products ON stock_opname_details.product_id = products.id").
		Where("stock_opname_details.stock_opname_id = ?", stockOpnameID).
		Select("stock_opname_details.product_id, products.name as product_name, stock_opname_details.quantity as physical_qty, stock_opname_details.system_qty, stock_opname_details.difference, stock_opname_details.notes").
		Scan(&report).Error
	return report, err
}
