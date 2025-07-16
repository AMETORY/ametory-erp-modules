package stock_opname

import (
	"errors"
	"math"
	"net/http"
	"time"

	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/inventory/product"
	stockmovement "github.com/AMETORY/ametory-erp-modules/inventory/stock_movement"
	"github.com/AMETORY/ametory-erp-modules/shared"
	"github.com/AMETORY/ametory-erp-modules/shared/models"
	"github.com/AMETORY/ametory-erp-modules/utils"
	"github.com/morkid/paginate"
	"gorm.io/gorm"
)

type StockOpnameService struct {
	db                   *gorm.DB
	ctx                  *context.ERPContext
	productService       *product.ProductService
	stockMovementService *stockmovement.StockMovementService
}

// NewStockOpnameService creates a new instance of StockOpnameService with the given database connection, context,
// product service and stock movement service.
func NewStockOpnameService(db *gorm.DB, ctx *context.ERPContext, productService *product.ProductService, stockMovementService *stockmovement.StockMovementService) *StockOpnameService {
	return &StockOpnameService{db: db, ctx: ctx, productService: productService, stockMovementService: stockMovementService}
}

// Migrate performs the database schema migration for StockOpnameHeader and StockOpnameDetail models.
func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(&models.StockOpnameHeader{}, &models.StockOpnameDetail{})
}

// CreateStockOpnameFromHeader creates a new stock opname header with the given data and returns an error if the
// creation is unsuccessful.
func (s *StockOpnameService) CreateStockOpnameFromHeader(data *models.StockOpnameHeader) error {
	return s.db.Create(data).Error
}

// UpdateStockOpname updates the stock opname header with the given ID with the given data.
// An error is returned if the update is unsuccessful.
func (s *StockOpnameService) UpdateStockOpname(stockOpnameID string, data *models.StockOpnameHeader) error {
	return s.db.Model(&models.StockOpnameHeader{}).
		Where("id = ?", stockOpnameID).
		Updates(data).
		Error
}

// GetStockOpnameByID retrieves a stock opname header by its ID. The function returns the stock opname header
// with its associated warehouse and created by user, as well as its details with associated products.
// An error is returned if the retrieval is unsuccessful.
func (s *StockOpnameService) GetStockOpnameByID(stockOpnameID string) (*models.StockOpnameHeader, error) {
	var stockOpnameHeader models.StockOpnameHeader
	if err := s.db.
		Preload("Warehouse").
		Preload("CreatedBy").
		Preload("Details.Product").
		First(&stockOpnameHeader, "id = ?", stockOpnameID).Error; err != nil {
		return nil, err
	}
	return &stockOpnameHeader, nil
}

// GetStockOpnames retrieves a paginated list of stock opnames from the database.
//
// It takes an HTTP request and a search query string as input. The function uses
// GORM to query the database for stock opnames, applying the search query to the
// stock opname number and notes fields. If the request contains a company ID
// header, the method also filters the result by the company ID or includes entries
// with a null company ID. The function utilizes pagination to manage the result set
// and applies any necessary request modifications using the utils.FixRequest utility.
//
// The function returns a paginated page of StockOpnameHeader and an error if
// the operation fails.
func (s *StockOpnameService) GetStockOpnames(request http.Request, search string) (paginate.Page, error) {
	pg := paginate.New()
	stmt := s.db
	if search != "" {
		stmt = stmt.Where("stock_opname_number ILIKE ? OR notes ILIKE ?",
			"%"+search+"%",
			"%"+search+"%",
		)
	}
	if request.Header.Get("ID-Company") != "" {
		stmt = stmt.Where("company_id = ? or company_id is null", request.Header.Get("ID-Company"))
	}
	request.URL.Query().Get("page")
	stmt = stmt.Model(&models.StockOpnameHeader{})
	utils.FixRequest(&request)
	page := pg.With(stmt).Request(request).Response(&[]models.StockOpnameHeader{})
	page.Page = page.Page + 1
	return page, nil
}

// AddItem adds a new item to the stock opname with the given ID.
//
// The function first retrieves the stock opname header with the given ID from the database.
// If the retrieval is unsuccessful, the function returns an error.
// The function then populates the stock opname detail with the given data and the retrieved
// stock opname header's ID. It also retrieves the product's current stock quantity from the
// product service and populates the stock opname detail with it.
// Finally, the function creates a new stock opname detail in the database and returns an error
// if the creation is unsuccessful.
func (s *StockOpnameService) AddItem(stockOpnameID string, data *models.StockOpnameDetail) error {
	var stockOpnameHeader models.StockOpnameHeader
	if err := s.db.First(&stockOpnameHeader, "id = ?", stockOpnameID).Error; err != nil {
		return err
	}
	data.StockOpnameID = stockOpnameID
	systemQty, err := s.productService.GetStock(data.ProductID, nil, &stockOpnameHeader.WarehouseID)
	if err != nil {
		return err
	}

	data.SystemQty = systemQty

	return s.db.Debug().Create(&data).Error
}

