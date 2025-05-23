package merchant

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/finance"
	"github.com/AMETORY/ametory-erp-modules/inventory"
	"github.com/AMETORY/ametory-erp-modules/shared/models"
	"github.com/AMETORY/ametory-erp-modules/utils"
	"github.com/google/uuid"
	"github.com/morkid/paginate"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type MerchantService struct {
	ctx              *context.ERPContext
	db               *gorm.DB
	financeService   *finance.FinanceService
	inventoryService *inventory.InventoryService
}

func NewMerchantService(db *gorm.DB, ctx *context.ERPContext, financeService *finance.FinanceService, inventoryService *inventory.InventoryService) *MerchantService {
	return &MerchantService{db: db, ctx: ctx, financeService: financeService, inventoryService: inventoryService}
}

func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&models.MerchantModel{},
		&models.MerchantTypeModel{},
		&models.MerchantUser{},
		&models.MerchantDesk{},
		&models.MerchantDeskLayout{},
		&models.MerchantOrder{},
		&models.MerchantStation{},
		&models.MerchantStationOrder{},
		&models.MerchantPayment{},
		&models.XenditModel{},
	)
}

func (s *MerchantService) GetNearbyMerchants(lat, lng float64, radius float64) ([]models.MerchantModel, error) {
	var merchants []models.MerchantModel

	rows, err := s.db.Raw(`
		SELECT * FROM (SELECT *, (
			6371 * acos(
				cos(radians(?)) * cos(radians(latitude)) * cos(radians(longitude) - radians(?)) +
				sin(radians(?)) * sin(radians(latitude))
			)
		) AS distance
		FROM pos_merchants
		WHERE status = 'ACTIVE') t
		WHERE distance <= ?
		ORDER BY distance
	`, lat, lng, lat, radius).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var merchant models.MerchantModel
		if err := s.db.ScanRows(rows, &merchant); err != nil {
			return nil, err
		}
		merchants = append(merchants, merchant)
	}
	return merchants, err
}

func (s *MerchantService) CreateMerchant(data *models.MerchantModel) error {
	if data.MerchantTypeID != nil {
		var merchantType models.MerchantTypeModel
		err := s.db.Where("id = ?", data.MerchantTypeID).First(&merchantType).Error
		if err != nil {
			return err
		}
		data.MerchantType = &merchantType.Name
	}
	return s.db.Create(data).Error
}

func (s *MerchantService) UpdateMerchant(id string, data *models.MerchantModel) error {
	if data.MerchantTypeID != nil {
		var merchantType models.MerchantTypeModel
		err := s.db.Where("id = ?", data.MerchantTypeID).First(&merchantType).Error
		if err != nil {
			return err
		}
		data.MerchantType = &merchantType.Name
	}
	return s.db.Where("id = ?", id).Omit("Xendit").Updates(data).Error
}

func (s *MerchantService) DeleteMerchant(id string) error {
	return s.db.Where("id = ?", id).Delete(&models.MerchantModel{}).Error
}

func (s *MerchantService) GetMerchantByID(id string) (*models.MerchantModel, error) {
	var merchant models.MerchantModel
	err := s.db.Preload("Company").Preload("DefaultWarehouse").
		Preload("Xendit").
		Preload("DefaultPriceCategory").Preload("User").Preload("Users").Where("id = ?", id).First(&merchant).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("merchant not found")
	}
	file, err := s.GetPicture(id)
	if err == nil {
		merchant.Picture = file
	}
	return &merchant, err
}
func (s *MerchantService) GetActiveMerchantByID(id, companyID string) (*models.MerchantModel, error) {
	var merchant models.MerchantModel
	err := s.db.Preload("Company").Preload("User").Preload("Stations").Preload("DefaultWarehouse").Where("id = ? AND company_id = ?", id, companyID).First(&merchant).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("merchant not found")
	}
	if merchant.Status == "PENDING" {
		return nil, errors.New("merchant is not active")
	}
	if merchant.Status == "SUSPENDED" {
		return nil, errors.New("merchant is suspended")
	}
	file, err := s.GetPicture(id)
	if err == nil {
		merchant.Picture = file
	}

	for i, v := range merchant.Stations {
		var productMerchants []models.ProductMerchant
		s.db.Model(&models.ProductMerchant{}).Where("merchant_model_id = ? AND merchant_station_id = ?", id, v.ID).Find(&productMerchants)
		for _, prod := range productMerchants {
			var product models.ProductModel
			err := s.db.Model(&models.ProductModel{}).Where("id = ?", prod.ProductModelID).First(&product).Error
			if err == nil {
				v.Products = append(v.Products, product)
			}
		}
		merchant.Stations[i] = v
	}

	return &merchant, err
}

func (s *MerchantService) GetPicture(id string) (*models.FileModel, error) {
	var picture models.FileModel
	s.db.Where("ref_id = ? AND ref_type = ?", id, "merchant").Order("created_at DESC").First(&picture)
	return &picture, nil
}

