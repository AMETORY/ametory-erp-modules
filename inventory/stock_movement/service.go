package stockmovement

import (
	"net/http"
	"time"

	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/shared/models"
	"github.com/AMETORY/ametory-erp-modules/utils"
	"github.com/morkid/paginate"
	"gorm.io/gorm"
)

type StockMovementService struct {
	db             *gorm.DB
	ctx            *context.ERPContext
	isMerchantMode bool
}

// NewStockMovementService creates a new instance of StockMovementService
// with the provided database connection and ERP context.
func NewStockMovementService(db *gorm.DB, ctx *context.ERPContext) *StockMovementService {
	return &StockMovementService{db: db, ctx: ctx}
}

// Migrate migrates the database schema needed for the StockMovementService.
//
// It uses GORM's AutoMigrate function to create the tables for StockMovementModel
// if they do not already exist.
//
// If the migration fails, the error is returned to the caller.
func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(&models.StockMovementModel{})
}
func (s *StockMovementService) SetDB(db *gorm.DB) {
	s.db = db
}

// CreateStockMovement creates a new stock movement in the database.
//
// The stock movement is created using GORM's Create method.
//
// If the creation fails, the error is returned to the caller.
func (s *StockMovementService) CreateStockMovement(movement *models.StockMovementModel) error {
	return s.db.Create(movement).Error
}

// SetMerchantMode sets the merchant mode status for the StockMovementService.
//
// This function takes a boolean parameter that determines whether the service
// should operate in merchant mode. When set to true, the service applies
// specific logic and constraints related to merchant operations. Otherwise,
// it operates in a general mode.
//
// Parameters:
//
//	isMerchantMode - a boolean indicating whether to enable merchant mode.
func (s *StockMovementService) SetMerchantMode(isMerchantMode bool) {
	s.isMerchantMode = isMerchantMode
}

// GetStockMovements retrieves a paginated list of stock movements from the database.
//
// It takes an HTTP request and a search query string as input. The function uses
// GORM to query the database for stock movements, applying the search query to
// the stock movement description and product name fields. If the request contains
// a company ID header and merchant mode is not enabled, the method filters the
// results by the company ID. Additional filters can be applied based on distributor,
// product, warehouse, and merchant IDs from the request parameters. When in merchant
// mode, it filters results by merchant ID from the request header.
//
// The function utilizes pagination to manage the result set and applies any
// necessary request modifications using the utils.FixRequest utility. It also
// allows sorting by a specified field and order. The function returns a paginated
// page of StockMovementModel and an error if the operation fails.
func (s *StockMovementService) GetStockMovements(request http.Request, search string) (paginate.Page, error) {
	pg := paginate.New()
	stmt := s.db.Preload("Unit", func(db *gorm.DB) *gorm.DB {
		return db.Select("id", "name", "code")
	}).Preload("Merchant", func(db *gorm.DB) *gorm.DB {
		return db.Select("id", "name")
	}).Preload("Product", func(db *gorm.DB) *gorm.DB {
		return db.Select("id", "name", "display_name")
	}).Preload("Warehouse", func(db *gorm.DB) *gorm.DB {
		return db.Select("id", "name")
	}).Joins("LEFT JOIN products ON stock_movements.product_id = products.id")

	if search != "" {
		stmt = stmt.Where("stock_movements.description ILIKE ? OR products.name ILIKE ?",
			"%"+search+"%",
			"%"+search+"%",
		)
	}
	if request.Header.Get("ID-Company") != "" && !s.isMerchantMode {
		if request.Header.Get("ID-Company") == "nil" || request.Header.Get("ID-Company") == "null" {
			stmt = stmt.Where("stock_movements.company_id is null")
		} else {
			stmt = stmt.Where("stock_movements.company_id = ?", request.Header.Get("ID-Company"))

		}
	}
	if request.Header.Get("ID-Distributor") != "" {
		stmt = stmt.Where("stock_movements.distributor = ?", request.Header.Get("ID-Distributor"))
	}
	if request.URL.Query().Get("product_id") != "" {
		stmt = stmt.Where("stock_movements.product_id = ?", request.URL.Query().Get("product_id"))
	}
	if request.URL.Query().Get("warehouse_id") != "" {
		stmt = stmt.Where("stock_movements.warehouse_id = ?", request.URL.Query().Get("warehouse_id"))
	}
	if request.URL.Query().Get("merchant_id") != "" {
		stmt = stmt.Where("stock_movements.merchant_id = ?", request.URL.Query().Get("merchant_id"))
	}
	if s.isMerchantMode {
		stmt = stmt.Where("stock_movements.merchant_id = ?", request.Header.Get("ID-Merchant"))
	}
	request.URL.Query().Get("page")
	stmt = stmt.Model(&models.StockMovementModel{})

	orderBy := request.URL.Query().Get("order_by")
	order := request.URL.Query().Get("order")
	if orderBy == "" {
		orderBy = "created_at"
	}
	if order == "" {
		order = "desc"
	}
	stmt = stmt.Order(orderBy + " " + order)
	utils.FixRequest(&request)
	page := pg.With(stmt).Request(request).Response(&[]models.StockMovementModel{})
	page.Page = page.Page + 1

	items := page.Items.(*[]models.StockMovementModel)
	newItems := make([]models.StockMovementModel, 0)
	for _, item := range *items {
		if item.ReferenceID != "" {
			if item.ReferenceType != nil {
				if *item.ReferenceType == "sales" {
					var salesRef models.SalesModel
					err := s.db.Where("id = ?", item.ReferenceID).First(&salesRef).Error
					if err == nil {
						item.SalesRef = &salesRef
					}
				}
				if *item.ReferenceType == "purchase" {
					var purchaseRef models.PurchaseOrderModel
					err := s.db.Where("id = ?", item.ReferenceID).First(&purchaseRef).Error
					if err == nil {
						item.PurchaseRef = &purchaseRef
					}
				}
				if *item.ReferenceType == "return_purchase" {
					var returnRef models.ReturnModel
					err := s.db.Where("id = ?", item.ReferenceID).First(&returnRef).Error
					if err == nil {
						item.ReturnRef = &returnRef
					}
				}
				if *item.ReferenceType == "stock_opname" {
					var stockOpnameRef models.StockOpnameHeader
					err := s.db.Where("id = ?", item.ReferenceID).First(&stockOpnameRef).Error
					if err == nil {
						item.StockOpnameRef = &stockOpnameRef
					}
				}
			}
		}
		newItems = append(newItems, item)
	}
	page.Items = &newItems
	return page, nil
}

