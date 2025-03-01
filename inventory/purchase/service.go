package purchase

import (
	"errors"
	"net/http"
	"time"

	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/finance"
	stockmovement "github.com/AMETORY/ametory-erp-modules/inventory/stock_movement"
	"github.com/AMETORY/ametory-erp-modules/shared/models"
	"github.com/AMETORY/ametory-erp-modules/utils"
	"github.com/morkid/paginate"
	"gorm.io/gorm"
)

type PurchaseService struct {
	db                   *gorm.DB
	ctx                  *context.ERPContext
	financeService       *finance.FinanceService
	stockMovementService *stockmovement.StockMovementService
}

func NewPurchaseService(db *gorm.DB, ctx *context.ERPContext, financeService *finance.FinanceService, stockMovementService *stockmovement.StockMovementService) *PurchaseService {
	return &PurchaseService{
		db:                   db,
		ctx:                  ctx,
		financeService:       financeService,
		stockMovementService: stockMovementService,
	}
}

func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(&models.PurchaseOrderModel{}, &models.PurchaseOrderItemModel{})
}

// CreatePurchaseOrder membuat purchase order baru
func (s *PurchaseService) CreatePurchaseOrder(data *models.PurchaseOrderModel) error {
	var companyID *string
	if s.ctx.Request.Header.Get("ID-Company") != "" {
		compID := s.ctx.Request.Header.Get("ID-Company")
		companyID = &compID
	}
	// Hitung total harga

	return s.db.Transaction(func(tx *gorm.DB) error {
		// Membuat purchase order
		data.CompanyID = companyID
		if err := s.db.Create(data).Error; err != nil {
			tx.Rollback()
			return err
		}
		paid := 0.0
		for _, v := range data.Items {
			if v.PurchaseAccountID != nil {
				s.financeService.TransactionService.CreateTransaction(&models.TransactionModel{
					Date:               data.PurchaseDate,
					AccountID:          v.PurchaseAccountID,
					Description:        "Pembelian " + data.PurchaseNumber,
					Notes:              data.Description,
					TransactionRefID:   &data.ID,
					TransactionRefType: "purchase",
					CompanyID:          companyID,
				}, v.Total)
			}
			if v.AssetAccountID != nil {
				s.financeService.TransactionService.CreateTransaction(&models.TransactionModel{
					Date:               data.PurchaseDate,
					AccountID:          v.AssetAccountID,
					Description:        "Pembelian " + data.PurchaseNumber,
					Notes:              data.Description,
					TransactionRefID:   &data.ID,
					TransactionRefType: "purchase",
					CompanyID:          companyID,
				}, v.Total)
				acc, err := s.financeService.AccountService.GetAccountByID(*v.AssetAccountID)
				if err != nil {
					return err
				}
				if acc.Type == models.ASSET {
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

		if err := tx.Commit().Error; err != nil {
			tx.Rollback()
			return err
		}

		return nil
	})
}

// ReceivePurchaseOrder menerima barang dari supplier dan menambah stok
func (s *PurchaseService) ReceivePurchaseOrder(date time.Time, poID, warehouseID string, description string) error {
	// companyID := s.ctx.Request.Header.Get("ID-Company")
	var po models.PurchaseOrderModel
	if err := s.db.First(&po, poID).Error; err != nil {
		return err
	}

	// Pastikan status PO adalah "pending"
	if po.StockStatus != "pending" {
		return errors.New("purchase order already processed")
	}

	err := s.ctx.DB.Transaction(func(tx *gorm.DB) error {
		// do some database operations in the transaction (use 'tx' from this point, not 'db')
		for _, v := range po.Items {
			if v.ProductID == nil || v.WarehouseID == nil {
				continue
			}
			if _, err := s.stockMovementService.AddMovement(date, *v.ProductID, *v.WarehouseID, v.VariantID, nil, nil, nil, v.Quantity, models.MovementTypeIn, po.ID, description); err != nil {
				tx.Rollback()
				return err
			}
		}

		// Update status PO menjadi "received"
		po.StockStatus = "received"
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
	var po models.PurchaseOrderModel
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

// CreatePayment membuat payment untuk purchase order
func (s *PurchaseService) CreatePayment(poID string, date time.Time, amount float64, accountPayableID *string, accountAssetID string) error {
	var companyID *string
	if s.ctx.Request.Header.Get("ID-Company") != "" {
		compID := s.ctx.Request.Header.Get("ID-Company")
		companyID = &compID
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		var data models.PurchaseOrderModel
		if err := s.db.First(&data, poID).Error; err != nil {
			return err
		}

		// Pastikan status PO adalah "pending"
		if data.Status != "pending" {
			return errors.New("purchase order already processed")
		}

		if data.Paid+amount > data.Total {
			return errors.New("amount is greater than total")
		}

		if err := s.financeService.TransactionService.CreateTransaction(&models.TransactionModel{
			Date:               date,
			AccountID:          &accountAssetID,
			Description:        "Pembayaran " + data.PurchaseNumber,
			Notes:              data.Description,
			TransactionRefID:   &data.ID,
			TransactionRefType: "purchase",
			CompanyID:          companyID,
		}, -amount); err != nil {
			return err
		}

		if accountPayableID != nil {
			if err := s.financeService.TransactionService.CreateTransaction(&models.TransactionModel{
				Date:               date,
				AccountID:          accountPayableID,
				Description:        "Pembayaran " + data.PurchaseNumber,
				Notes:              data.Description,
				TransactionRefID:   &data.ID,
				TransactionRefType: "purchase",
				CompanyID:          companyID,
			}, amount); err != nil {
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

		if err := tx.Commit().Error; err != nil {
			tx.Rollback()
			return err
		}

		return nil
	})
}

func (s *PurchaseService) GetPurchases(request http.Request, search string) (paginate.Page, error) {
	pg := paginate.New()
	stmt := s.db
	if search != "" {
		stmt = stmt.Where("description ILIKE ? OR purchase_number ILIKE ?",
			"%"+search+"%",
			"%"+search+"%",
		)
	}
	if request.Header.Get("ID-Company") != "" {
		stmt = stmt.Where("company_id = ?", request.Header.Get("ID-Company"))
	}
	stmt = stmt.Model(&models.PurchaseOrderModel{})
	utils.FixRequest(&request)
	page := pg.With(stmt).Request(request).Response(&[]models.PurchaseOrderModel{})
	page.Page = page.Page + 1
	return page, nil
}