// UpdateItem updates an existing stock opname detail with the specified ID using the provided data.
//
// The function first retrieves the stock opname detail with the given ID from the database.
// If the retrieval is unsuccessful, it returns an error.
// It then updates the stock opname detail with the new data provided.
// An error is returned if the update operation in the database fails.
func (s *StockOpnameService) UpdateItem(stockOpnameDetailID string, data *models.StockOpnameDetail) error {
	var stockOpnameDetail models.StockOpnameDetail
	if err := s.db.First(&stockOpnameDetail, "id = ?", stockOpnameDetailID).Error; err != nil {
		return err
	}
	if err := s.db.Model(&models.StockOpnameDetail{}).
		Where("id = ?", stockOpnameDetailID).
		Updates(data).
		Error; err != nil {
		return err
	}
	return nil
}

// DeleteItem deletes the stock opname detail with the given ID.
//
// The function performs the deletion in a database transaction.
// If the deletion is unsuccessful, it returns an error.
func (s *StockOpnameService) DeleteItem(stockOpnameDetailID string) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		// Delete stock opname detail
		if err := tx.Where("id = ?", stockOpnameDetailID).Unscoped().Delete(&models.StockOpnameDetail{}).Error; err != nil {
			return err
		}
		return nil
	})
}

// DeleteStockOpname deletes a stock opname and its related records from the database.
//
// This function deletes the stock opname details, related stock movements, and the stock
// opname header associated with the given stockOpnameID. It also deletes related inventory
// transactions unless the skipTransaction flag is set to true. The operation is performed
// within a database transaction, ensuring that all deletions are completed successfully or
// none at all.
//
// Args:
//   - stockOpnameID: the ID of the stock opname to be deleted.
//   - skipTransaction: a boolean flag indicating whether to skip the deletion of related
//     inventory transactions.
//
// Returns:
//   - An error if any deletion operation fails.
func (s *StockOpnameService) DeleteStockOpname(stockOpnameID string, skipTransaction bool) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		// Delete stock opname detail
		if err := tx.Where("stock_opname_id = ?", stockOpnameID).Delete(&models.StockOpnameDetail{}).Error; err != nil {
			return err
		}

		// Delete related inventory transactions
		if err := tx.Where("reference_id = ?", stockOpnameID).Delete(&models.StockMovementModel{}).Error; err != nil {
			return err
		}

		if !skipTransaction {
			// Delete related inventory transactions
			if err := tx.Where("transaction_secondary_ref_id = ?", stockOpnameID).Delete(&models.TransactionModel{}).Error; err != nil {
				return err
			}
		}

		// Delete stock opname header
		return tx.Where("id = ?", stockOpnameID).Delete(&models.StockOpnameHeader{}).Error
	})
}

// CreateStockOpname creates a new stock opname header with the given warehouse ID and notes.
// It also creates stock opname details for each product in the given list, using the
// current system stock quantity from the product service. The function first creates
// the stock opname header in the database, then creates the stock opname details.
// If any error occurs during the creation, the function returns an error.
//
// Args:
//   - warehouseID: the ID of the warehouse for the stock opname.
//   - products: a list of products to be included in the stock opname.
//   - notes: an optional notes field for the stock opname.
//
// Returns:
//   - A pointer to the newly created stock opname header, or an error if the
//     creation fails.
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

