package pos

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"strings"
	"time"

	"github.com/AMETORY/ametory-erp-modules/auth"
	"github.com/AMETORY/ametory-erp-modules/contact"
	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/finance"
	"github.com/AMETORY/ametory-erp-modules/inventory"
	"github.com/AMETORY/ametory-erp-modules/shared/models"
	"github.com/AMETORY/ametory-erp-modules/shared/objects"
	"github.com/AMETORY/ametory-erp-modules/utils"
	"github.com/morkid/paginate"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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

func (s *POSService) CreatePosFromCart(cart models.CartModel, paymentID *string, salesNumber, paymentType, paymentTypeProvider, userPaymentStatus string, taxAmount float64, assetAccountID, saleAccountID *string) (*models.POSModel, *objects.NewUserData, error) {
	var notifUserData *objects.NewUserData
	customerData := struct {
		FullName         string `json:"full_name"`
		Email            string `json:"email"`
		PhoneNumber      string `json:"phone_number"`
		Address          string `json:"address"`
		AutoRegistration bool   `json:"auto_registration"`
	}{}
	var payment models.PaymentModel
	err := s.db.Find(&payment, "id = ?", *paymentID).Error
	if err != nil {
		return nil, nil, err
	}
	if cart.Merchant == nil {
		return nil, nil, errors.New("merchant not found")
	}
	var merchant models.MerchantModel = *cart.Merchant
	totalDiscount := float64(0)
	json.Unmarshal([]byte(cart.CustomerData), &customerData)
	var items []models.POSSalesItemModel
	for _, v := range cart.Items {

		items = append(items, models.POSSalesItemModel{
			ProductID:               &v.ProductID,
			VariantID:               v.VariantID,
			Quantity:                v.Quantity,
			UnitPrice:               v.Price,
			UnitPriceBeforeDiscount: v.OriginalPrice,
			Subtotal:                v.SubTotal,
			SubtotalBeforeDisc:      v.SubTotalBeforeDiscount,
			Height:                  v.Height,
			Length:                  v.Length,
			Weight:                  v.Weight,
			Width:                   v.Weight,
			DiscountPercent:         v.DiscountRate,
			DiscountAmount:          v.DiscountAmount,
			DiscountType:            v.DiscountType,
		})
		totalDiscount += v.DiscountAmount
	}
	var contactID *string
	if customerData.AutoRegistration {
		var existingContact models.UserModel
		if err := s.db.Where("email = ?", customerData.Email).First(&existingContact).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				authSrv, ok := s.ctx.AuthService.(*auth.AuthService)
				if ok {
					var randomPassword = utils.RandomStringNumber(6, false)
					user, err := authSrv.Register(customerData.FullName, utils.CreateUsernameFromFullName(customerData.FullName), customerData.Email, randomPassword, "")
					if err == nil {
						existingContact = *user
						notifUserData = &objects.NewUserData{
							Email:             user.Email,
							FullName:          user.FullName,
							Password:          randomPassword,
							VerificationToken: user.VerificationToken,
						}
					}
				}

			}

		}
		contact, err := s.contactService.CreateContactFromUser(&existingContact, "", true, false, false, merchant.CompanyID)
		if err != nil {
			return nil, nil, err
		}
		contactID = &contact.ID
	}
	status := "PENDING"

	paid := float64(0)
	if paymentType == "CASH" {
		paid = cart.Total
		cart.ServiceFee = 0
		status = "COMPLETED"
	}

	pos := models.POSModel{
		ContactID:              contactID,
		Code:                   utils.RandString(7, true),
		MerchantID:             cart.MerchantID,
		Total:                  payment.Total,
		Subtotal:               cart.SubTotal,
		SubTotalBeforeDiscount: cart.SubTotalBeforeDiscount,
		PaymentFee:             payment.PaymentFee,
		Tax:                    cart.Tax,
		TaxAmount:              cart.TaxAmount,
		TaxType:                cart.TaxType,
		ServiceFee:             cart.ServiceFee,
		Paid:                   paid,
		CompanyID:              merchant.CompanyID,
		Status:                 status,
		UserPaymentStatus:      userPaymentStatus,
		PaymentID:              paymentID,
		CartID:                 &cart.ID,
		ContactData:            cart.CustomerData,
		SalesDate:              time.Now(),
		DueDate:                time.Now().Add(time.Hour * 24),
		PaymentType:            paymentType,
		SalesNumber:            salesNumber,
		PaymentProviderType:    models.PaymentProviderType(paymentTypeProvider),
		Items:                  items,
		AssetAccountID:         assetAccountID,
		SaleAccountID:          saleAccountID,
		TotalDiscount:          totalDiscount,
	}

	if err := s.db.Create(&pos).Error; err != nil {
		return nil, nil, err
	}

	if (strings.ToLower(userPaymentStatus) == "paid" || strings.ToLower(userPaymentStatus) == "complete") && pos.SaleAccountID != nil && pos.AssetAccountID != nil {
		if s.financeService.TransactionService != nil {
			// Tambahkan transaksi ke jurnal
			err := s.UpdateTransaction(&pos, merchant)
			if err != nil {
				return nil, nil, err
			}
		}

	}

	return &pos, notifUserData, nil
}

