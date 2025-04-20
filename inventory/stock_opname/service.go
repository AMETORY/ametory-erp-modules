package stock_opname

import (
	"errors"
	"math"
	"time"

	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/inventory/product"
	stockmovement "github.com/AMETORY/ametory-erp-modules/inventory/stock_movement"
	"github.com/AMETORY/ametory-erp-modules/shared"
	"github.com/AMETORY/ametory-erp-modules/shared/models"
	"github.com/AMETORY/ametory-erp-modules/utils"
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

func (s *StockOpnameService) CreateStockOpnameFromHeader(warehouseID string, data *models.StockOpnameHeader, notes string) error {
	return s.db.Create(data).Error
}

func (s *StockOpnameService) UpdateStockOpname(stockOpnameID string, data *models.StockOpnameHeader) error {
	return s.db.Model(&models.StockOpnameHeader{}).
		Where("id = ?", stockOpnameID).
		Updates(data).
		Error
}

func (s *StockOpnameService) DeleteStockOpname(stockOpnameID string) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		// Delete stock opname detail
		if err := tx.Where("stock_opname_id = ?", stockOpnameID).Delete(&models.StockOpnameDetail{}).Error; err != nil {
			return err
		}

		// Delete related inventory transactions
		if err := tx.Where("reference_id = ?", stockOpnameID).Delete(&models.StockMovementModel{}).Error; err != nil {
			return err
		}

		// Delete related inventory transactions
		if err := tx.Where("transaction_secondary_ref_id = ?", stockOpnameID).Delete(&models.TransactionModel{}).Error; err != nil {
			return err
		}

		// Delete stock opname header
		return tx.Where("id = ?", stockOpnameID).Delete(&models.StockOpnameHeader{}).Error
	})
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

func (s *StockOpnameService) CompleteStockOpname(stockOpnameID string, date time.Time, userID string, inventoryID, expenseID, revenueID *string) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		var stockOpnameHeader models.StockOpnameHeader
		if err := tx.Preload("Details").First(&stockOpnameHeader, stockOpnameID).Error; err != nil {
			return err
		}

		// Update stok di sistem untuk setiap produk
		for _, detail := range stockOpnameHeader.Details {
			if detail.Difference != 0 {
				movement, err := s.stockMovementService.CreateAdjustment(
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
				)
				if err != nil {
					return err
				}

				refType := "stock_opname"
				secRefType := "stock_opname_detail"
				movement.ReferenceType = &refType
				movement.SecondaryRefID = &detail.ID
				movement.SecondaryRefType = &secRefType
				movement.Value = detail.UnitValue
				movement.UnitID = detail.UnitID

				err = tx.Save(movement).Error
				if err != nil {
					return err
				}

				if inventoryID != nil && expenseID != nil && revenueID != nil {
					inventoryTransID := utils.Uuid()
					totalPrice := math.Abs(detail.Difference * detail.UnitValue * detail.UnitPrice)
					if detail.Difference > 0 {
						incomeTransID := utils.Uuid()
						incomeTrans := models.TransactionModel{
							BaseModel: shared.BaseModel{
								ID: incomeTransID,
							},
							Date:                        date,
							AccountID:                   revenueID,
							Description:                 "Pendapatan Lain-lain / Penyesuaian Persediaan " + stockOpnameHeader.StockOpnameNumber,
							TransactionRefID:            &inventoryTransID,
							TransactionRefType:          "transaction",
							TransactionSecondaryRefID:   &stockOpnameHeader.ID,
							TransactionSecondaryRefType: refType,
							CompanyID:                   stockOpnameHeader.CompanyID,
							Credit:                      totalPrice,
							Amount:                      totalPrice,
							UserID:                      &userID,
						}
						err := tx.Create(&incomeTrans).Error
						if err != nil {
							return err
						}

						inventoryTrans := models.TransactionModel{
							BaseModel: shared.BaseModel{
								ID: inventoryTransID,
							},
							Date:                        date,
							AccountID:                   inventoryID,
							Description:                 "Penyesuaian Stock Opname " + stockOpnameHeader.StockOpnameNumber,
							TransactionRefID:            &inventoryTransID,
							TransactionRefType:          "transaction",
							TransactionSecondaryRefID:   &stockOpnameHeader.ID,
							TransactionSecondaryRefType: refType,
							CompanyID:                   stockOpnameHeader.CompanyID,
							Debit:                       totalPrice,
							Amount:                      totalPrice,
							UserID:                      &userID,
						}
						err = tx.Create(&inventoryTrans).Error
						if err != nil {
							return err
						}

					}
					if detail.Difference < 0 {
						expenseTransID := utils.Uuid()
						expenseTrans := models.TransactionModel{
							BaseModel: shared.BaseModel{
								ID: expenseTransID,
							},
							Date:                        date,
							AccountID:                   expenseID,
							Description:                 "Kerugian Selisih Persediaan " + stockOpnameHeader.StockOpnameNumber,
							TransactionRefID:            &inventoryTransID,
							TransactionRefType:          "transaction",
							TransactionSecondaryRefID:   &stockOpnameHeader.ID,
							TransactionSecondaryRefType: refType,
							CompanyID:                   stockOpnameHeader.CompanyID,
							Debit:                       totalPrice,
							Amount:                      totalPrice,
							UserID:                      &userID,
						}
						err := tx.Create(&expenseTrans).Error
						if err != nil {
							return err
						}

						inventoryTrans := models.TransactionModel{
							BaseModel: shared.BaseModel{
								ID: inventoryTransID,
							},
							Date:                        date,
							AccountID:                   inventoryID,
							Description:                 "Penyesuaian Stock Opname " + stockOpnameHeader.StockOpnameNumber,
							TransactionRefID:            &inventoryTransID,
							TransactionRefType:          "transaction",
							TransactionSecondaryRefID:   &stockOpnameHeader.ID,
							TransactionSecondaryRefType: refType,
							CompanyID:                   stockOpnameHeader.CompanyID,
							Credit:                      totalPrice,
							Amount:                      totalPrice,
							UserID:                      &userID,
						}
						err = tx.Create(&inventoryTrans).Error
						if err != nil {
							return err
						}
					}
				}

			}
		}
		// Update status stock opname menjadi "COMPLETED"
		return tx.Model(&stockOpnameHeader).Update("status", models.StatusCompleted).Error

	})
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
