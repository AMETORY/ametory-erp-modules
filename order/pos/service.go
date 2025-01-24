package pos

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/AMETORY/ametory-erp-modules/contact"
	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/finance"
	"github.com/AMETORY/ametory-erp-modules/inventory"
	"github.com/AMETORY/ametory-erp-modules/shared/models"
	"github.com/AMETORY/ametory-erp-modules/utils"
	"github.com/morkid/paginate"
	"gorm.io/gorm"
)

type POSService struct {
	ctx            *context.ERPContext
	db             *gorm.DB
	financeService *finance.FinanceService
	contactService *contact.ContactService
}

func NewPOSService(db *gorm.DB, ctx *context.ERPContext, financeService *finance.FinanceService) *POSService {
	var contactSrv *contact.ContactService
	contactService, ok := ctx.ContactService.(*contact.ContactService)
	if ok {
		contactSrv = contactService
	}

	return &POSService{
		db:             db,
		ctx:            ctx,
		financeService: financeService,
		contactService: contactSrv,
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

func (s *POSService) CreatePosFromOffer(offer models.OfferModel, paymentID, salesNumber, paymentType, paymentTypeProvider string) (*models.POSModel, error) {
	var shippingData struct {
		FullName        string  `json:"full_name"`
		Email           string  `json:"email"`
		Phone           string  `json:"phone_number"`
		Latitude        float64 `json:"latitude"`
		Longitude       float64 `json:"longitude"`
		ShippingAddress string  `json:"shipping_address"`
	}
	var orderRequest models.OrderRequestModel
	if err := s.db.First(&orderRequest, "id = ?", offer.OrderRequestID).Error; err != nil {
		return nil, err
	}
	var contactID *string
	var user models.UserModel

	if err := s.db.First(&user, "id = ?", offer.UserID).Error; err != nil {
		return nil, err
	}

	err := json.Unmarshal([]byte(orderRequest.ShippingData), &shippingData)
	if err != nil {
		return nil, err
	}
	var merchantProductAvailable models.MerchantAvailableProduct
	err = json.Unmarshal([]byte(offer.MerchantAvailableProductData), &merchantProductAvailable)
	if err != nil {
		return nil, err
	}

	var merchant models.MerchantModel
	if err := s.db.First(&merchant, "id = ?", offer.MerchantID).Error; err != nil {
		return nil, err
	}

	if s.contactService != nil {
		if ctct, err := s.contactService.CreateContactFromUser(&user, "", true, false, false, merchant.CompanyID); err != nil {
			if err != nil {
				fmt.Println("ERROR CreateContactFromUser", err.Error())
			} else {
				contactID = &ctct.ID
			}
		}
	}
	var items []models.POSSalesItemModel
	for _, v := range merchantProductAvailable.Items {
		items = append(items, models.POSSalesItemModel{
			ProductID:          &v.ProductID,
			VariantID:          v.VariantID,
			Quantity:           v.Quantity,
			UnitPrice:          v.UnitPrice,
			SubtotalBeforeDisc: v.SubTotal,
			WarehouseID:        merchant.DefaultWarehouseID,
		})
	}

	pos := models.POSModel{
		ContactID:           contactID,
		Code:                utils.RandString(7, true),
		MerchantID:          &offer.MerchantID,
		Total:               offer.TotalPrice,
		Subtotal:            offer.SubTotal,
		ShippingFee:         offer.ShippingFee,
		CompanyID:           merchant.CompanyID,
		Status:              "PENDING",
		PaymentID:           &paymentID,
		OfferID:             &offer.ID,
		ContactData:         orderRequest.ShippingData,
		SalesDate:           time.Now(),
		DueDate:             time.Now().Add(time.Hour * 24),
		PaymentType:         paymentType,
		SalesNumber:         salesNumber,
		PaymentProviderType: models.PaymentProviderType(paymentTypeProvider),
		Items:               items,
	}
	if err := s.db.Create(&pos).Error; err != nil {
		return nil, err
	}

	return &pos, nil
}

// CreatePOSTransaction membuat transaksi POS baru dengan multi-item
func (s *POSService) CreatePOSTransaction(merchantID *string, contactID *string, warehouseID string, items []models.POSSalesItemModel, description string) (*models.POSModel, error) {
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
		Status:     "PENDING",
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

		if s.financeService.TransactionService != nil {
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

func (s *POSService) GetPosSales(request http.Request, search string) (paginate.Page, error) {
	pg := paginate.New()
	stmt := s.db.Preload("Items").Preload("Payment")
	if search != "" {
		stmt = stmt.Where("pos_sales.code ILIKE ? OR pos_sales.description ILIKE ? OR pos_sales.sales_number ILIKE ?",
			"%"+search+"%",
			"%"+search+"%",
			"%"+search+"%",
		)
	}
	if request.Header.Get("ID-Company") != "" {
		stmt = stmt.Where("company_id = ?", request.Header.Get("ID-Company"))
	}

	if request.Header.Get("ID-Merchant") != "" {
		stmt = stmt.Where("merchant_id = ?", request.Header.Get("ID-Merchant"))
	}
	orderBy := request.URL.Query().Get("order_by")
	order := request.URL.Query().Get("order")
	if orderBy == "" {
		orderBy = "created_at"
	}
	if order == "" {
		order = "desc"
	}
	stmt = stmt.Order(orderBy + " " + order)

	stmt = stmt.Model(&models.POSModel{})
	utils.FixRequest(&request)
	page := pg.With(stmt).Request(request).Response(&[]models.POSModel{})
	page.Page = page.Page + 1
	return page, nil
}

func (s *POSService) GetPosSalesDetail(id string) (*models.POSModel, error) {
	var pos models.POSModel
	if err := s.db.Preload("Offer").Preload("Items", func(tx *gorm.DB) *gorm.DB {
		return tx.Preload("Product.Tags").Preload("Variant.Tags")
	}).Preload("Payment").Where("id = ?", id).First(&pos).Error; err != nil {
		return nil, err
	}
	return &pos, nil
}