// UpdateLastStock updates the last stock quantity of a product in a merchant.
//
// Args:
//
//	productID: the ID of the product.
//	merchant_model_id: the ID of the merchant.
//	quantity: the new quantity of the last stock.
//	lastUpdatedAt: the time when the last stock was updated.
//
// Returns:
//
//	an error if the update fails.
func (s *StockMovementService) UpdateLastStock(productID, merchant_model_id string, quantity float64, lastUpdatedAt time.Time) error {
	return s.db.Model(&models.ProductMerchant{}).
		Where("product_model_id = ? AND merchant_model_id = ?", productID, merchant_model_id).
		Order("created_at desc").
		Limit(1).
		Updates(map[string]interface{}{"last_stock": quantity, "last_updated_stock": lastUpdatedAt}).Error
}

// UpdateVariantLastStock updates the last stock quantity and last updated time
// of a variant for a specific merchant.
//
// Args:
//    varianID: the ID of the variant.
//    merchantID: the ID of the merchant.
//    quantity: the new quantity of the last stock.
//    lastUpdatedAt: the time when the last stock was updated.
//
// Returns:
//    an error if the update fails.

func (s *StockMovementService) UpdateVariantLastStock(varianID, merchantID string, quantity float64, lastUpdatedAt time.Time) error {
	return s.db.Model(&models.VarianMerchant{}).
		Where("variant_id = ? AND merchant_id = ?", varianID, merchantID).
		Order("created_at desc").
		Limit(1).
		Updates(map[string]interface{}{"last_stock": quantity, "last_updated_stock": lastUpdatedAt}).Error
}

// AddMovement creates a new stock movement record in the database.
//
// The function takes various parameters to define the stock movement, including
// the date, product ID, warehouse ID, optional variant and merchant IDs, optional
// distributor and company IDs, quantity, movement type, reference ID, and
// description. It constructs a StockMovementModel with these details and saves
// it to the database using GORM's Create method.
//
// Parameters:
//   - date: the date of the stock movement.
//   - productID: the ID of the product involved in the movement.
//   - warehouseID: the ID of the warehouse where the movement occurs.
//   - variantID: an optional ID of the product variant.
//   - merchantID: an optional ID of the merchant involved.
//   - distributorID: an optional ID of the distributor involved.
//   - companyID: an optional ID of the company involved.
//   - quantity: the quantity of stock being moved.
//   - movementType: the type of stock movement (e.g., IN, OUT, TRANSFER, etc.).
//   - referenceID: an optional ID for referencing related records.
//   - description: a brief description of the stock movement.
//
// Returns:
//   - A pointer to the newly created StockMovementModel, or an error if the creation fails.
func (s *StockMovementService) AddMovement(date time.Time, productID, warehouseID string, variantID, merchantID *string, distributorID, companyID *string, quantity float64, movementType models.MovementType, referenceID, description string) (*models.StockMovementModel, error) {
	movement := models.StockMovementModel{
		Date:          date,
		ProductID:     productID,
		VariantID:     variantID,
		WarehouseID:   warehouseID,
		Quantity:      quantity,
		Type:          movementType,
		MerchantID:    merchantID,
		DistributorID: distributorID,
		CompanyID:     companyID,
		ReferenceID:   referenceID,
		Description:   description,
	}

	if err := s.db.Create(&movement).Error; err != nil {
		return nil, err
	}

	return &movement, nil
}