// CompleteStockOpname completes a stock opname with the given ID. The function
// updates the product stock quantities and creates stock movement records for
// each product with a difference between the counted quantity and the system
// quantity. The function also creates journal entries for the stock opname if
// the inventoryID parameter is not nil. The function returns an error if any
// error occurs during the process.
//
// Args:
//   - stockOpnameID: the ID of the stock opname to be completed.
//   - date: the date of the stock opname.
//   - userID: the ID of the user who is completing the stock opname.
//   - inventoryID: the ID of the inventory account to be used for the journal
//     entries. If nil, the function will not create journal entries.
//
// Returns:
//   - An error if any error occurs during the process.
func (s *StockOpnameService) CompleteStockOpname(stockOpnameID string, date time.Time, userID string, inventoryID *string) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		var stockOpnameHeader models.StockOpnameHeader
		if err := tx.Preload("Details").First(&stockOpnameHeader, "id = ?", stockOpnameID).Error; err != nil {
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
				movement.CompanyID = stockOpnameHeader.CompanyID

				err = tx.Save(movement).Error
				if err != nil {
					return err
				}

				if inventoryID != nil {

					inventoryTransID := utils.Uuid()
					totalPrice := math.Abs(detail.Difference * detail.UnitValue * detail.UnitPrice)
					code := utils.RandString(8, false)
					if detail.Difference > 0 {
						var stockOpnameAccount models.AccountModel
						err := s.db.Where("is_stock_opname_account = ? and company_id = ? and type = ?", true, *stockOpnameHeader.CompanyID, models.REVENUE).First(&stockOpnameAccount).Error
						if err != nil {
							return err
						}

						incomeTransID := utils.Uuid()
						incomeTrans := models.TransactionModel{
							BaseModel: shared.BaseModel{
								ID: incomeTransID,
							},
							Code:                        code,
							Date:                        date,
							AccountID:                   &stockOpnameAccount.ID,
							Description:                 "Pendapatan Lain-lain / Penyesuaian Persediaan " + stockOpnameHeader.StockOpnameNumber,
							Notes:                       detail.Notes,
							TransactionRefID:            &inventoryTransID,
							TransactionRefType:          "transaction",
							TransactionSecondaryRefID:   &stockOpnameHeader.ID,
							TransactionSecondaryRefType: refType,
							CompanyID:                   stockOpnameHeader.CompanyID,
							Credit:                      totalPrice,
							Amount:                      totalPrice,
							UserID:                      &userID,
						}
						err = tx.Create(&incomeTrans).Error
						if err != nil {
							return err
						}

						inventoryTrans := models.TransactionModel{
							BaseModel: shared.BaseModel{
								ID: inventoryTransID,
							},
							Code:                        code,
							Date:                        date,
							AccountID:                   inventoryID,
							Description:                 "Penyesuaian Stock Opname " + stockOpnameHeader.StockOpnameNumber,
							Notes:                       detail.Notes,
							TransactionRefID:            &incomeTransID,
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
						var stockOpnameAccount models.AccountModel
						err := s.db.Where("is_stock_opname_account = ? and company_id = ? and type = ?", true, *stockOpnameHeader.CompanyID, models.EXPENSE).First(&stockOpnameAccount).Error
						if err != nil {
							return err
						}

						expenseTransID := utils.Uuid()
						expenseTrans := models.TransactionModel{
							BaseModel: shared.BaseModel{
								ID: expenseTransID,
							},
							Code:                        code,
							Date:                        date,
							AccountID:                   &stockOpnameAccount.ID,
							Description:                 "Kerugian Selisih Persediaan " + stockOpnameHeader.StockOpnameNumber,
							Notes:                       detail.Notes,
							TransactionRefID:            &inventoryTransID,
							TransactionRefType:          "transaction",
							TransactionSecondaryRefID:   &stockOpnameHeader.ID,
							TransactionSecondaryRefType: refType,
							CompanyID:                   stockOpnameHeader.CompanyID,
							Debit:                       totalPrice,
							Amount:                      totalPrice,
							UserID:                      &userID,
						}
						err = tx.Create(&expenseTrans).Error
						if err != nil {
							return err
						}

						inventoryTrans := models.TransactionModel{
							BaseModel: shared.BaseModel{
								ID: inventoryTransID,
							},
							Code:                        code,
							Date:                        date,
							AccountID:                   inventoryID,
							Description:                 "Penyesuaian Stock Opname " + stockOpnameHeader.StockOpnameNumber,
							Notes:                       detail.Notes,
							TransactionRefID:            &expenseTransID,
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

// GenerateDiscrepancyReport generates a discrepancy report for a given stock opname ID.
//
// The discrepancy report is a list of products with their physical quantity, system quantity, difference, and notes.
// The function first joins the stock_opname_details table with the products table and then selects the required columns.
// The function then scans the results into a slice of StockDiscrepancyReport and returns it along with an error if any.
func (s *StockOpnameService) GenerateDiscrepancyReport(stockOpnameID string) ([]StockDiscrepancyReport, error) {
	var report []StockDiscrepancyReport
	err := s.db.Table("stock_opname_details").
		Joins("JOIN products ON stock_opname_details.product_id = products.id").
		Where("stock_opname_details.stock_opname_id = ?", stockOpnameID).
		Select("stock_opname_details.product_id, products.name as product_name, stock_opname_details.quantity as physical_qty, stock_opname_details.system_qty, stock_opname_details.difference, stock_opname_details.notes").
		Scan(&report).Error
	return report, err
}
