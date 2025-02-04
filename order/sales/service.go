package sales

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/finance"
	"github.com/AMETORY/ametory-erp-modules/inventory"
	"github.com/AMETORY/ametory-erp-modules/shared/models"
	"github.com/AMETORY/ametory-erp-modules/utils"
	"github.com/morkid/paginate"
	"gorm.io/gorm"
)

type SalesService struct {
	ctx            *context.ERPContext
	db             *gorm.DB
	financeService *finance.FinanceService
}

func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(&models.SalesModel{}, &models.SalesItemModel{})
}

func NewSalesService(db *gorm.DB, ctx *context.ERPContext, financeService *finance.FinanceService) *SalesService {
	return &SalesService{db: db, ctx: ctx, financeService: financeService}
}

func (s *SalesService) CreateSales(data *models.SalesModel) error {
	var companyID *string
	if s.ctx.Request.Header.Get("ID-Company") != "" {
		compID := s.ctx.Request.Header.Get("ID-Company")
		companyID = &compID
	}
	return s.db.Transaction(func(tx *gorm.DB) error {
		data.CompanyID = companyID
		if err := s.db.Create(data).Error; err != nil {
			tx.Rollback()
			return err
		}
		if s.financeService.TransactionService != nil {
			paid := 0.0
			for _, v := range data.Items {
				if v.SaleAccountID != nil {
					s.financeService.TransactionService.CreateTransaction(&models.TransactionModel{
						Date:               data.SalesDate,
						AccountID:          v.SaleAccountID,
						Description:        "Penjualan " + data.SalesNumber,
						Notes:              data.Description,
						TransactionRefID:   &data.ID,
						TransactionRefType: "sales",
						CompanyID:          companyID,
					}, v.Total)
				}
				if v.AssetAccountID != nil {
					s.financeService.TransactionService.CreateTransaction(&models.TransactionModel{
						Date:               data.SalesDate,
						AccountID:          v.AssetAccountID,
						Description:        "Penjualan " + data.SalesNumber,
						Notes:              data.Description,
						TransactionRefID:   &data.ID,
						TransactionRefType: "sales",
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
		}

		if err := tx.Commit().Error; err != nil {
			tx.Rollback()
			return err
		}

		return nil
	})

}

func (s *SalesService) CreatePayment(salesID string, date time.Time, amount float64, accountReceivableID *string, accountAssetID string) error {
	var companyID *string
	if s.ctx.Request.Header.Get("ID-Company") != "" {
		compID := s.ctx.Request.Header.Get("ID-Company")
		companyID = &compID
	}
	return s.db.Transaction(func(tx *gorm.DB) error {

		var data models.SalesModel
		if err := tx.Where("id = ?", salesID).First(&data).Error; err != nil {
			return err
		}

		if data.Paid+amount > data.Total {
			return errors.New("amount is greater than total")
		}

		if err := s.financeService.TransactionService.CreateTransaction(&models.TransactionModel{
			Date:               date,
			AccountID:          &accountAssetID,
			Description:        "Pembayaran " + data.SalesNumber,
			Notes:              data.Description,
			TransactionRefID:   &data.ID,
			TransactionRefType: "sales",
			CompanyID:          companyID,
		}, amount); err != nil {
			return err
		}
		if accountReceivableID != nil {
			if err := s.financeService.TransactionService.CreateTransaction(&models.TransactionModel{
				Date:               date,
				AccountID:          accountReceivableID,
				Description:        "Pembayaran " + data.SalesNumber,
				Notes:              data.Description,
				TransactionRefID:   &data.ID,
				TransactionRefType: "sales",
				CompanyID:          companyID,
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

		if err := tx.Commit().Error; err != nil {
			tx.Rollback()
			return err
		}

		return nil
	})
}

func (s *SalesService) UpdateSales(id string, data *models.SalesModel) error {
	return s.db.Where("id = ?", id).Updates(data).Error
}

func (s *SalesService) DeleteSales(id string) error {
	return s.db.Where("id = ?", id).Delete(&models.SalesModel{}).Error
}

func (s *SalesService) GetSalesByID(id string) (*models.SalesModel, error) {
	var sales models.SalesModel
	err := s.db.Where("id = ?", id).First(&sales).Error
	return &sales, err
}

func (s *SalesService) GetSalesByCode(code string) (*models.SalesModel, error) {
	var sales models.SalesModel
	err := s.db.Where("code = ?", code).First(&sales).Error
	return &sales, err
}

func (s *SalesService) GetSalesBySalesNumber(salesNumber string) (*models.SalesModel, error) {
	var sales models.SalesModel
	err := s.db.Where("sales_number = ?", salesNumber).First(&sales).Error
	return &sales, err
}

func (s *SalesService) GetSales(request http.Request, search string) (paginate.Page, error) {
	pg := paginate.New()
	stmt := s.db
	if search != "" {
		stmt = stmt.Where("sales.description ILIKE ? OR sales.code ILIKE ? OR sales.sales_number ILIKE ?",
			"%"+search+"%",
			"%"+search+"%",
			"%"+search+"%",
		)
	}
	if request.Header.Get("ID-Company") != "" {
		stmt = stmt.Where("company_id = ?", request.Header.Get("ID-Company"))
	}
	stmt = stmt.Model(&models.SalesModel{})
	utils.FixRequest(&request)
	page := pg.With(stmt).Request(request).Response(&[]models.SalesModel{})
	page.Page = page.Page + 1
	return page, nil
}

func (s *SalesService) UpdateStock(salesID, warehouseID string, description string) error {
	var sales models.SalesModel
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
			_, err := invSrv.StockMovementService.AddMovement(sales.SalesDate, *v.ProductID, *v.WarehouseID, v.VariantID, nil, nil, nil, -v.Quantity, models.MovementTypeIn, sales.ID, description)
			if err != nil {
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

func (s *SalesService) CreateSalesFromOrderRequest(orderRequest *models.OrderRequestModel, salesNumber string, taxPercent float64, description string) error {
	var companyID *string
	if s.ctx.Request.Header.Get("ID-Company") != "" {
		compID := s.ctx.Request.Header.Get("ID-Company")
		companyID = &compID
	}
	if orderRequest.ContactID == nil {
		return errors.New("contact ID is required")
	}
	contactData, err := json.Marshal(*orderRequest.Contact)
	if err != nil {
		return err
	}

	data := &models.SalesModel{
		SalesNumber:     salesNumber,
		Code:            utils.RandString(10, true),
		SalesDate:       orderRequest.CreatedAt,
		DueDate:         orderRequest.ExpiresAt,
		TotalBeforeTax:  0,
		TotalBeforeDisc: 0,
		Subtotal:        0,
		Paid:            0,
		CompanyID:       companyID,
		ContactID:       *orderRequest.ContactID,
		ContactData:     string(contactData),
		Type:            models.ECOMMERCE,
		Items:           []models.SalesItemModel{},
	}
	var totalBeforeTax, totalBeforeDisc float64
	for _, v := range orderRequest.Items {
		data.Items = append(data.Items, models.SalesItemModel{
			ProductID:          v.ProductID,
			Quantity:           v.Quantity,
			UnitPrice:          v.UnitPrice,
			Total:              v.Quantity * v.UnitPrice,
			DiscountPercent:    v.DiscountPercent,
			DiscountAmount:     v.DiscountAmount,
			SubtotalBeforeDisc: v.Quantity * v.UnitPrice,
		})
		totalBeforeDisc += v.Quantity * v.UnitPrice
		if v.DiscountPercent > 0 {
			totalBeforeTax += v.Quantity * (v.UnitPrice - (v.UnitPrice * v.DiscountPercent / 100))
		} else {
			totalBeforeTax += v.Quantity * (v.UnitPrice - v.DiscountAmount)
		}

	}
	data.TotalBeforeTax = totalBeforeTax
	data.TotalBeforeDisc = totalBeforeDisc
	data.Subtotal = totalBeforeTax * (1 + taxPercent/100)
	return s.CreateSales(data)
}