// GetCurrentStock retrieves the total stock of a product in a warehouse.
//
// Args:
//
//	productID: the ID of the product.
//	warehouseID: the ID of the warehouse.
//
// Returns:
//
//	the total stock quantity of the product if found, and an error if any error occurs.
func (s *StockMovementService) GetCurrentStock(productID, warehouseID string) (float64, error) {
	var totalStock float64
	if err := s.db.Model(&models.StockMovementModel{}).
		Where("product_id = ? AND warehouse_id = ?", productID, warehouseID).
		Select("COALESCE(SUM(quantity), 0)").
		Scan(&totalStock).Error; err != nil {
		return 0, err
	}

	return totalStock, nil
}

// GetCurrentStockByMerchantID retrieves the total stock of a product for a specific merchant.
//
// Args:
//
//	productID: the ID of the product.
//	merchantID: the ID of the merchant.
//
// Returns:
//
//	the total stock quantity of the product if found, and an error if any error occurs.
func (s *StockMovementService) GetCurrentStockByMerchantID(productID, merchantID string) (float64, error) {
	var totalStock float64
	if err := s.db.Model(&models.StockMovementModel{}).
		Where("product_id = ? AND merchant_id = ?", productID, merchantID).
		Select("COALESCE(SUM(quantity), 0)").
		Scan(&totalStock).Error; err != nil {
		return 0, err
	}

	return totalStock, nil
}

// GetVarianCurrentStock retrieves the total stock of a product variant in a warehouse.
//
// Args:
//
//	productID: the ID of the product.
//	varianID: the ID of the variant.
//	warehouseID: the ID of the warehouse.
//
// Returns:
//
//	the total stock quantity of the product variant if found, and an error if any error occurs.
func (s *StockMovementService) GetVarianCurrentStock(productID, varianID, warehouseID string) (float64, error) {
	var totalStock float64
	if err := s.db.Model(&models.StockMovementModel{}).
		Where("product_id = ? AND variant_id = ? AND warehouse_id = ?", productID, varianID, warehouseID).
		Select("COALESCE(SUM(quantity), 0)").
		Scan(&totalStock).Error; err != nil {
		return 0, err
	}

	return totalStock, nil
}

// GetVarianCurrentStockByMerchantID retrieves the total stock of a product variant in a merchant.
//
// Args:
//
//	productID: the ID of the product.
//	varianID: the ID of the variant.
//	merchant_id: the ID of the merchant.
//
// Returns:
//
//	the total stock quantity of the product variant if found, and an error if any error occurs.
func (s *StockMovementService) GetVarianCurrentStockByMerchantID(productID, varianID, merchant_id string) (float64, error) {
	var totalStock float64
	if err := s.db.Model(&models.StockMovementModel{}).
		Where("product_id = ? AND variant_id = ? AND merchant_id = ?", productID, varianID, merchant_id).
		Select("COALESCE(SUM(quantity), 0)").
		Scan(&totalStock).Error; err != nil {
		return 0, err
	}

	return totalStock, nil
}

// GetMovementHistory retrieves the stock movement history of a product in a warehouse.
//
// Args:
//   - productID: the ID of the product.
//   - warehouseID: the ID of the warehouse.
//
// Returns:
//   - a slice of StockMovementModel containing the stock movement history of the product
//     if found, and an error if any error occurs.
func (s *StockMovementService) GetMovementHistory(productID, warehouseID string) ([]models.StockMovementModel, error) {
	var movements []models.StockMovementModel
	if err := s.db.Where("product_id = ? AND warehouse_id = ?", productID, warehouseID).
		Find(&movements).Error; err != nil {
		return nil, err
	}

	return movements, nil
}

// GetMovementByProductID retrieves the stock movement history of a product.
//
// Args:
//   - productID: the ID of the product.
//
// Returns:
//   - a slice of StockMovementModel containing the stock movement history of the product
//     if found, and an error if any error occurs.
func (s *StockMovementService) GetMovementByProductID(productID string) ([]models.StockMovementModel, error) {
	var movements []models.StockMovementModel
	if err := s.db.Where("product_id = ?", productID).
		Find(&movements).Error; err != nil {
		return nil, err
	}

	return movements, nil
}

