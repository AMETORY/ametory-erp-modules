package stockmovement

import (
	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type StockMovementService struct {
	db  *gorm.DB
	ctx *context.ERPContext
}

func NewStockMovementService(db *gorm.DB, ctx *context.ERPContext) *StockMovementService {
	return &StockMovementService{db: db, ctx: ctx}
}

// AddMovement menambahkan pergerakan stok
func (s *StockMovementService) AddMovement(productID, warehouseID string, quantity float64, movementType MovementType, referenceID string) error {
	movement := StockMovementModel{
		ProductID:   productID,
		WarehouseID: warehouseID,
		Quantity:    quantity,
		Type:        movementType,
		ReferenceID: referenceID,
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
func (s *StockMovementService) CreateAdjustment(productID, warehouseID string, quantity float64, referenceID string) error {
	return s.AddMovement(productID, warehouseID, quantity, MovementTypeAdjust, referenceID)
}

// TransferStock melakukan transfer stok dari gudang sumber ke gudang tujuan
func (s *StockMovementService) TransferStock(sourceWarehouseID, destinationWarehouseID string, productID string, quantity float64) error {
	if err := s.db.Transaction(func(tx *gorm.DB) error {
		// membuat pergerakan stok di gudang sumber
		if err := s.AddMovement(productID, sourceWarehouseID, -quantity, MovementTypeTransfer, uuid.New().String()); err != nil {
			return err
		}

		// membuat pergerakan stok di gudang tujuan
		if err := s.AddMovement(productID, destinationWarehouseID, quantity, MovementTypeTransfer, uuid.New().String()); err != nil {
			return err
		}

		return nil
	}); err != nil {
		return err
	}

	return nil
}
