package pos

import (
	"errors"

	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/finance"
	"github.com/AMETORY/ametory-erp-modules/inventory"
	stockmovement "github.com/AMETORY/ametory-erp-modules/inventory/stock_movement"
	"gorm.io/gorm"
)

type POSService struct {
	ctx            *context.ERPContext
	db             *gorm.DB
	financeService *finance.FinanceService
}

func NewPOSService(db *gorm.DB, ctx *context.ERPContext, financeService *finance.FinanceService) *POSService {
	return &POSService{
		db:             db,
		ctx:            ctx,
		financeService: financeService,
	}
}

// CreateMerchant membuat merchant baru
func (s *POSService) CreateMerchant(name, address, phone string) (*MerchantModel, error) {
	merchant := MerchantModel{
		Name:    name,
		Address: address,
		Phone:   phone,
	}

	if err := s.db.Create(&merchant).Error; err != nil {
		return nil, err
	}

	return &merchant, nil
}

// CreatePOSTransaction membuat transaksi POS baru dengan multi-item
func (s *POSService) CreatePOSTransaction(merchantID *string, contactID, warehouseID string, items []POSSalesItemModel) (*POSModel, error) {
	invSrv, ok := s.ctx.InventoryService.(*inventory.InventoryService)
	if !ok {
		return nil, errors.New("invalid inventory service")
	}

	// Hitung total harga transaksi
	var totalPrice float64
	for _, item := range items {
		totalPrice += item.Total
	}
	pos := POSModel{
		MerchantID: merchantID,
		ContactID:  contactID,
		Total:      totalPrice,
		Status:     "pending",
		Items:      items,
	}

	err := s.ctx.DB.Transaction(func(tx *gorm.DB) error {

		// Simpan transaksi POS ke database
		if err := tx.Create(&pos).Error; err != nil {
			tx.Rollback()
			return err
		}

		// Kurangi stok untuk setiap item
		for _, item := range items {
			if err := invSrv.StockMovementService.AddMovement(*item.ProductID, warehouseID, merchantID, -item.Quantity, stockmovement.MovementTypeOut, pos.ID); err != nil {
				return err
			}
		}

		// Update status transaksi menjadi "completed"
		pos.Status = "completed"
		if err := tx.Save(&pos).Error; err != nil {
			tx.Rollback()
			return err
		}

		if err := tx.Commit().Error; err != nil {
			tx.Rollback()
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &pos, nil
}

// GetTransactionsByMerchant mengambil semua transaksi POS berdasarkan merchant
func (s *POSService) GetTransactionsByMerchant(merchantID uint) ([]POSModel, error) {
	var transactions []POSModel
	if err := s.db.Preload("Items").Where("merchant_id = ?", merchantID).Find(&transactions).Error; err != nil {
		return nil, err
	}
	return transactions, nil
}