// GetMovementByWarehouseID retrieves the stock movement history of a warehouse.
//
// Args:
//   - warehouseID: the ID of the warehouse.
//
// Returns:
//   - a slice of StockMovementModel containing the stock movement history of the warehouse
//     if found, and an error if any error occurs.
func (s *StockMovementService) GetMovementByWarehouseID(warehouseID string) ([]models.StockMovementModel, error) {
	var movements []models.StockMovementModel
	if err := s.db.Where("warehouse_id = ?", warehouseID).
		Find(&movements).Error; err != nil {
		return nil, err
	}

	return movements, nil
}

// CreateAdjustment creates a new stock movement with type ADJUST for a product in a warehouse.
//
// Args:
//   - date: the date of the stock movement.
//   - productID: the ID of the product.
//   - warehouseID: the ID of the warehouse.
//   - variantID: an optional ID of the product variant.
//   - merchantID: an optional ID of the merchant involved.
//   - distributorID: an optional ID of the distributor involved.
//   - companyID: an optional ID of the company involved.
//   - quantity: the quantity of stock being adjusted.
//   - referenceID: an optional ID for referencing related records.
//   - description: a brief description of the stock movement.
//
// Returns:
//   - A pointer to the newly created StockMovementModel, or an error if the creation fails.
func (s *StockMovementService) CreateAdjustment(date time.Time, productID, warehouseID string, variantID, merchantID *string, distributorID, companyID *string, quantity float64, referenceID, description string) (*models.StockMovementModel, error) {
	return s.AddMovement(date, productID, warehouseID, variantID, merchantID, distributorID, companyID, quantity, models.MovementTypeAdjust, referenceID, description)
}

// TransferStock creates two new stock movements with type TRANSFER for a product between two warehouses.
//
// Args:
//   - date: the date of the stock movement.
//   - sourceWarehouseID: the ID of the warehouse where the stock is being transferred from.
//   - destinationWarehouseID: the ID of the warehouse where the stock is being transferred to.
//   - productID: the ID of the product.
//   - variantID: an optional ID of the product variant.
//   - quantity: the quantity of stock being transferred.
//   - description: a brief description of the stock movement.
//
// Returns:
//   - A pointer to the newly created StockMovementModel, or an error if the creation fails.
func (s *StockMovementService) TransferStock(date time.Time, sourceWarehouseID, destinationWarehouseID string, productID string, variantID *string, quantity float64, description string) (*models.StockMovementModel, error) {
	if err := s.db.Transaction(func(tx *gorm.DB) error {
		var movement *models.StockMovementModel
		// membuat pergerakan stok di gudang sumber
		movement, err := s.AddMovement(date, productID, sourceWarehouseID, variantID, nil, nil, nil, -quantity, models.MovementTypeTransfer, "", description)
		if err != nil {
			return err
		}

		// membuat pergerakan stok di gudang tujuan
		movement2, err := s.AddMovement(date, productID, destinationWarehouseID, variantID, nil, nil, nil, quantity, models.MovementTypeTransfer, movement.ID, description)
		if err != nil {
			return err
		}

		movement.ReferenceID = movement2.ID
		if err := tx.Save(movement).Error; err != nil {
			return err
		}

		return nil
	}); err != nil {
		return nil, err
	}

	return nil, nil
}

// UpdateStockMovement updates a stock movement in the database.
//
// The function takes an ID of the stock movement to be updated and a pointer to
// a StockMovementModel containing the updated data. The function uses GORM's
// Updates method to update the stock movement in the database.
//
// Parameters:
//   - id: the ID of the stock movement to be updated.
//   - data: a pointer to a StockMovementModel containing the updated data.
//
// Returns:
//   - an error if the update fails.
func (s *StockMovementService) UpdateStockMovement(id string, data *models.StockMovementModel) error {
	return s.db.Where("id = ?", id).Updates(data).Error
}

// DeleteStockMovement removes a stock movement from the database by its ID.
//
// Args:
//   - id: the ID of the stock movement to be deleted.
//
// Returns:
//   - an error if the deletion fails.
func (s *StockMovementService) DeleteStockMovement(id string) error {
	return s.db.Where("id = ?", id).Delete(&models.StockMovementModel{}).Error
}

// GetStockMovementByID retrieves a stock movement from the database by its ID.
//
// Args:
//   - id: the ID of the stock movement to be retrieved.
//
// Returns:
//   - a pointer to a StockMovementModel containing the retrieved data, or an error if the retrieval fails.
func (s *StockMovementService) GetStockMovementByID(id string) (*models.StockMovementModel, error) {
	var invoice models.StockMovementModel
	err := s.db.Preload("Product", func(db *gorm.DB) *gorm.DB {
		return db.Select("id", "name")
	}).Preload("Warehouse", func(db *gorm.DB) *gorm.DB {
		return db.Select("id", "name")
	}).Where("id = ?", id).First(&invoice).Error
	return &invoice, err
}