func (s *MerchantService) GetMerchantsByUserID(request http.Request, userID, companyID, search string) (paginate.Page, error) {
	pg := paginate.New()
	stmt := s.db.Preload("Xendit")
	if search != "" {
		stmt = stmt.Where("pos_merchants.description ILIKE ? OR pos_merchants.name ILIKE ?",
			"%"+search+"%",
			"%"+search+"%",
		)
	}
	if request.Header.Get("ID-Company") != "" {
		stmt = stmt.Where("pos_merchants.company_id = ?", request.Header.Get("ID-Company"))
	}
	if request.URL.Query().Get("status") != "" {
		stmt = stmt.Where("pos_merchants.status = ?", request.URL.Query().Get("status"))
	}
	stmt = stmt.Joins("JOIN merchant_users ON merchant_users.merchant_model_id = pos_merchants.id").
		Where("merchant_users.user_model_id = ?", userID)

	stmt = stmt.Where("pos_merchants.company_id = ?", companyID)
	stmt = stmt.Model(&models.MerchantModel{})
	utils.FixRequest(&request)
	page := pg.With(stmt).Request(request).Response(&[]models.MerchantModel{})
	page.Page = page.Page + 1
	items := page.Items.(*[]models.MerchantModel)
	newItems := make([]models.MerchantModel, 0)

	for _, v := range *items {
		if v.CompanyID != nil {
			var company models.CompanyModel
			err := s.db.Select("name", "id").Where("id = ?", v.CompanyID).First(&company).Error
			if err == nil {
				v.Company = &company
			}
		}
		pic, _ := s.GetPicture(v.ID)
		v.Picture = pic
		newItems = append(newItems, v)

	}
	page.Items = &newItems
	return page, nil
}
func (s *MerchantService) GetMerchants(request http.Request, search string) (paginate.Page, error) {
	pg := paginate.New()
	stmt := s.db
	if search != "" {
		stmt = stmt.Where("merchants.description ILIKE ? OR merchants.name ILIKE ?",
			"%"+search+"%",
			"%"+search+"%",
		)
	}
	if request.Header.Get("ID-Company") != "" {
		stmt = stmt.Where("company_id = ?", request.Header.Get("ID-Company"))
	}
	if request.URL.Query().Get("status") != "" {
		stmt = stmt.Where("status = ?", request.URL.Query().Get("status"))
	}
	stmt = stmt.Model(&models.MerchantModel{})
	utils.FixRequest(&request)
	page := pg.With(stmt).Request(request).Response(&[]models.MerchantModel{})
	page.Page = page.Page + 1
	items := page.Items.(*[]models.MerchantModel)
	newItems := make([]models.MerchantModel, 0)

	for _, v := range *items {
		if v.CompanyID != nil {
			var company models.CompanyModel
			err := s.db.Select("name", "id").Where("id = ?", v.CompanyID).First(&company).Error
			if err == nil {
				v.Company = &company
			}
		}
		pic, _ := s.GetPicture(v.ID)
		v.Picture = pic
		newItems = append(newItems, v)

	}
	page.Items = &newItems
	return page, nil
}

