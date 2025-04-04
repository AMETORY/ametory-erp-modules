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

func NewStockMovementService(db *gorm.DB, ctx *context.ERPContext) *StockMovementService {
	return &StockMovementService{db: db, ctx: ctx}
}

func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(&models.StockMovementModel{})
}
func (s *StockMovementService) SetDB(db *gorm.DB) {
	s.db = db
}

// CreateStockMovement membuat pergerakan stok
func (s *StockMovementService) CreateStockMovement(movement *models.StockMovementModel) error {
	return s.db.Create(movement).Error
}

func (s *StockMovementService) SetMerchantMode(isMerchantMode bool) {
	s.isMerchantMode = isMerchantMode
}

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
	utils.FixRequest(&request)

	orderBy := request.URL.Query().Get("order_by")
	order := request.URL.Query().Get("order")
	if orderBy == "" {
		orderBy = "created_at"
	}
	if order == "" {
		order = "desc"
	}
	stmt = stmt.Order(orderBy + " " + order)

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
			}
		}
		newItems = append(newItems, item)
	}
	page.Items = &newItems
	return page, nil
}

// UpdateLastStock update last stock berdasarkan riwayat pergerakan
func (s *StockMovementService) UpdateLastStock(productID, merchant_model_id string, quantity float64, lastUpdatedAt time.Time) error {
	return s.db.Model(&models.ProductMerchant{}).
		Where("product_model_id = ? AND merchant_model_id = ?", productID, merchant_model_id).
		Order("created_at desc").
		Limit(1).
		Updates(map[string]interface{}{"last_stock": quantity, "last_updated_stock": lastUpdatedAt}).Error
}
func (s *StockMovementService) UpdateVariantLastStock(varianID, merchantID string, quantity float64, lastUpdatedAt time.Time) error {
	return s.db.Model(&models.VarianMerchant{}).
		Where("variant_id = ? AND merchant_id = ?", varianID, merchantID).
		Order("created_at desc").
		Limit(1).
		Updates(map[string]interface{}{"last_stock": quantity, "last_updated_stock": lastUpdatedAt}).Error
}

// AddMovement menambahkan pergerakan stok
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

// GetCurrentStock menghitung stok saat ini berdasarkan riwayat pergerakan
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

// GetMovementHistory mengambil riwayat pergerakan stok
func (s *StockMovementService) GetMovementHistory(productID, warehouseID string) ([]models.StockMovementModel, error) {
	var movements []models.StockMovementModel
	if err := s.db.Where("product_id = ? AND warehouse_id = ?", productID, warehouseID).
		Find(&movements).Error; err != nil {
		return nil, err
	}

	return movements, nil
}

// GetMovementByProductID mengambil pergerakan stok berdasarkan ID produk
func (s *StockMovementService) GetMovementByProductID(productID string) ([]models.StockMovementModel, error) {
	var movements []models.StockMovementModel
	if err := s.db.Where("product_id = ?", productID).
		Find(&movements).Error; err != nil {
		return nil, err
	}

	return movements, nil
}

// GetMovementByWarehouseID mengambil pergerakan stok berdasarkan ID gudang
func (s *StockMovementService) GetMovementByWarehouseID(warehouseID string) ([]models.StockMovementModel, error) {
	var movements []models.StockMovementModel
	if err := s.db.Where("warehouse_id = ?", warehouseID).
		Find(&movements).Error; err != nil {
		return nil, err
	}

	return movements, nil
}

// CreateAdjustment menambahkan pergerakan stok dengan tipe ADJUST
func (s *StockMovementService) CreateAdjustment(date time.Time, productID, warehouseID string, variantID, merchantID *string, distributorID, companyID *string, quantity float64, referenceID, description string) (*models.StockMovementModel, error) {
	return s.AddMovement(date, productID, warehouseID, variantID, merchantID, distributorID, companyID, quantity, models.MovementTypeAdjust, referenceID, description)
}

// TransferStock melakukan transfer stok dari gudang sumber ke gudang tujuan
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

func (s *StockMovementService) UpdateStockMovement(id string, data *models.StockMovementModel) error {
	return s.db.Where("id = ?", id).Updates(data).Error
}

func (s *StockMovementService) DeleteStockMovement(id string) error {
	return s.db.Where("id = ?", id).Delete(&models.StockMovementModel{}).Error
}

func (s *StockMovementService) GetStockMovementByID(id string) (*models.StockMovementModel, error) {
	var invoice models.StockMovementModel
	err := s.db.Preload("Product", func(db *gorm.DB) *gorm.DB {
		return db.Select("id", "name")
	}).Preload("Warehouse", func(db *gorm.DB) *gorm.DB {
		return db.Select("id", "name")
	}).Where("id = ?", id).First(&invoice).Error
	return &invoice, err
}