func (s *POSService) UpdateTransaction(pos *models.POSModel, merchant models.MerchantModel) error {
	var existingTransaction models.TransactionModel
	err := s.db.Where("transaction_ref_type = ? AND transaction_ref_id = ?", "pos_sales", pos.ID).First(&existingTransaction).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	if err == nil {
		// Transaction exists, handle accordingly
		return nil
	}
	// Transaction does not exist, proceed with creating a new one

	now := time.Now()
	if pos.SaleAccountID != nil {
		if err := s.financeService.TransactionService.CreateTransaction(&models.TransactionModel{
			Date:               now,
			AccountID:          pos.SaleAccountID,
			Description:        fmt.Sprintf("Penjualan [%s] %s ", merchant.Name, pos.SalesNumber),
			Notes:              pos.Description,
			TransactionRefID:   &pos.ID,
			TransactionRefType: "pos_sales",
			CompanyID:          pos.CompanyID,
		}, pos.Total); err != nil {
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
		}, pos.Total); err != nil {
			return err
		}
	}

	return nil
}
func (s *POSService) CreatePosFromOffer(offer models.OfferModel, paymentID, salesNumber, paymentType, paymentTypeProvider, userPaymentStatus string, assetAccountID, saleAccountID *string, orderType string) (*models.POSModel, error) {
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
	var payment models.PaymentModel
	if err := s.db.First(&payment, "id = ?", paymentID).Error; err != nil {
		return nil, err
	}
	var items []models.POSSalesItemModel
	totalDiscount := float64(0)
	for _, v := range merchantProductAvailable.Items {
		var Height, Length, Weight, Width float64
		product := models.ProductModel{}
		s.db.Select("height, length, weight, width").First(&product, "id = ?", v.ProductID)
		Height = product.Height
		Length = product.Length
		Weight = product.Weight
		Width = product.Width
		if v.VariantID != nil {
			variant := models.VariantModel{}
			s.db.Select("height, length, weight, width").First(&variant, "id = ?", v.VariantID)
			Height = variant.Height
			Length = variant.Length
			Weight = variant.Weight
			Width = variant.Width
		}
		items = append(items, models.POSSalesItemModel{
			ProductID:               &v.ProductID,
			VariantID:               v.VariantID,
			Quantity:                v.Quantity,
			UnitPrice:               v.UnitPrice,
			UnitPriceBeforeDiscount: v.UnitPriceBeforeDiscount,
			SubtotalBeforeDisc:      v.SubTotalBeforeDiscount,
			Subtotal:                v.SubTotal,
			WarehouseID:             merchant.DefaultWarehouseID,
			Height:                  Height,
			Length:                  Length,
			Weight:                  Weight,
			Width:                   Width,
			DiscountPercent:         v.DiscountValue,
			DiscountAmount:          v.DiscountAmount,
			DiscountType:            v.DiscountType,
		})
		totalDiscount += v.DiscountAmount
	}

	if orderType == "" {
		orderType = "ONLINE"
	}

	pos := models.POSModel{
		ContactID:              contactID,
		Code:                   utils.RandString(7, true),
		MerchantID:             &offer.MerchantID,
		Total:                  offer.TotalPrice + payment.PaymentFee,
		Subtotal:               offer.SubTotal,
		SubTotalBeforeDiscount: offer.SubTotalBeforeDiscount,
		ShippingFee:            offer.ShippingFee,
		PaymentFee:             payment.PaymentFee,
		Tax:                    offer.Tax,
		TaxAmount:              offer.TaxAmount,
		TaxType:                offer.TaxType,
		ServiceFee:             offer.ServiceFee,
		CompanyID:              merchant.CompanyID,
		Status:                 "PENDING",
		UserPaymentStatus:      userPaymentStatus,
		PaymentID:              &paymentID,
		OfferID:                &offer.ID,
		ContactData:            orderRequest.ShippingData,
		SalesDate:              time.Now(),
		DueDate:                time.Now().Add(time.Hour * 24),
		PaymentType:            paymentType,
		SalesNumber:            salesNumber,
		PaymentProviderType:    models.PaymentProviderType(paymentTypeProvider),
		Items:                  items,
		AssetAccountID:         assetAccountID,
		SaleAccountID:          saleAccountID,
		OrderType:              orderType,
		TotalBeforeDisc:        totalDiscount,
	}
	if err := s.db.Create(&pos).Error; err != nil {
		return nil, err
	}

	if (strings.ToLower(userPaymentStatus) == "paid" || strings.ToLower(userPaymentStatus) == "complete") && pos.SaleAccountID != nil && pos.AssetAccountID != nil {
		if s.financeService.TransactionService != nil {
			// Tambahkan transaksi ke jurnal
			err := s.UpdateTransaction(&pos, merchant)
			if err != nil {
				return nil, err
			}
		}

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
			_, err := invSrv.StockMovementService.AddMovement(now, *item.ProductID, warehouseID, item.VariantID, merchantID, nil, nil, -item.Quantity, models.MovementTypeOut, pos.ID, description)
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
	if err := s.db.Preload("Contact").Preload("Items", func(db *gorm.DB) *gorm.DB {
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

// GetUserPosSaleDetail mengambil detail transaksi POS berdasarkan user dan id
func (s *POSService) GetUserPosSaleDetail(id string) (*models.POSModel, error) {
	var pos models.POSModel
	if err := s.db.Preload("Contact").Preload("Items", func(db *gorm.DB) *gorm.DB {
		return db.Preload("Product", func(db *gorm.DB) *gorm.DB {
			return db.Select("display_name", "id", "category_id", "brand_id").Preload("Category").Preload("Brand")
		}).Preload("Variant", func(db *gorm.DB) *gorm.DB {
			return db.Select("display_name", "id")
		})
	}).Preload("Payment").Where("id = ?", id).First(&pos).Error; err != nil {
		return nil, err
	}
	for i, v := range pos.Items {
		productImages, _ := s.inventoryService.ProductService.ListImagesOfProduct(*v.ProductID)
		v.Product.ProductImages = productImages
		pos.Items[i] = v
	}

	pos.ShippingStatus = "PENDING"

	var shipping models.ShippingModel
	err := s.db.First(&shipping, "order_id = ?", pos.ID).Error
	if err == nil {
		pos.Shipping = &shipping
		pos.ShippingStatus = shipping.Status

	}
	return &pos, nil
}
func (s *POSService) GetUserPosSales(request http.Request, search, userID string) (paginate.Page, error) {
	pg := paginate.New()
	stmt := s.db.Preload("Items", func(db *gorm.DB) *gorm.DB {
		return db.Preload("Product", func(db *gorm.DB) *gorm.DB {
			return db.Select("display_name", "id", "category_id", "brand_id").Preload("Category").Preload("Brand")
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
	items := page.Items.(*[]models.POSModel)
	newItems := make([]models.POSModel, 0)
	for _, item := range *items {
		for _, v := range item.Items {
			images, _ := s.inventoryService.ProductService.ListImagesOfProduct(*v.ProductID)
			v.Product.ProductImages = images
		}

		item.ShippingStatus = "PENDING"

		var shipping models.ShippingModel
		err := s.db.First(&shipping, "order_id = ?", item.ID).Error
		if err == nil {
			item.Shipping = &shipping
			item.ShippingStatus = shipping.Status

		}
		newItems = append(newItems, item)
	}
	page.Items = &newItems
	return page, nil
}

func (s *POSService) GetPosSalesDetail(id string) (*models.POSModel, error) {
	var pos models.POSModel
	if err := s.db.Preload("Contact.User").Preload("Merchant").Preload("Offer.Merchant", func(tx *gorm.DB) *gorm.DB {
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
	pos.ShippingStatus = "PENDING"

	var shipping models.ShippingModel
	err := s.db.First(&shipping, "order_id = ?", id).Error
	if err == nil {
		pos.Shipping = &shipping
		pos.ShippingStatus = shipping.Status

	}
	return &pos, nil
}

func (s *POSService) GetPosSales(request http.Request, search string) (paginate.Page, error) {
	pg := paginate.New()
	stmt := s.db.Preload("Merchant").Preload("Items", func(tx *gorm.DB) *gorm.DB {
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

	if request.URL.Query().Get("order_type") != "" {
		stmt = stmt.Where("order_type = ?", request.URL.Query().Get("order_type"))
	}
	if request.URL.Query().Get("status") != "" {
		stmt = stmt.Where("status = ?", request.URL.Query().Get("status"))
	}
	if request.URL.Query().Get("shipping_status") != "" {
		if request.URL.Query().Get("shipping_status") == "PENDING" {
			stmt = stmt.Joins("LEFT JOIN shippings ON shippings.order_id = pos_sales.id").Where("shippings.status is null")
		} else {
			stmt = stmt.Joins("LEFT JOIN shippings ON shippings.order_id = pos_sales.id").Where("shippings.status = ?", request.URL.Query().Get("shipping_status"))
		}

	}

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

		item.ShippingStatus = "PENDING"

		var shipping models.ShippingModel
		err := s.db.First(&shipping, "order_id = ?", item.ID).Error
		if err == nil {
			item.Shipping = &shipping
			item.ShippingStatus = shipping.Status

		}
		newItems = append(newItems, item)
	}
	page.Items = &newItems

	return page, nil
}

func (s *POSService) GetPushTokenFromID(id string) ([]string, error) {
	var pos models.POSModel
	err := s.db.Preload("Contact.User").Find(&pos, "id = ?", id).Error
	if err != nil {
		return []string{}, err
	}

	userIDs := []string{pos.Contact.User.ID}
	// merchant.Company.Users = make([]models.UserModel, 0)
	// merchant.Company.Users = append(merchant.Company.Users, *merchant.User)

	var pushToken []models.PushTokenModel
	err = s.db.Where("user_id IN (?)", userIDs).Find(&pushToken).Error
	if err != nil {
		return nil, err
	}

	tokens := make([]string, 0)
	for _, v := range pushToken {
		tokens = append(tokens, v.Token)
	}
	return tokens, nil
}

func (s *POSService) UpdatePickedByID(id string) error {
	fmt.Println("UPDATE PICKED BY ID", id)
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
				nil,
				-v.Quantity,
				models.MovementTypeSale,
				pos.ID,
				fmt.Sprintf("Sales #%s", pos.SalesNumber))
		}
	}

	pos.StockStatus = "IN_DELIVERY"
	if err := s.db.Omit(clause.Associations).Save(&pos).Error; err != nil {
		return err
	}

	return nil
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
				nil,
				-v.Quantity,
				models.MovementTypeSale,
				pos.ID,
				fmt.Sprintf("Sales #%s", pos.SalesNumber))
		}
	}

	pos.StockStatus = "DELIVERED"
	if err := s.db.Omit(clause.Associations).Save(&pos).Error; err != nil {
		return err
	}

	return nil
}

func (s *POSService) CountPosSalesByStatus(status string) (int64, error) {

	var count int64
	if err := s.db.Model(&models.POSModel{}).Where("status = ?", status).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (s *POSService) GetPosSalesByDate(dateRange string, status string) ([]models.POSModel, error) {
	var pos []models.POSModel
	var start, end time.Time
	switch dateRange {
	case "TODAY":
		start = time.Now().Truncate(24 * time.Hour)
		end = start.Add(24 * time.Hour)
	case "THIS_WEEK":
		start = time.Now().Truncate(24*time.Hour).AddDate(0, 0, -int(time.Now().Weekday()))
		end = start.Add(7 * 24 * time.Hour)
	case "THIS_MONTH":
		start = time.Now().Truncate(24*time.Hour).AddDate(0, 0, -int(time.Now().Day()))
		end = start.AddDate(0, 1, 0).Add(-time.Nanosecond)
	case "THIS_YEAR":
		start = time.Now().Truncate(24*time.Hour).AddDate(0, 0, -int(time.Now().Day()))
		end = start.AddDate(1, 0, 0).Add(-time.Nanosecond)
	case "LAST_7_DAYS":
		start = time.Now().Truncate(24 * time.Hour).Add(-7 * 24 * time.Hour)
		end = time.Now().Truncate(24 * time.Hour)
	case "LAST_MONTH":
		start = time.Now().Truncate(24*time.Hour).AddDate(0, 0, -int(time.Now().Day()))
		end = start.AddDate(0, 1, 0).Add(-time.Nanosecond)
	case "LAST_QUARTER":
		now := time.Now()
		startMonth := (now.Month() - 1) / 3 * 3
		start := time.Date(now.Year(), time.Month(startMonth+1), 1, 0, 0, 0, 0, now.Location())
		end = start.AddDate(0, 3, 0).Add(-time.Nanosecond)
	default:
		return nil, errors.New("invalid date range")
	}

	if err := s.db.Where("created_at BETWEEN ? AND ?", start, end).Where("status = ?", status).Find(&pos).Error; err != nil {
		return nil, err
	}

	return pos, nil
}

// get pos sales data and generate to invoice html
func (s *POSService) DownloadInvoice(id, layout, body string) ([]byte, error) {
	var pos models.POSModel
	if err := s.db.Preload("Contact.User").Preload("Payment").Preload("Items.Product").Preload("Items.Variant").Preload("Merchant.Company").First(&pos, "id = ?", id).Error; err != nil {
		return nil, err
	}

	var items []map[string]interface{}
	for _, item := range pos.Items {
		imageUrl := ""
		images, _ := s.inventoryService.ProductService.ListImagesOfProduct(*item.ProductID)
		if len(images) > 0 {
			imageUrl = images[0].URL
		}
		productName := item.Product.DisplayName
		if item.VariantID != nil {
			productName = item.Variant.DisplayName
		}
		disc := ""
		if item.DiscountPercent > 0 {
			disc = utils.FormatCurrency(item.DiscountPercent) + "%"
		}
		items = append(items, map[string]interface{}{
			"ProductName":        productName,
			"Quantity":           utils.FormatCurrency(item.Quantity),
			"Price":              utils.FormatCurrency(item.UnitPriceBeforeDiscount),
			"SubTotal":           utils.FormatCurrency(item.SubtotalBeforeDisc),
			"Image":              imageUrl,
			"DiscountPercentage": disc,
			"Description":        item.Product.Description,
		})
	}
	var buyerAddress = ""
	buyerAddress, ok := pos.DataContact["address"].(string)
	if !ok {
		buyerAddress, _ = pos.DataContact["shipping_address"].(string)

	}

	pdfData := struct {
		SalesNumber            string
		Items                  []map[string]interface{}
		SalesDate              string
		DueDate                string
		CustomerData           map[string]interface{}
		BuyerName              string
		BuyerAddress           string
		BuyerPhone             string
		BuyerEmail             string
		Code                   string
		Description            string
		Notes                  string
		Total                  string
		DiscountAmount         string
		Subtotal               string
		SubTotalBeforeDiscount string
		ShippingFee            string
		ServiceFee             string
		PaymentFee             string
		Tax                    string
		TaxType                string
		TaxAmount              string
		MerchantName           string
		MerchantAddress        string
		MerchantPhone          string
		MerchantEmail          string
		PaymentMethod          string
		UserPaymentStatus      string
	}{
		MerchantName:           pos.Merchant.Name,
		MerchantAddress:        pos.Merchant.Address,
		MerchantPhone:          pos.Merchant.Phone,
		MerchantEmail:          pos.Merchant.Company.Email,
		SalesNumber:            pos.SalesNumber,
		Items:                  items,
		SalesDate:              pos.SalesDate.Format("02/01/2006"),
		DueDate:                pos.DueDate.Format("02/01/2006"),
		CustomerData:           pos.DataContact,
		Code:                   pos.Code,
		Description:            pos.Description,
		Notes:                  pos.Notes,
		Total:                  utils.FormatCurrency(pos.Total),
		Subtotal:               utils.FormatCurrency(pos.Subtotal),
		SubTotalBeforeDiscount: utils.FormatCurrency(pos.SubTotalBeforeDiscount),
		ShippingFee:            utils.FormatCurrency(pos.ShippingFee),
		ServiceFee:             utils.FormatCurrency(pos.ServiceFee),
		PaymentFee:             utils.FormatCurrency(pos.PaymentFee),
		Tax:                    utils.FormatCurrency(pos.Tax),
		DiscountAmount:         utils.FormatCurrency(pos.SubTotalBeforeDiscount - pos.Subtotal),
		TaxType:                pos.TaxType,
		TaxAmount:              utils.FormatCurrency(pos.TaxAmount),
		PaymentMethod:          pos.Payment.PaymentMethod,
		UserPaymentStatus:      pos.UserPaymentStatus,
		BuyerName:              pos.DataContact["full_name"].(string),
		BuyerAddress:           buyerAddress,
		BuyerPhone:             pos.DataContact["phone_number"].(string),
		BuyerEmail:             pos.DataContact["email"].(string),
	}

	t := template.Must(template.ParseFiles(layout, body))

	var buf bytes.Buffer
	if err := t.ExecuteTemplate(&buf, "layout", pdfData); err != nil {
		return nil, err
	}

	return utils.GeneratePDF(s.ctx.Config.WkhtmltopdfPath, s.ctx.Config.PdfFooter, buf.String())
}
