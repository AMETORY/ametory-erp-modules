package purchase

import (
	"errors"

	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/inventory"
	stockmovement "github.com/AMETORY/ametory-erp-modules/inventory/stock_movement"
	"gorm.io/gorm"
)

type PurchaseService struct {
	db  *gorm.DB
	ctx *context.ERPContext
}

func NewPurchaseService(db *gorm.DB, ctx *context.ERPContext) *PurchaseService {
	return &PurchaseService{
		db:  db,
		ctx: ctx,
	}
}

// CreatePurchaseOrder membuat purchase order baru
func (s *PurchaseService) CreatePurchaseOrder(data *PurchaseOrderModel) error {
	// Hitung total harga

	return s.db.Create(data).Error
}

// ReceivePurchaseOrder menerima barang dari supplier dan menambah stok
func (s *PurchaseService) ReceivePurchaseOrder(poID, warehouseID string) error {
	var po PurchaseOrderModel
	if err := s.db.First(&po, poID).Error; err != nil {
		return err
	}

	invSrv, ok := s.ctx.InventoryService.(*inventory.InventoryService)
	if !ok {
		return errors.New("invalid inventory service")
	}

	// Pastikan status PO adalah "pending"
	if po.Status != "pending" {
		return errors.New("purchase order already processed")
	}

	err := s.ctx.DB.Transaction(func(tx *gorm.DB) error {
		// do some database operations in the transaction (use 'tx' from this point, not 'db')
		for _, v := range po.Items {
			if v.ProductID == nil || v.WarehouseID == nil {
				continue
			}
			if err := invSrv.StockMovementService.AddMovement(*v.ProductID, *v.WarehouseID, v.Quantity, stockmovement.MovementTypeIn, po.ID); err != nil {
				tx.Rollback()
				return err
			}
		}

		// Update status PO menjadi "received"
		po.Status = "received"
		if err := tx.Save(&po).Error; err != nil {
			tx.Rollback()
			return err
		}

		if err := tx.Commit().Error; err != nil {
			tx.Rollback()
			return err
		}

		// return nil will commit the whole transaction
		return nil
	})

	if err != nil {
		return err
	}

	return nil
}

// CancelPurchaseOrder membatalkan purchase order
func (s *PurchaseService) CancelPurchaseOrder(poID uint) error {
	var po PurchaseOrderModel
	if err := s.db.First(&po, poID).Error; err != nil {
		return err
	}

	// Pastikan status PO adalah "pending"
	if po.Status != "pending" {
		return errors.New("purchase order already processed")
	}

	// Update status PO menjadi "cancelled"
	po.Status = "cancelled"
	if err := s.db.Save(&po).Error; err != nil {
		return err
	}

	return nil
}
