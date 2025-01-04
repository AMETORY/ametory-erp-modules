package sales

import (
	"errors"
	"net/http"

	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/inventory"
	stockmovement "github.com/AMETORY/ametory-erp-modules/inventory/stock_movement"
	"github.com/morkid/paginate"
	"gorm.io/gorm"
)

type SalesService struct {
	ctx *context.ERPContext
	db  *gorm.DB
}

func NewSalesService(db *gorm.DB, ctx *context.ERPContext) *SalesService {
	return &SalesService{db: db, ctx: ctx}
}

func (s *SalesService) CreateSales(data *SalesModel) error {
	return s.db.Create(data).Error
}

func (s *SalesService) UpdateSales(id string, data *SalesModel) error {
	return s.db.Where("id = ?", id).Updates(data).Error
}

func (s *SalesService) DeleteSales(id string) error {
	return s.db.Where("id = ?", id).Delete(&SalesModel{}).Error
}

func (s *SalesService) GetSalesByID(id string) (*SalesModel, error) {
	var sales SalesModel
	err := s.db.Where("id = ?", id).First(&sales).Error
	return &sales, err
}

func (s *SalesService) GetSalesByCode(code string) (*SalesModel, error) {
	var sales SalesModel
	err := s.db.Where("code = ?", code).First(&sales).Error
	return &sales, err
}

func (s *SalesService) GetSalesBySalesNumber(salesNumber string) (*SalesModel, error) {
	var sales SalesModel
	err := s.db.Where("sales_number = ?", salesNumber).First(&sales).Error
	return &sales, err
}

func (s *SalesService) GetSales(request http.Request, search string) (paginate.Page, error) {
	pg := paginate.New()
	stmt := s.db
	if search != "" {
		stmt = stmt.Where("sales.description LIKE ? OR sales.code LIKE ? OR sales.sales_number LIKE ?",
			"%"+search+"%",
			"%"+search+"%",
			"%"+search+"%",
		)
	}
	stmt = stmt.Model(&SalesModel{})
	page := pg.With(stmt).Request(request).Response(&[]SalesModel{})
	return page, nil
}

func (s *SalesService) UpdateStock(salesID, warehouseID string) error {
	var sales SalesModel
	if err := s.db.First(&sales, salesID).Error; err != nil {
		return err
	}

	invSrv, ok := s.ctx.InventoryService.(*inventory.InventoryService)
	if !ok {
		return errors.New("invalid inventory service")
	}

	// Pastikan status PO adalah "pending"
	if sales.Status != "pending" {
		return errors.New("purchase order already processed")
	}

	err := s.ctx.DB.Transaction(func(tx *gorm.DB) error {
		for _, v := range sales.Items {
			if v.ProductID == nil || v.WarehouseID == nil {
				continue
			}
			if err := invSrv.StockMovementService.AddMovement(*v.ProductID, *v.WarehouseID, -v.Quantity, stockmovement.MovementTypeIn, sales.ID); err != nil {
				tx.Rollback()
				return err
			}
		}

		sales.Status = "updated"
		if err := tx.Save(&sales).Error; err != nil {
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
