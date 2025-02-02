package merchant

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/finance"
	"github.com/AMETORY/ametory-erp-modules/inventory"
	"github.com/AMETORY/ametory-erp-modules/shared/models"
	"github.com/AMETORY/ametory-erp-modules/utils"
	"github.com/morkid/paginate"
	"gorm.io/gorm"
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
	return db.AutoMigrate(&models.MerchantModel{}, &models.MerchantTypeModel{})
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
	return s.db.Where("id = ?", id).Updates(data).Error
}

func (s *MerchantService) DeleteMerchant(id string) error {
	return s.db.Where("id = ?", id).Delete(&models.MerchantModel{}).Error
}

func (s *MerchantService) GetMerchantByID(id string) (*models.MerchantModel, error) {
	var invoice models.MerchantModel
	err := s.db.Preload("Company").Preload("User").Where("id = ?", id).First(&invoice).Error
	return &invoice, err
}
func (s *MerchantService) GetActiveMerchantByID(id string) (*models.MerchantModel, error) {
	var invoice models.MerchantModel
	err := s.db.Preload("Company").Preload("User").Preload("DefaultWarehouse").Where("id = ? ", id).First(&invoice).Error
	if invoice.Status == "PENDING" {
		return nil, errors.New("merchant is not active")
	}
	if invoice.Status == "SUSPENDED" {
		return nil, errors.New("merchant is suspended")
	}
	return &invoice, err
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
		newItems = append(newItems, v)

	}
	page.Items = &newItems
	return page, nil
}

func (s *MerchantService) GetMerchantProducts(request http.Request, search string, merchantID string, warehouseID *string) (paginate.Page, error) {
	pg := paginate.New()
	var products []models.ProductModel
	var merchant models.MerchantModel
	s.db.Select("default_warehouse_id").Where("id = ?", merchantID).First(&merchant)
	if warehouseID == nil {
		warehouseID = merchant.DefaultWarehouseID
	}

	stmt := s.db.Joins("JOIN product_merchants ON product_merchants.product_model_id = products.id").
		Joins("JOIN brands ON brands.id = products.brand_id").
		Joins("LEFT JOIN product_variants ON product_variants.product_id = products.id").
		Where("product_merchants.merchant_model_id = ?", merchantID)

	if search != "" {
		stmt = stmt.Where("products.name ILIKE ? OR products.sku ILIKE ? OR products.description ILIKE ? OR brands.name ILIKE ? OR product_variants.display_name ILIKE ?",
			"%"+search+"%",
			"%"+search+"%",
			"%"+search+"%",
			"%"+search+"%",
			"%"+search+"%")
	}
	stmt = stmt.Distinct("products.id")
	stmt = stmt.Select("products.*", "product_merchants.price as price").Preload("Variants.Attributes.Attribute").Preload("Brand").Preload("Tags").Model(&models.ProductModel{})

	utils.FixRequest(&request)
	page := pg.With(stmt).Request(request).Response(&products)
	page.Page = page.Page + 1

	items := page.Items.(*[]models.ProductModel)
	newItems := make([]models.ProductModel, 0)

	for _, v := range *items {
		if warehouseID != nil {
			totalStock, _ := s.inventoryService.StockMovementService.GetCurrentStock(v.ID, *warehouseID)
			v.TotalStock = totalStock
		}
		for j, variant := range v.Variants {
			if warehouseID != nil {
				totalVariantStock, _ := s.inventoryService.StockMovementService.GetVarianCurrentStock(v.ID, variant.ID, *warehouseID)
				variant.TotalStock = totalVariantStock
				variant.Price = s.inventoryService.ProductService.GetVariantPrice(merchantID, &variant)
				v.Variants[j] = variant
			}
		}

		var ProductMerchant models.ProductMerchant
		err := s.db.Select("last_updated_stock", "last_stock", "price").Where("product_model_id = ? AND merchant_model_id = ?", v.ID, merchantID).First(&ProductMerchant).Error
		if err == nil {
			v.LastUpdatedStock = ProductMerchant.LastUpdatedStock
			v.LastStock = ProductMerchant.LastStock
			v.Price = ProductMerchant.Price
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
			Price:           product.Price,
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
			"price": price,
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
			variantMerchant.VariantID = variantID
			variantMerchant.MerchantID = merchantID
			variantMerchant.Price = price
			tx.Create(&variantMerchant)
		}

		if err := tx.Model(&models.VarianMerchant{}).Where("variant_id = ? AND merchant_id = ?", variantID, merchantID).
			Updates(map[string]interface{}{
				"price": price,
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
		var availableStock, price float64
		price = product.Price
		var variantDisplayName *string
		if item.VariantID != nil {
			var variant models.VariantModel
			s.db.Select("price", "id", "display_name").Find(&variant, "id = ?", *item.VariantID)
			price = variant.Price
			var variantMerchant models.VarianMerchant
			s.db.Select("price", "id").Find(&variantMerchant, "variant_id = ? AND merchant_id = ?", *item.VariantID, merchant.ID)
			if variantMerchant.Price != 0 {
				price = variantMerchant.Price
			}
			availableStock, _ = s.inventoryService.ProductService.GetVariantStock(*item.ProductID, *item.VariantID, nil, merchant.DefaultWarehouseID)
			variantDisplayName = &variant.DisplayName
		} else {
			availableStock, _ = s.inventoryService.ProductService.GetStock(*item.ProductID, nil, merchant.DefaultWarehouseID)
		}

		_, discAmount, discValue, discType, err := s.inventoryService.ProductService.CalculateDiscountedPrice(*item.ProductID, price)
		if err != nil {
			return nil, err
		}

		fmt.Println("AVAILABLE STOCK", merchant.Name, *item.ProductID, *item.VariantID, *merchant.DefaultWarehouseID, availableStock)
		if availableStock < item.Quantity {
			item.Status = "OUT_OF_STOCK"
		} else {
			item.Status = "AVAILABLE"
			subTotal += item.Quantity * (price - discAmount)
			subTotalBeforeDiscount += item.Quantity * price
		}
		totalDiscAmount += discAmount
		// orderRequest.Items[i] = item
		merchantAvailable.Items[i] = models.MerchantAvailableProductItem{
			ProductID:               *item.ProductID,
			ProductDisplayName:      productDisplayName,
			VariantDisplayName:      variantDisplayName,
			VariantID:               item.VariantID,
			Quantity:                item.Quantity,
			UnitPrice:               price - discAmount,
			UnitPriceBeforeDiscount: price,
			Status:                  item.Status,
			SubTotalBeforeDiscount:  item.Quantity * price,
			SubTotal:                item.Quantity * (price - discAmount),
			DiscountAmount:          discAmount,
			DiscountValue:           discValue,
			DiscountType:            discType,
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
