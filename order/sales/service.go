package sales

import (
	"errors"
	"net/http"
	"time"

	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/finance"
	"github.com/AMETORY/ametory-erp-modules/finance/account"
	"github.com/AMETORY/ametory-erp-modules/finance/transaction"
	"github.com/AMETORY/ametory-erp-modules/inventory"
	stockmovement "github.com/AMETORY/ametory-erp-modules/inventory/stock_movement"
	"github.com/morkid/paginate"
	"gorm.io/gorm"
)

type SalesService struct {
	ctx            *context.ERPContext
	db             *gorm.DB
	financeService *finance.FinanceService
}

func NewSalesService(db *gorm.DB, ctx *context.ERPContext, financeService *finance.FinanceService) *SalesService {
	return &SalesService{db: db, ctx: ctx, financeService: financeService}
}

func (s *SalesService) CreateSales(data *SalesModel) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := s.db.Create(data).Error; err != nil {
			tx.Rollback()
			return err
		}
		paid := 0.0
		for _, v := range data.Items {
			if v.SaleAccountID != nil {
				s.financeService.TransactionService.CreateTransaction(&transaction.TransactionModel{
					Date:               data.SalesDate,
					AccountID:          v.SaleAccountID,
					Description:        "Penjualan " + data.SalesNumber,
					Notes:              data.Description,
					TransactionRefID:   &data.ID,
					TransactionRefType: "sales",
				}, v.Total)
			}
			if v.AssetAccountID != nil {
				s.financeService.TransactionService.CreateTransaction(&transaction.TransactionModel{
					Date:               data.SalesDate,
					AccountID:          v.AssetAccountID,
					Description:        "Penjualan " + data.SalesNumber,
					Notes:              data.Description,
					TransactionRefID:   &data.ID,
					TransactionRefType: "sales",
				}, v.Total)
				acc, err := s.financeService.AccountService.GetAccountByID(*v.AssetAccountID)
				if err != nil {
					return err
				}
				if acc.Type == account.ASSET {
					paid += v.Total
				}
			}

		}
		if paid > 0 {
			data.Paid = paid
			if err := tx.Save(data).Error; err != nil {
				tx.Rollback()
				return err
			}
		}

		if paid < data.Total {
			data.Status = "partial"
			if err := tx.Save(data).Error; err != nil {
				tx.Rollback()
				return err
			}
		}

		return nil
	})

}

func (s *SalesService) CreatePayment(salesID string, date time.Time, amount float64, accountReceivableID *string, accountAssetID string) error {
	return s.db.Transaction(func(tx *gorm.DB) error {

		var data SalesModel
		if err := tx.Where("id = ?", salesID).First(&data).Error; err != nil {
			return err
		}

		if data.Paid+amount > data.Total {
			return errors.New("amount is greater than total")
		}

		if err := s.financeService.TransactionService.CreateTransaction(&transaction.TransactionModel{
			Date:               date,
			AccountID:          &accountAssetID,
			Description:        "Pembayaran " + data.SalesNumber,
			Notes:              data.Description,
			TransactionRefID:   &data.ID,
			TransactionRefType: "sales",
		}, amount); err != nil {
			return err
		}
		if accountReceivableID != nil {
			if err := s.financeService.TransactionService.CreateTransaction(&transaction.TransactionModel{
				Date:               date,
				AccountID:          accountReceivableID,
				Description:        "Pembayaran " + data.SalesNumber,
				Notes:              data.Description,
				TransactionRefID:   &data.ID,
				TransactionRefType: "sales",
			}, -amount); err != nil {
				return err
			}
		}

		data.Paid += amount
		if err := tx.Save(data).Error; err != nil {
			return err
		}

		if data.Paid == data.Total {
			data.Status = "paid"
			if err := tx.Save(data).Error; err != nil {
				return err
			}
		}

		return nil
	})
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
	if request.Header.Get("ID-Company") != "" {
		stmt = stmt.Where("company_id = ?", request.Header.Get("ID-Company"))
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
	if sales.StockStatus != "pending" {
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

		sales.StockStatus = "updated"
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
