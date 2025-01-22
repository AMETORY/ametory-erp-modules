package pos

import (
	"errors"
	"fmt"
	"time"

	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/finance"
	"github.com/AMETORY/ametory-erp-modules/inventory"
	"github.com/AMETORY/ametory-erp-modules/shared/models"
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

func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(&models.POSModel{}, &models.POSSalesItemModel{})
}

// CreateMerchant membuat merchant baru
func (s *POSService) CreateMerchant(name, address, phone string) (*models.MerchantModel, error) {
	merchant := models.MerchantModel{
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
func (s *POSService) CreatePOSTransaction(merchantID *string, contactID, warehouseID string, items []models.POSSalesItemModel, description string) (*models.POSModel, error) {
	invSrv, ok := s.ctx.InventoryService.(*inventory.InventoryService)
	if !ok {
		return nil, errors.New("invalid inventory service")
	}

	// Hitung total harga transaksi
	var totalPrice float64
	for _, item := range items {
		totalPrice += item.Total
	}
	if merchantID == nil {
		return nil, errors.New("no merchant")
	}

	merchant := models.MerchantModel{}
	if err := s.db.Where("id = ?", merchantID).First(&merchant).Error; err != nil {
		return nil, err
	}
	pos := models.POSModel{
		MerchantID: merchantID,
		ContactID:  contactID,
		Total:      totalPrice,
		Status:     "pending",
		Items:      items,
	}

	now := time.Now()

	err := s.ctx.DB.Transaction(func(tx *gorm.DB) error {

		// Simpan transaksi POS ke database
		if err := tx.Create(&pos).Error; err != nil {
			tx.Rollback()
			return err
		}

		// Kurangi stok untuk setiap item
		for _, item := range items {
			_, err := invSrv.StockMovementService.AddMovement(now, *item.ProductID, warehouseID, item.VariantID, merchantID, nil, -item.Quantity, models.MovementTypeOut, pos.ID, description)
			if err != nil {
				return err
			}
		}

		// Update status transaksi menjadi "completed"
		pos.Status = "completed"
		if err := tx.Save(&pos).Error; err != nil {
			tx.Rollback()
			return err
		}

		// Tambahkan transaksi ke jurnal
		if pos.SaleAccountID != nil {
			if err := s.financeService.TransactionService.CreateTransaction(&models.TransactionModel{
				Date:               now,
				AccountID:          pos.SaleAccountID,
				Description:        fmt.Sprintf("Penjualan [%s] %s ", merchant.Name, pos.SalesNumber),
				Notes:              pos.Description,
				TransactionRefID:   &pos.ID,
				TransactionRefType: "pos_sales",
				CompanyID:          pos.CompanyID,
			}, totalPrice); err != nil {
				tx.Rollback()
				return err
			}
		}
		if pos.AssetAccountID != nil {
			if err := s.financeService.TransactionService.CreateTransaction(&models.TransactionModel{
				Date:               now,
				AccountID:          pos.AssetAccountID,
				Description:        fmt.Sprintf("Penjualan [%s] %s ", merchant.Name, pos.SalesNumber),
				Notes:              pos.Description,
				TransactionRefID:   &pos.ID,
				TransactionRefType: "pos_sales",
				CompanyID:          pos.CompanyID,
			}, totalPrice); err != nil {
				tx.Rollback()
				return err
			}
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
func (s *POSService) GetTransactionsByMerchant(merchantID uint) ([]models.POSModel, error) {
	var transactions []models.POSModel
	if err := s.db.Preload("Items").Where("merchant_id = ?", merchantID).Find(&transactions).Error; err != nil {
		return nil, err
	}
	return transactions, nil
}