func (s *MerchantService) CreateProduct(data *models.ProductModel, merchantID, companyID string) error {
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	data.CompanyID = &companyID
	data.Status = "PENDING"
	err := tx.Create(data).Error
	if err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Model(&models.ProductMerchant{}).Where("product_model_id = ? AND merchant_model_id = ?", data.ID, merchantID).FirstOrCreate(&models.ProductMerchant{
		ProductModelID:  data.ID,
		MerchantModelID: merchantID,
		Price:           data.Price,
		AdjustmentPrice: data.AdjustmentPrice,
	}).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

func (s *MerchantService) GetMerchantProductDetail(id, merchantID string, warehouseID *string) (*models.ProductModel, error) {
	var merchant models.MerchantModel
	s.db.Select("default_warehouse_id").Where("id = ?", merchantID).First(&merchant)
	if warehouseID == nil {
		warehouseID = merchant.DefaultWarehouseID
	}

	var product models.ProductModel
	err := s.db.Select("products.*", "product_merchants.price as price").Preload("Variants.Attributes.Attribute").Preload("Brand").Preload("Tags").
		Joins("JOIN product_merchants ON product_merchants.product_model_id = products.id").
		Joins("JOIN brands ON brands.id = products.brand_id").
		Joins("LEFT JOIN product_variants ON product_variants.product_id = products.id").
		Where("product_merchants.merchant_model_id = ?", merchantID).
		First(&product, "products.id = ?", id).Error

	if err != nil {
		return nil, err
	}

	if warehouseID != nil {
		totalStock, _ := s.inventoryService.StockMovementService.GetCurrentStock(id, *warehouseID)
		product.TotalStock = totalStock
	}
	product.MerchantID = &merchantID

	for j, variant := range product.Variants {
		variant.MerchantID = &merchantID
		variant.GetPriceAndDiscount(s.db)
		if warehouseID != nil {
			totalVariantStock, _ := s.inventoryService.StockMovementService.GetVarianCurrentStock(id, variant.ID, *warehouseID)
			variant.TotalStock = totalVariantStock
			variant.Price = s.inventoryService.ProductService.GetVariantPrice(merchantID, &variant)
			product.Variants[j] = variant
		}
	}

	var ProductMerchant models.ProductMerchant
	err = s.db.Select("last_updated_stock", "last_stock", "price").Where("product_model_id = ? AND merchant_model_id = ?", id, merchantID).First(&ProductMerchant).Error
	if err == nil {
		product.LastUpdatedStock = ProductMerchant.LastUpdatedStock
		product.LastStock = ProductMerchant.LastStock
		product.Price = ProductMerchant.Price
		product.GetPriceAndDiscount(s.db)
	}

	return &product, nil
}
func (s *MerchantService) GetMerchantProducts(request http.Request, search string, merchantID string, warehouseID *string, status []string, useBrand bool) (paginate.Page, error) {
	pg := paginate.New()

	var products []models.ProductModel
	var merchant models.MerchantModel
	s.db.Select("default_warehouse_id").Where("id = ?", merchantID).First(&merchant)
	if warehouseID == nil {
		warehouseID = merchant.DefaultWarehouseID
	}

	stmt := s.db.Joins("JOIN product_merchants ON product_merchants.product_model_id = products.id").Preload("Category").
		Joins("LEFT JOIN product_variants ON product_variants.product_id = products.id")

	if useBrand {
		stmt = stmt.Joins("JOIN brands ON brands.id = products.brand_id")
	}
	stmt = stmt.Where("product_merchants.merchant_model_id = ?", merchantID)
	if search != "" {
		if useBrand {
			stmt = stmt.Where("products.name ILIKE ? OR products.sku ILIKE ? OR products.description ILIKE ? OR brands.name ILIKE ? OR product_variants.display_name ILIKE ?",
				"%"+search+"%",
				"%"+search+"%",
				"%"+search+"%",
				"%"+search+"%",
				"%"+search+"%")
		} else {
			stmt = stmt.Where("products.name ILIKE ? OR products.sku ILIKE ? OR products.description ILIKE ? OR product_variants.display_name ILIKE ?",
				"%"+search+"%",
				"%"+search+"%",
				"%"+search+"%",
				"%"+search+"%")
		}

	}
	stmt = stmt.Distinct("products.id")
	stmt = stmt.Select("products.*", "product_merchants.price as price").Preload("Variants.Attributes.Attribute").Preload("Brand").Preload("Tags").Model(&models.ProductModel{})

	if request.URL.Query().Get("brand_id") != "" {
		stmt = stmt.Where("brand_id = ?", request.URL.Query().Get("brand_id"))
	}
	if request.URL.Query().Get("category_id") != "" {
		stmt = stmt.Where("category_id = ?", request.URL.Query().Get("category_id"))
	}

	if len(status) > 0 {
		stmt = stmt.Where("status IN (?)", status)
	}

	utils.FixRequest(&request)
	page := pg.With(stmt).Request(request).Response(&products)
	page.Page = page.Page + 1

	items := page.Items.(*[]models.ProductModel)
	newItems := make([]models.ProductModel, 0)

	for _, v := range *items {
		v.MerchantID = &merchantID
		v.GetPriceAndDiscount(s.db)
		img, err := s.inventoryService.ProductService.ListImagesOfProduct(v.ID)
		if err == nil {
			v.ProductImages = img
		}
		if warehouseID != nil {
			totalStock, _ := s.inventoryService.StockMovementService.GetCurrentStock(v.ID, *warehouseID)
			v.TotalStock = totalStock
		}
		for j, variant := range v.Variants {
			variant.MerchantID = &merchantID
			variant.GetPriceAndDiscount(s.db)
			if warehouseID != nil {
				totalVariantStock, _ := s.inventoryService.StockMovementService.GetVarianCurrentStock(v.ID, variant.ID, *warehouseID)
				variant.TotalStock = totalVariantStock
				// variant.Price = s.inventoryService.ProductService.GetVariantPrice(merchantID, &variant)
				v.Variants[j] = variant
			}
		}

		fmt.Println("GET MERCHANT PRODUCT", v.ID, merchantID)
		var ProductMerchant models.ProductMerchant
		err = s.db.Select("last_updated_stock", "last_stock", "price", "merchant_station_id").Where("product_model_id = ? AND merchant_model_id = ?", v.ID, merchantID).First(&ProductMerchant).Error
		if err == nil {
			v.LastUpdatedStock = ProductMerchant.LastUpdatedStock
			v.LastStock = ProductMerchant.LastStock
			v.Price = ProductMerchant.Price
			v.AdjustmentPrice = ProductMerchant.AdjustmentPrice
			v.MerchantStationID = ProductMerchant.MerchantStationID
			fmt.Println("GET MERCHANT PRODUCT #1", v.ID, merchantID, ProductMerchant.MerchantStationID)
		}
		prices, err := s.ListPricesOfProduct(v.ID)
		if err == nil {
			v.Prices = prices
		}

		newItems = append(newItems, v)
	}
	page.Items = &newItems

	return page, nil
}

func (s *MerchantService) CountMerchantByStatus(status string) (int64, error) {

	var count int64
	if err := s.db.Model(&models.MerchantModel{}).Where("status = ?", status).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (s *MerchantService) AddProductsToMerchant(merchantID string, productIDs []string) error {

	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	for _, productID := range productIDs {
		var product models.ProductModel
		if err := tx.Select("id", "price").Where("id = ?", productID).First(&product).Error; err != nil {
			return err
		}

		if err := tx.Model(&models.ProductMerchant{}).Where("product_model_id = ? AND merchant_model_id = ?", productID, merchantID).FirstOrCreate(&models.ProductMerchant{
			ProductModelID:  productID,
			MerchantModelID: merchantID,
			Price:           product.OriginalPrice,
		}).Error; err != nil {
			tx.Rollback()
			return err
		}

	}

	return tx.Commit().Error
}

func (s *MerchantService) DeleteProductsFromMerchant(merchantID string, productIDs []string) error {
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := tx.Where("merchant_model_id = ? AND product_model_id IN (?)", merchantID, productIDs).
		Delete(&models.ProductMerchant{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

func (s *MerchantService) EditProductPrice(merchantID, productID string, price float64) error {
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := tx.Model(&models.ProductMerchant{}).Where("product_model_id = ? AND merchant_model_id = ?", productID, merchantID).
		Updates(map[string]interface{}{
			"adjustment_price": price,
		}).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

func (s *MerchantService) EditVariantPrice(merchantID, variantID string, price float64) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		var variantMerchant models.VarianMerchant

		err := tx.Where("variant_id = ? AND merchant_id = ?", variantID, merchantID).
			First(&variantMerchant).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			var variant models.VariantModel
			err = tx.Where("id = ?", variantID).First(&variant).Error
			if err != nil {
				return err
			}
			variantMerchant.VariantID = variantID
			variantMerchant.MerchantID = merchantID
			variantMerchant.Price = variant.OriginalPrice
			variantMerchant.AdjustmentPrice = price
			tx.Create(&variantMerchant)
		}

		if err := tx.Model(&models.VarianMerchant{}).Where("variant_id = ? AND merchant_id = ?", variantID, merchantID).
			Updates(map[string]interface{}{
				"adjustment_price": price,
			}).Error; err != nil {
			return err
		}
		return nil
	})

}

func (s *MerchantService) GetProductAvailableByMerchant(merchant models.MerchantModel, orderRequest *models.OrderRequestModel) (*models.MerchantAvailableProduct, error) {
	var subTotal, totalDiscAmount, subTotalBeforeDiscount float64
	var merchantAvailable models.MerchantAvailableProduct
	merchantAvailable.MerchantID = merchant.ID
	merchantAvailable.Name = merchant.Name
	merchantAvailable.Items = make([]models.MerchantAvailableProductItem, len(orderRequest.Items))
	for i, item := range orderRequest.Items {

		fmt.Println("CHECK STOCK", *item.ProductID, *item.VariantID, *merchant.DefaultWarehouseID)
		var product models.ProductModel
		s.db.Select("price", "id", "display_name").Find(&product, "id = ?", item.ProductID)
		var productDisplayName string = product.DisplayName
		var availableStock float64
		// price = item.UnitPrice
		var variantDisplayName *string
		if item.VariantID != nil {
			var variant models.VariantModel
			s.db.Select("price", "id", "display_name").Find(&variant, "id = ?", *item.VariantID)
			availableStock, _ = s.inventoryService.ProductService.GetVariantStock(*item.ProductID, *item.VariantID, nil, merchant.DefaultWarehouseID)
			variantDisplayName = &variant.DisplayName
		} else {
			availableStock, _ = s.inventoryService.ProductService.GetStock(*item.ProductID, nil, merchant.DefaultWarehouseID)
		}

		// _, discAmount, discValue, discType, err := s.inventoryService.ProductService.CalculateDiscountedPrice(*item.ProductID, price)
		// if err != nil {
		// 	return nil, err
		// }

		fmt.Println("AVAILABLE STOCK", merchant.Name, *item.ProductID, *item.VariantID, *merchant.DefaultWarehouseID, availableStock)
		if availableStock < item.Quantity {
			item.Status = "OUT_OF_STOCK"
		} else {
			item.Status = "AVAILABLE"
			subTotal += item.Total
			subTotalBeforeDiscount += item.TotalBeforeDiscount
		}
		totalDiscAmount += item.DiscountAmount
		// orderRequest.Items[i] = item
		merchantAvailable.Items[i] = models.MerchantAvailableProductItem{
			ProductID:               *item.ProductID,
			ProductDisplayName:      productDisplayName,
			VariantDisplayName:      variantDisplayName,
			VariantID:               item.VariantID,
			Quantity:                item.Quantity,
			UnitPrice:               item.UnitPrice,
			UnitPriceBeforeDiscount: item.OriginalPrice,
			Status:                  item.Status,
			SubTotalBeforeDiscount:  item.TotalBeforeDiscount,
			SubTotal:                item.Total,
			DiscountAmount:          item.DiscountAmount,
			DiscountValue:           item.DiscountPercent,
			DiscountType:            item.DiscountType,
		}

	}
	merchantAvailable.SubTotal = subTotal
	merchantAvailable.SubTotalBeforeDiscount = subTotalBeforeDiscount
	merchantAvailable.OrderRequestID = orderRequest.ID
	merchantAvailable.TotalDiscountAmount = totalDiscAmount
	return &merchantAvailable, nil
}

func (s *MerchantService) GetPhoneNumberFromMerchantID(merchantID string) ([]string, error) {
	var merchant models.MerchantModel
	err := s.db.Preload("Company.Users").Find(&merchant, "id = ?", merchantID).Error
	if err != nil {
		return nil, err
	}

	phoneNumbers := make([]string, 0)
	for _, v := range merchant.Company.Users {
		if v.PhoneNumber != nil {
			phoneNumbers = append(phoneNumbers, *v.PhoneNumber)
		}
	}

	return phoneNumbers, nil
}
func (s *MerchantService) GetPushTokenFromMerchantID(merchantID string) ([]string, error) {
	var merchant models.MerchantModel
	err := s.db.Preload("Company.Users").Find(&merchant, "id = ?", merchantID).Error
	if err != nil {
		return nil, err
	}

	userIDs := make([]string, 0)
	for _, v := range merchant.Company.Users {
		userIDs = append(userIDs, v.ID)
	}
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

func (s *MerchantService) CreateMerchantType(data *models.MerchantTypeModel) error {

	err := s.db.Create(&data).Error
	if err != nil {
		return err
	}

	return nil
}

func (s *MerchantService) GetAllMerchantType(request http.Request, search string) (paginate.Page, error) {
	pg := paginate.New()
	var merchantType []models.MerchantTypeModel

	stmt := s.db

	if search != "" {
		stmt = stmt.Where("name ILIKE ? OR description ILIKE ? ",
			"%"+search+"%",
			"%"+search+"%")
	}

	stmt = stmt.Model(&models.MerchantTypeModel{})
	utils.FixRequest(&request)
	page := pg.With(stmt).Request(request).Response(&merchantType)
	page.Page = page.Page + 1

	return page, nil
}

func (s *MerchantService) GetMerchantType(id string) (*models.MerchantTypeModel, error) {
	var merchantType models.MerchantTypeModel
	err := s.db.Where("id = ?", id).First(&merchantType).Error
	if err != nil {
		return nil, err
	}

	return &merchantType, nil
}

func (s *MerchantService) UpdateMerchantType(data *models.MerchantTypeModel) error {

	err := s.db.Save(data).Error
	if err != nil {
		return err
	}

	return nil
}

func (s *MerchantService) DeleteMerchantType(id string) error {
	err := s.db.Delete(&models.MerchantTypeModel{}, "id = ?", id).Error
	if err != nil {
		return err
	}

	return nil
}

func (s *MerchantService) UpdateProduct(id, merchantID, companyID string, data *models.ProductModel) error {
	var products []models.ProductModel

	err := s.db.First(&products, "id = ? and company_id = ?", id, companyID).Error
	if err != nil {
		return err
	}
	data.ID = id
	data.CompanyID = &companyID

	if err := s.db.Model(&models.ProductMerchant{}).Where("product_model_id = ? AND merchant_model_id = ?", data.ID, merchantID).Updates(&models.ProductMerchant{
		ProductModelID:  data.ID,
		MerchantModelID: merchantID,
		Price:           data.Price,
	}).Error; err != nil {
		return err
	}

	return s.db.Omit(clause.Associations).Save(data).Error
}

func (s *MerchantService) GetProductByID(id string, request *http.Request) (*models.ProductModel, error) {
	idCompany := ""
	idMerchant := ""
	if request != nil {
		if request.Header != nil {
			idCompany = request.Header.Get("ID-Company")
			idMerchant = request.Header.Get("ID-Merchant")
		}
	} else {
		return nil, errors.New("request is nil")
	}
	var product models.ProductModel
	err := s.db.Preload("Tags").Preload("Variants").Preload("MasterProduct").Preload("Category", func(db *gorm.DB) *gorm.DB {
		return db.Select("id", "name")
	}).Preload("Brand", func(db *gorm.DB) *gorm.DB {
		return db.Select("id", "name")
	}).Where("id = ? and company_id = ?", id, idCompany).First(&product).Error
	product.Prices, _ = s.inventoryService.ProductService.ListPricesOfProduct(product.ID)
	product.ProductImages, _ = s.inventoryService.ProductService.ListImagesOfProduct(product.ID)
	var warehouseID *string
	if request != nil {
		warehouseIDStr := request.Header.Get("ID-Warehouse")
		if warehouseIDStr != "" {
			warehouseID = &warehouseIDStr
		}
	}
	stock, _ := s.inventoryService.ProductService.GetStock(product.ID, request, warehouseID)
	if idMerchant != "" {
		product.MerchantID = &idMerchant
	}
	product.GetPriceAndDiscount(s.db)
	product.TotalStock = stock
	for i, v := range product.Variants {
		if idMerchant != "" {
			v.MerchantID = &idMerchant
		}
		v.GetPriceAndDiscount(s.db)
		variantStock, _ := s.inventoryService.ProductService.GetVariantStock(product.ID, v.ID, request, warehouseID)
		v.TotalStock = variantStock
		product.Variants[i] = v
		fmt.Println("VARIANT STOCK", v.ID, variantStock)
	}

	return &product, err
}

func (s *MerchantService) GetSalesCountByBrand(request *http.Request, merchantID, warehouseID, companyID, distributorID *string, startDate, endDate *time.Time) ([]map[string]interface{}, error) {
	salesCountByBrand := make([]map[string]interface{}, 0)
	db := s.db.Table("stock_movements")

	if warehouseID != nil {
		db = db.Where("stock_movements.warehouse_id = ?", *warehouseID)
	}
	if merchantID != nil {
		db = db.Where("stock_movements.merchant_id = ?", *merchantID)
	}

	if distributorID != nil {
		db = db.Where("stock_movements.distributor_id = ?", *distributorID)
	}
	if companyID != nil {
		db = db.Where("stock_movements.company_id = ?", *companyID)
	}
	if startDate != nil {
		db = db.Where("stock_movements.created_at >= ?", startDate)
	}
	if endDate != nil {
		db = db.Where("stock_movements.created_at <= ?", endDate)
	}

	db = db.Where("type in (?)", []models.MovementType{models.MovementTypeReturn, models.MovementTypeSale})
	rows, err := db.Joins("JOIN products ON products.id = stock_movements.product_id").
		Joins("JOIN brands ON brands.id = products.brand_id").
		Select("products.brand_id, brands.name, COALESCE(SUM(quantity), 0) as total_quantity").
		Group("products.brand_id").
		Group("brands.name").
		Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var brandID string
		var brandName string
		var totalStock float64
		if err := rows.Scan(&brandID, &brandName, &totalStock); err != nil {
			return nil, err
		}
		salesCountByBrand = append(salesCountByBrand, map[string]interface{}{
			"brand_id": brandID,
			"name":     brandName,
			"total":    -totalStock,
		})
	}

	return salesCountByBrand, nil
}

func (s *MerchantService) GetSalesCountByBrandAndOrderType(request *http.Request, merchantID, warehouseID, companyID, distributorID *string) (interface{}, error) {
	salesCountByBrand := make([]map[string]interface{}, 0)
	db := s.db.Table("stock_movements")

	if warehouseID != nil {
		db = db.Where("stock_movements.warehouse_id = ?", *warehouseID)
	}
	if merchantID != nil {
		db = db.Where("stock_movements.merchant_id = ?", *merchantID)
	}

	if distributorID != nil {
		db = db.Where("stock_movements.distributor_id = ?", *distributorID)
	}
	if companyID != nil {
		db = db.Where("stock_movements.company_id = ?", *companyID)
	}
	db = db.Where("type in (?)", []models.MovementType{models.MovementTypeReturn, models.MovementTypeSale})
	rows, err := db.Joins("JOIN products ON products.id = stock_movements.product_id").
		Joins("JOIN brands ON brands.id = products.brand_id").
		Joins("JOIN pos_sales ON pos_sales.id = stock_movements.reference_id").
		Select("products.brand_id, brands.name, pos_sales.order_type, COALESCE(SUM(quantity), 0) as total_quantity").
		Group("products.brand_id").
		Group("pos_sales.order_type").
		Group("brands.name").
		Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var brandID string
		var brandName string
		var orderType string
		var totalStock float64
		if err := rows.Scan(&brandID, &brandName, &orderType, &totalStock); err != nil {
			return nil, err
		}
		salesCountByBrand = append(salesCountByBrand, map[string]interface{}{
			"brand_id":   brandID,
			"name":       brandName,
			"order_type": orderType,
			"total":      -totalStock,
		})
	}

	grouped := make(map[string][]map[string]interface{})
	for _, item := range salesCountByBrand {
		orderType, ok := item["order_type"]
		if !ok {
			continue
		}
		grouped[orderType.(string)] = append(grouped[orderType.(string)], item)
	}
	// salesCountByBrand = make([]map[string]interface{}, 0, len(grouped))
	// for _, value := range grouped {
	// 	salesCountByBrand = append(salesCountByBrand, value...)
	// }
	return grouped, nil
}

func (s *MerchantService) GetMerchantUsers(request http.Request, search, merchantID, companyID string) (paginate.Page, error) {
	pg := paginate.New()

	var users []models.UserModel

	stmt := s.db.Joins("JOIN merchant_users ON merchant_users.user_model_id = users.id")

	stmt = stmt.Where("merchant_users.merchant_model_id = ?", merchantID)
	if search != "" {
		stmt = stmt.Where("users.full_name ILIKE ? OR users.email ILIKE ? OR users.phone ILIKE ?",
			"%"+search+"%",
			"%"+search+"%",
			"%"+search+"%")

	}
	stmt = stmt.Distinct("users.id")
	stmt = stmt.Select("users.*").Preload("Roles", "roles.company_id = ?", companyID).Model(&models.UserModel{})

	utils.FixRequest(&request)
	page := pg.With(stmt).Request(request).Response(&users)
	page.Page = page.Page + 1

	return page, nil
}

func (s *MerchantService) AddMerchantUser(merchantID string, userIDs []string) error {
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	for _, userID := range userIDs {

		if err := tx.Model(&models.MerchantUser{}).Where("user_model_id = ? AND merchant_model_id = ?", userID, merchantID).FirstOrCreate(&models.MerchantUser{
			UserModelID:     userID,
			MerchantModelID: merchantID,
		}).Error; err != nil {
			tx.Rollback()
			return err
		}

	}

	return tx.Commit().Error
}

func (s *MerchantService) DeleteUserFromMerchant(merchantID string, userIDs []string) error {
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := tx.Where("merchant_model_id = ? AND user_model_id IN (?)", merchantID, userIDs).
		Delete(&models.MerchantUser{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

func (s *MerchantService) ListPricesOfProduct(productID string) ([]models.PriceModel, error) {
	var prices []models.PriceModel
	err := s.db.Preload("PriceCategory").Where("product_id = ?", productID).Find(&prices).Error
	return prices, err
}

func (s *MerchantService) GetDesksFromID(request http.Request, merchantID string) (paginate.Page, error) {
	pg := paginate.New()
	var desks []models.MerchantDesk
	stmt := s.db.Where("merchant_id = ?", merchantID)
	if request.URL.Query().Get("search") != "" {
		stmt = stmt.Where("desk_name LIKE ?", "%"+request.URL.Query().Get("search")+"%")
	}
	stmt = stmt.Order("order_number").Model(&models.MerchantDesk{})
	utils.FixRequest(&request)
	page := pg.With(stmt).Request(request).Response(&desks)
	page.Page = page.Page + 1
	return page, nil
}

func (s *MerchantService) AddDeskToMerchant(merchantID string, desk *models.MerchantDesk) error {

	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := tx.Model(&models.MerchantDesk{}).FirstOrCreate(desk).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

func (s *MerchantService) UpdateMerchantDesk(merchantID string, deskId string, desk *models.MerchantDesk) error {
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := tx.Model(&models.MerchantDesk{}).Where("merchant_id = ? AND id = ?", merchantID, deskId).Updates(desk).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

func (s *MerchantService) DeleteDeskFromMerchant(merchantID string, deskID string) error {
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := tx.Where("merchant_id = ? AND id = ?", merchantID, deskID).
		Delete(&models.MerchantDesk{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

func (s *MerchantService) GetLayoutDetailFromID(merchantID string, layoutID string) (*models.MerchantDeskLayout, error) {
	var layout models.MerchantDeskLayout
	if err := s.db.Preload("MerchantDesks").Where("merchant_id = ? AND id = ?", merchantID, layoutID).First(&layout).Error; err != nil {
		return nil, err
	}
	return &layout, nil
}
func (s *MerchantService) GetLayoutsFromID(request http.Request, merchantID string) (paginate.Page, error) {
	pg := paginate.New()
	var layouts []models.MerchantDeskLayout
	stmt := s.db.Preload("MerchantDesks").Where("merchant_id = ?", merchantID)
	if request.URL.Query().Get("search") != "" {
		stmt = stmt.Where("layout_name LIKE ?", "%"+request.URL.Query().Get("search")+"%")
	}
	stmt = stmt.Order("created_at").Model(&models.MerchantDeskLayout{})
	utils.FixRequest(&request)
	page := pg.With(stmt).Request(request).Response(&layouts)
	page.Page = page.Page + 1
	return page, nil
}
func (s *MerchantService) AddLayoutToMerchant(merchantID string, layout *models.MerchantDeskLayout) error {
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := tx.Model(&models.MerchantDeskLayout{}).FirstOrCreate(layout).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

func (s *MerchantService) UpdateLayoutMerchant(merchantID string, layoutId string, layout *models.MerchantDeskLayout) error {
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := tx.Model(&models.MerchantDeskLayout{}).Where("merchant_id = ? AND id = ?", merchantID, layoutId).Updates(layout).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

func (s *MerchantService) DeleteLayoutMerchant(merchantID string, layoutID string) error {
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := tx.Where("merchant_id = ? AND id = ?", merchantID, layoutID).
		Delete(&models.MerchantDeskLayout{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

func (s *MerchantService) UpdateTableStatus(merchantID string, tableID string, status string) error {
	var table models.MerchantDesk
	if err := s.db.Model(&table).Where("merchant_id = ? AND id = ?", merchantID, tableID).First(&table).Error; err != nil {
		return err
	}
	if strings.ToUpper(*table.Status) != "AVAILABLE" {
		return errors.New("table status must be available")
	}
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := tx.Model(&models.MerchantDesk{}).Where("merchant_id = ? AND id = ?", merchantID, tableID).
		Update("status", strings.ToUpper(status)).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

func (s *MerchantService) GetMerchantStations(request http.Request, merchantID string) (paginate.Page, error) {
	pg := paginate.New()
	var search = request.URL.Query().Get("search")
	stmt := s.db
	if search != "" {
		stmt = stmt.Where("station_name ILIKE ? OR description ILIKE ?",
			"%"+search+"%",
			"%"+search+"%",
		)
	}

	request.URL.Query().Get("page")
	stmt = stmt.Model(&models.MerchantStation{})
	utils.FixRequest(&request)
	page := pg.With(stmt).Request(request).Response(&[]models.MerchantStation{})
	page.Page = page.Page + 1
	return page, nil
}

func (s *MerchantService) GetMerchantStationDetail(merchantID string, stationID string) (*models.MerchantStation, error) {
	var station models.MerchantStation
	if err := s.db.Where("merchant_id = ? AND id = ?", merchantID, stationID).First(&station).Error; err != nil {
		return nil, err
	}

	var productMerchants []models.ProductMerchant
	s.db.Model(&models.ProductMerchant{}).Where("merchant_model_id = ? AND merchant_station_id = ?", merchantID, stationID).Find(&productMerchants)
	for _, prod := range productMerchants {
		var product models.ProductModel
		err := s.db.Model(&models.ProductModel{}).Where("id = ?", prod.ProductModelID).First(&product).Error
		if err == nil {
			station.Products = append(station.Products, product)
		}
	}
	return &station, nil
}

func (s *MerchantService) GetOrdersFromStation(request http.Request, merchantID string, stationID string, status []string) (paginate.Page, error) {
	pg := paginate.New()
	var orders []models.MerchantStationOrder
	stmt := s.db.Preload("MerchantStation").Preload("Order.MerchantDesk").Model(&models.MerchantStationOrder{}).Where("merchant_station_id = ? AND status IN (?)", stationID, status)
	stmt = stmt.Joins("JOIN merchant_stations ON merchant_stations.id = merchant_station_orders.merchant_station_id")
	stmt = stmt.Order("merchant_station_orders.created_at DESC")
	utils.FixRequest(&request)
	page := pg.With(stmt).Request(request).Response(&orders)
	page.Page = page.Page + 1
	return page, nil
}

func (s *MerchantService) GetMerchantOrderStation(orderStationID string, stationID string) (*models.MerchantStationOrder, error) {
	var orderStation models.MerchantStationOrder
	if err := s.db.Where("id = ? AND  merchant_station_id = ?", orderStationID, stationID).First(&orderStation).Error; err != nil {
		return nil, err
	}
	return &orderStation, nil
}

func (s *MerchantService) GetProductsFromMerchantStation(merchantID string, stationID string) ([]models.ProductModel, error) {
	productMerchants := []models.ProductMerchant{}
	if err := s.db.Where("merchant_model_id = ? AND merchant_station_id = ?", merchantID, stationID).Find(&productMerchants).Error; err != nil {
		return nil, err
	}

	products := []models.ProductModel{}
	for _, prodMerchant := range productMerchants {
		var product models.ProductModel
		if err := s.db.Where("id = ?", prodMerchant.ProductModelID).First(&product).Error; err == nil {
			img, err := s.inventoryService.ProductService.ListImagesOfProduct(product.ID)
			if err == nil {
				product.ProductImages = img
			}
			products = append(products, product)
		}
	}

	return products, nil
}

func (s *MerchantService) CreateMerchantStation(merchantID string, station *models.MerchantStation) error {
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	station.MerchantID = &merchantID
	station.ID = uuid.New().String()

	if err := tx.Create(station).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

func (s *MerchantService) UpdateMerchantStation(merchantID string, stationID string, station *models.MerchantStation) error {
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := tx.Model(&models.MerchantStation{}).Where("merchant_id = ? AND id = ?", merchantID, stationID).Updates(station).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

func (s *MerchantService) DeleteMerchantStation(merchantID string, stationID string) error {
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := tx.Where("merchant_id = ? AND id = ?", merchantID, stationID).Delete(&models.MerchantStation{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

func (s *MerchantService) AddProductsToMerchantStation(merchantID string, stationID string, productIDs []string) error {
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	for _, pid := range productIDs {
		var productMerchant models.ProductMerchant
		if err := tx.Model(&models.ProductMerchant{}).Where("product_model_id = ? AND merchant_model_id = ?", pid, merchantID).
			First(&productMerchant).Error; err == nil {
			productMerchant.MerchantStationID = &stationID
			if err := tx.Save(&productMerchant).Error; err != nil {
				tx.Rollback()
				return err
			}
		}
	}

	return tx.Commit().Error
}

func (s *MerchantService) DeleteProductFromMerchantStation(merchantID string, stationID string, productIDs []string) error {
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	for _, pid := range productIDs {
		if err := tx.Model(&models.ProductMerchant{}).Where("product_model_id = ? AND merchant_model_id = ?", pid, merchantID).Update("merchant_station_id", nil).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit().Error
}

func (s *MerchantService) CreateOrder(merchantID string, order *models.MerchantOrder) error {
	var existingOrder models.MerchantOrder
	err := s.db.Where("merchant_id = ? AND merchant_desk_id = ? AND order_status = ?", merchantID, order.MerchantDeskID, "ACTIVE").First(&existingOrder).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		order.ID = utils.Uuid()
		order.MerchantID = &merchantID
		order.OrderStatus = "ACTIVE"
		order.Code = strings.ToUpper(utils.GenerateRandomString(6))
		return s.db.Create(order).Error

	}
	existingItems := parseItems(existingOrder.Items)
	newItems := parseItems(order.Items)
	existingItems = append(existingItems, newItems...)
	b, _ := json.Marshal(existingItems)
	existingOrder.Items = b
	order.ID = existingOrder.ID
	var total, subTotal float64
	for _, v := range existingItems {
		total += v.Subtotal
		subTotal += v.SubtotalBeforeDisc
	}

	existingOrder.Total = total
	existingOrder.SubTotal = subTotal
	return s.db.Save(&existingOrder).Error
}

func parseItems(orderItems json.RawMessage) []models.MerchantOrderItem {
	items := []models.MerchantOrderItem{}

	err := json.Unmarshal(orderItems, &items)
	if err != nil {
		return []models.MerchantOrderItem{}
	}
	return items
}
func (s *MerchantService) DistributeOrder(merchantID string, order *models.MerchantOrder) ([]models.MerchantStationOrder, error) {

	items := []models.MerchantOrderItem{}

	err := json.Unmarshal(order.Items, &items)
	if err != nil {
		return nil, err
	}
	// fmt.Println(string(order.Items))
	// utils.LogJson(items)
	orderStations := []models.MerchantStationOrder{}
	for _, v := range items {
		var productMerchant models.ProductMerchant
		if err := s.db.Model(&productMerchant).Where("product_model_id = ? AND merchant_model_id = ?", v.ProductID, merchantID).First(&productMerchant).Error; err != nil {
			return nil, err
		}

		if productMerchant.MerchantStationID != nil {
			itemStation, err := json.Marshal(v)
			if err != nil {
				return nil, err
			}
			var orderStation models.MerchantStationOrder = models.MerchantStationOrder{
				MerchantStationID: productMerchant.MerchantStationID,
				OrderID:           order.ID,
				Status:            "PENDING",
				Item:              itemStation,
				MerchantDeskID:    order.MerchantDeskID,
			}
			orderStation.ID = utils.Uuid()
			if err := s.db.Create(&orderStation).Error; err != nil {
				return nil, err
			}
			orderStations = append(orderStations, orderStation)

		}

	}

	return orderStations, nil
}

func (s *MerchantService) GetOrderDetail(merchantID string, orderID string) (*models.MerchantOrder, error) {
	var order models.MerchantOrder
	if err := s.db.Preload("MerchantDesk").
		Preload("Payments").
		Preload("MerchantStationOrders.MerchantStation").Preload("Contact").Where("id = ? AND merchant_id = ?", orderID, merchantID).First(&order).Error; err != nil {
		return nil, err
	}

	orderStationMap := make(map[string][]models.MerchantStationOrder)
	for _, v := range order.MerchantStationOrders {
		orderStationMap[*v.MerchantStationID] = append(orderStationMap[*v.MerchantStationID], v)
	}

	var orderStations []models.MerchantStation
	var station models.MerchantStation
	for stationID, stations := range orderStationMap {
		fmt.Printf("Station ID: %s\n", stationID)

		if len(stations) > 0 {
			station = *stations[0].MerchantStation
		}
		station.Orders = stations
		orderStations = append(orderStations, station)
	}

	order.MerchantStations = orderStations
	return &order, nil
}
func (s *MerchantService) GetOrders(request http.Request, merchantID string) (paginate.Page, error) {
	pg := paginate.New()

	var orders []models.MerchantOrder

	stmt := s.db.Preload("MerchantDesk").Preload("Contact").Where("merchant_id = ?", merchantID).Model(&models.MerchantOrder{})

	utils.FixRequest(&request)
	page := pg.With(stmt).Request(request).Response(&orders)
	page.Page = page.Page + 1

	return page, nil
}

func (s *MerchantService) UpdateStationOrderStatus(stationID, stationOrderID string, status string) error {
	var orderStation models.MerchantStationOrder
	if err := s.db.Model(&models.MerchantStationOrder{}).Where("merchant_station_id = ? AND id = ?", stationID, stationOrderID).First(&orderStation).Error; err != nil {
		return err
	}
	orderStation.Status = status
	if err := s.db.Save(&orderStation).Error; err != nil {
		return err
	}
	return nil
}

// func (s *MerchantService) MerchantOrderPayment(orderID string, payment *models.MerchantOrderPayment) error {
// 	tx := s.db.Begin()
// 	defer func() {
// 		if r := recover(); r != nil {
// 			tx.Rollback()
// 		}
// 	}()

// 	payment.ID = utils.Uuid()
// 	payment.OrderID = orderID

// 	if err := tx.Create(payment).Error; err != nil {
// 		tx.Rollback()
// 		return err
// 	}

// 	return tx.Commit().Error
// }
