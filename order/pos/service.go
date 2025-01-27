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
	ctx              *context.ERPContext
	db               *gorm.DB
	financeService   *finance.FinanceService
	contactService   *contact.ContactService
	inventoryService *inventory.InventoryService
}

func NewPOSService(db *gorm.DB, ctx *context.ERPContext, financeService *finance.FinanceService) *POSService {
	var contactSrv *contact.ContactService
	contactService, ok := ctx.ContactService.(*contact.ContactService)
	if ok {
		contactSrv = contactService
	}
	var inventorySrv *inventory.InventoryService
	inventoryService, ok := ctx.InventoryService.(*inventory.InventoryService)
	if ok {
		inventorySrv = inventoryService
	}

	return &POSService{
		db:               db,
		ctx:              ctx,
		financeService:   financeService,
		contactService:   contactSrv,
		inventoryService: inventorySrv,
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

	if s.contactService == nil {
		return nil, errors.New("contact service not found")
	}
	ctct, err := s.contactService.CreateContactFromUser(&user, "", true, false, false, merchant.CompanyID)
	if err != nil {
		return nil, err
	}
	contactID = &ctct.ID
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
		Tax:                 offer.Tax,
		TaxAmount:           offer.TaxAmount,
		TaxType:             offer.TaxType,
		ServiceFee:          offer.ServiceFee,
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

// GetUserPosSaleByID mengambil transaksi POS berdasarkan user dan id
func (s *POSService) GetUserPosSaleByID(userID string, id string) (*models.POSModel, error) {
	var pos models.POSModel
	if err := s.db.Preload("Items", func(db *gorm.DB) *gorm.DB {
		return db.Preload("Product", func(db *gorm.DB) *gorm.DB {
			return db.Select("display_name", "id")
		}).Preload("Variant", func(db *gorm.DB) *gorm.DB {
			return db.Select("display_name", "id")
		})
	}).Preload("Payment").Where("user_id = ? AND id = ?", userID, id).First(&pos).Error; err != nil {
		return nil, err
	}
	return &pos, nil
}
func (s *POSService) GetUserPosSales(request http.Request, search, userID string) (paginate.Page, error) {
	pg := paginate.New()
	stmt := s.db.Preload("Items", func(db *gorm.DB) *gorm.DB {
		return db.Preload("Product", func(db *gorm.DB) *gorm.DB {
			return db.Select("display_name", "id")
		}).Preload("Variant", func(db *gorm.DB) *gorm.DB {
			return db.Select("display_name", "id")
		})
	}).Preload("Payment")
	if search != "" {
		stmt = stmt.Where("pos_sales.code ILIKE ? OR pos_sales.description ILIKE ? OR pos_sales.sales_number ILIKE ?",
			"%"+search+"%",
			"%"+search+"%",
			"%"+search+"%",
		)
	}
	stmt = stmt.Joins("JOIN contacts ON pos_sales.contact_id = contacts.id").
		Joins("JOIN users ON contacts.user_id = users.id").
		Where("users.id = ?", userID)
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
func (s *POSService) GetPosSales(request http.Request, search string) (paginate.Page, error) {
	pg := paginate.New()
	stmt := s.db.Preload("Items", func(tx *gorm.DB) *gorm.DB {
		return tx.Preload("Product").Preload("Variant")
	}).Preload("Payment")
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
	items := page.Items.(*[]models.POSModel)
	newItems := make([]models.POSModel, 0)
	for _, item := range *items {
		for _, v := range item.Items {
			images, _ := s.inventoryService.ProductService.ListImagesOfProduct(*v.ProductID)
			v.Product.ProductImages = images
		}
		newItems = append(newItems, item)
	}
	page.Items = &newItems

	return page, nil
}

func (s *POSService) UpdateDeliveredByID(id string) error {
	fmt.Println("UPDATE DELIVERED BY ID", id)
	var pos models.POSModel
	err := s.db.Preload("Merchant").Preload("Items").Find(&pos, "id = ?", id).Error
	if err != nil {
		return err
	}

	var shipping models.ShippingModel
	err = s.db.Find(&shipping, "order_id = ?", id).Error
	if err != nil {
		return err
	}

	var stockMovement models.StockMovementModel
	err = s.db.First(&stockMovement, "reference_id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		for _, v := range pos.Items {
			s.inventoryService.StockMovementService.AddMovement(
				time.Now(),
				*v.ProductID,
				*pos.Merchant.DefaultWarehouseID,
				v.VariantID,
				pos.MerchantID,
				nil,
				-v.Quantity,
				models.MovementTypeSale,
				pos.ID,
				fmt.Sprintf("Sales #%s", pos.SalesNumber))
		}
	}

	return nil
}
func (s *POSService) GetPosSalesDetail(id string) (*models.POSModel, error) {
	var pos models.POSModel
	if err := s.db.Preload("Merchant").Preload("Offer.Merchant", func(tx *gorm.DB) *gorm.DB {
		return tx.Preload("Company").Preload("User")
	}).Preload("Items", func(tx *gorm.DB) *gorm.DB {
		return tx.Preload("Product.Tags").Preload("Variant.Tags")
	}).Preload("Payment").Where("id = ?", id).First(&pos).Error; err != nil {
		return nil, err
	}

	for i, v := range pos.Items {
		images, _ := s.inventoryService.ProductService.ListImagesOfProduct(*v.ProductID)
		v.Product.ProductImages = images
		pos.Items[i] = v
	}
	return &pos, nil
}
