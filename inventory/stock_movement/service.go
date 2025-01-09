package stockmovement

import (
	"net/http"

	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/utils"
	"github.com/google/uuid"
	"github.com/morkid/paginate"
	"gorm.io/gorm"
)

type StockMovementService struct {
	db  *gorm.DB
	ctx *context.ERPContext
}

func NewStockMovementService(db *gorm.DB, ctx *context.ERPContext) *StockMovementService {
	return &StockMovementService{db: db, ctx: ctx}
}

// CreateStockMovement membuat pergerakan stok
func (s *StockMovementService) CreateStockMovement(movement *StockMovementModel) error {
	return s.db.Create(movement).Error
}

func (s *StockMovementService) GetStockMovements(request http.Request, search string) (paginate.Page, error) {
	pg := paginate.New()
	stmt := s.db.Joins("LEFT JOIN products ON stock_movements.product_id = products.id")

	if search != "" {
		stmt = stmt.Where("stock_movements.description ILIKE ? OR products.name ILIKE ?",
			"%"+search+"%",
			"%"+search+"%",
		)
	}
	if request.Header.Get("ID-Company") != "" {
		stmt = stmt.Where("company_id = ?", request.Header.Get("ID-Company"))
	}
	if request.Header.Get("ID-Distributor") != "" {
		stmt = stmt.Where("stock_movements.distributor = ?", request.Header.Get("ID-Distributor"))
	}
	request.URL.Query().Get("page")
	stmt = stmt.Model(&StockMovementModel{})
	utils.FixRequest(&request)
	page := pg.With(stmt).Request(request).Response(&[]StockMovementModel{})
	page.Page = page.Page + 1
	return page, nil
}

// AddMovement menambahkan pergerakan stok
func (s *StockMovementService) AddMovement(productID, warehouseID string, merchantID *string, distributorID *string, quantity float64, movementType MovementType, referenceID, description string) error {
	movement := StockMovementModel{
		ProductID:     productID,
		WarehouseID:   warehouseID,
		Quantity:      quantity,
		Type:          movementType,
		MerchantID:    merchantID,
		DistributorID: distributorID,
		ReferenceID:   referenceID,
		Description:   description,
	}

	if err := s.db.Create(&movement).Error; err != nil {
		return err
	}

	return nil
}

// GetCurrentStock menghitung stok saat ini berdasarkan riwayat pergerakan
func (s *StockMovementService) GetCurrentStock(productID, warehouseID uint) (float64, error) {
	var totalStock float64
	if err := s.db.Model(&StockMovementModel{}).
		Where("product_id = ? AND warehouse_id = ?", productID, warehouseID).
		Select("COALESCE(SUM(quantity), 0)").
		Scan(&totalStock).Error; err != nil {
		return 0, err
	}

	return totalStock, nil
}

// GetMovementHistory mengambil riwayat pergerakan stok
func (s *StockMovementService) GetMovementHistory(productID, warehouseID string) ([]StockMovementModel, error) {
	var movements []StockMovementModel
	if err := s.db.Where("product_id = ? AND warehouse_id = ?", productID, warehouseID).
		Find(&movements).Error; err != nil {
		return nil, err
	}

	return movements, nil
}

// GetMovementByProductID mengambil pergerakan stok berdasarkan ID produk
func (s *StockMovementService) GetMovementByProductID(productID string) ([]StockMovementModel, error) {
	var movements []StockMovementModel
	if err := s.db.Where("product_id = ?", productID).
		Find(&movements).Error; err != nil {
		return nil, err
	}

	return movements, nil
}

// GetMovementByWarehouseID mengambil pergerakan stok berdasarkan ID gudang
func (s *StockMovementService) GetMovementByWarehouseID(warehouseID string) ([]StockMovementModel, error) {
	var movements []StockMovementModel
	if err := s.db.Where("warehouse_id = ?", warehouseID).
		Find(&movements).Error; err != nil {
		return nil, err
	}

	return movements, nil
}

// CreateAdjustment menambahkan pergerakan stok dengan tipe ADJUST
func (s *StockMovementService) CreateAdjustment(productID, warehouseID string, merchantID *string, distributorID *string, quantity float64, referenceID, description string) error {
	return s.AddMovement(productID, warehouseID, merchantID, distributorID, quantity, MovementTypeAdjust, referenceID, description)
}

// TransferStock melakukan transfer stok dari gudang sumber ke gudang tujuan
func (s *StockMovementService) TransferStock(sourceWarehouseID, destinationWarehouseID string, productID string, quantity float64, description string) error {
	if err := s.db.Transaction(func(tx *gorm.DB) error {
		// membuat pergerakan stok di gudang sumber
		if err := s.AddMovement(productID, sourceWarehouseID, nil, nil, -quantity, MovementTypeTransfer, uuid.New().String(), description); err != nil {
			return err
		}

		// membuat pergerakan stok di gudang tujuan
		if err := s.AddMovement(productID, destinationWarehouseID, nil, nil, quantity, MovementTypeTransfer, uuid.New().String(), description); err != nil {
			return err
		}

		return nil
	}); err != nil {
		return err
	}

	return nil
}
