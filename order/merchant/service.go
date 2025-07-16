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

// NewMerchantService returns a new instance of MerchantService.
//
// db is the database service.
//
// ctx is the application context.
//
// financeService is the finance service.
//
// inventoryService is the inventory service.
func NewMerchantService(db *gorm.DB, ctx *context.ERPContext, financeService *finance.FinanceService, inventoryService *inventory.InventoryService) *MerchantService {
	return &MerchantService{db: db, ctx: ctx, financeService: financeService, inventoryService: inventoryService}
}

// Migrate applies the database schema migrations for the merchant-related models.
// It creates or updates the tables corresponding to each model in the database.

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

// GetNearbyMerchants retrieves a list of nearby merchants.
//
// The method takes three parameters as input: the latitude and longitude of the
// user, and the maximum distance (in kilometers) to search for merchants. It
// returns a slice of MerchantModel that are within the specified radius.
//
// The method uses the Haversine formula to calculate the distance between the
// user's location and each merchant's location. The results are sorted by
// distance, with the closest merchants first.
//
// The method only returns merchants that are currently active.
//
// If the retrieval fails, the method returns an error.
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

// CreateMerchant creates a new merchant in the database.
//
// The function takes a pointer to MerchantModel as its argument. If the
// MerchantTypeID field is set, the function looks up the corresponding
// merchant type name and sets the MerchantType field accordingly.
//
// If the creation fails, the function returns an error.
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

// UpdateMerchant updates an existing merchant in the database.
//
// The function takes two parameters as input: the ID of the merchant to update,
// and a pointer to MerchantModel containing the new data. If the
// MerchantTypeID field is set, the function looks up the corresponding merchant
// type name and sets the MerchantType field accordingly.
//
// If the update fails, the function returns an error.
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

// DeleteMerchant deletes a merchant from the database.
//
// The function takes the ID of the merchant to delete as its parameter.
// If the deletion is successful, it returns nil. Otherwise, it returns an error.

func (s *MerchantService) DeleteMerchant(id string) error {
	return s.db.Where("id = ?", id).Delete(&models.MerchantModel{}).Error
}

// GetMerchantByID retrieves a merchant by its ID.
//
// The function takes the ID of the merchant to be retrieved as a string and returns a pointer to a MerchantModel if the merchant is found,
// otherwise an error is returned if the retrieval fails.
//
// The function also populates the merchant with its picture if it exists.
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

// GetActiveMerchantByID retrieves a merchant by its ID, but only if it is active (i.e. not PENDING or SUSPENDED).
//
// The function takes the ID of the merchant to be retrieved as a string, as well as the ID of the company that the merchant belongs to.
// If the merchant is not found, or if it is not active, the function returns an error.
// If the merchant is active, the function returns a pointer to a MerchantModel.
// The MerchantModel is populated with its picture if it exists, as well as its stations and products.
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

// GetPicture retrieves the latest picture of the merchant with the given ID.
//
// The function takes the ID of the merchant as a string and returns a pointer to a FileModel if the picture is found,
// otherwise an error is returned if the retrieval fails.
func (s *MerchantService) GetPicture(id string) (*models.FileModel, error) {
	var picture models.FileModel
	s.db.Where("ref_id = ? AND ref_type = ?", id, "merchant").Order("created_at DESC").First(&picture)
	return &picture, nil
}

// GetMerchantsByUserID retrieves a list of merchants associated with a specific user ID.
//
// The function takes an http.Request, a userID, a companyID, and a search query string as input. The method uses
// GORM to query the database for merchants, applying the search query to the merchant name and description fields.
// If the request contains a company ID header, the method also filters the result by the company ID. The function
// also filters the result by the given user ID. The function utilizes pagination to manage the result set and
// applies any necessary request modifications using the utils.FixRequest utility.
//
// The function returns a paginated page of MerchantModel and an error if the operation fails.
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

// GetMerchants retrieves a paginated list of merchants from the database.
//
// It takes an http.Request and a search query string as input. The method uses
// GORM to query the database for merchants, applying the search query to
// the merchant name and description fields. If the request contains a company ID
// header, the method also filters the result by the company ID. If the request
// contains a status query parameter, the method also filters the result by the
// given status. The function utilizes pagination to manage the result set and
// applies any necessary request modifications using the utils.FixRequest utility.
//
// The function returns a paginated page of MerchantModel and an error if the
// operation fails. The MerchantModel is populated with its picture if it exists,
// as well as its company.
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

// CreateProduct creates a new product and associates it with the given merchant and company.
//
// The function takes a pointer to a ProductModel, which contains the data for the new product,
// the ID of the merchant, and the ID of the company. The product is created with a status of
// "PENDING". The function also creates a new ProductMerchant record associating the product with
// the merchant, using the original price and adjustment price from the product. The function
// returns an error if the creation of the product or the product merchant record fails.
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

// GetMerchantProductDetail retrieves a product by its ID for a specific merchant.
//
// The function takes the ID of the product, the ID of the merchant, and an optional warehouse ID.
// The function returns a pointer to a ProductModel if the product is found, otherwise an error is
// returned if the retrieval fails.
//
// The function also populates the product with its price and discount information.
// If the warehouse ID is not provided, the total stock of the product is calculated without considering
// the warehouse ID.
// The total stock is calculated by calling the GetStock function, which returns the total stock of the
// product with the given merchant ID and warehouse ID.
// The function also updates the product with the discounted price if there is an active discount.
// The discounted price is calculated by calling the CalculateDiscountedPrice function, which returns the
// discounted price of the product with the given merchant ID and price.
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

// GetMerchantProducts retrieves a list of products associated with a specific merchant.
//
// The function takes the ID of the merchant, a search query string, and an optional warehouse ID.
// The function also takes a boolean indicating whether to use the brand or not.
// The function returns a paginated page of ProductModel and an error if the operation fails.
//
// If the search query is not empty, the function applies the search query to various product fields.
// If the warehouse ID is not provided, the total stock of the product is calculated without considering the warehouse ID.
// The total stock is calculated by calling the GetStock function, which returns the total stock of the product with the given merchant ID and warehouse ID.
// The function also populates each product with its discounted price if there is an active discount.
// The discounted price is calculated by calling the CalculateDiscountedPrice function, which returns the discounted price of the product with the given merchant ID and price.
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

// CountMerchantByStatus returns the count of merchants with a specific status.
//
// The function takes a status as a string and queries the database to count the
// number of merchants that have this status. It returns the count of such merchants
// and an error if any occurs during the database operation.

func (s *MerchantService) CountMerchantByStatus(status string) (int64, error) {

	var count int64
	if err := s.db.Model(&models.MerchantModel{}).Where("status = ?", status).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

// AddProductsToMerchant adds a list of products to a merchant.
//
// The function takes the ID of the merchant and a slice of product IDs as input.
// It returns an error if the addition fails. The function uses a transaction to
// ensure that if any error occurs during the addition process, the transaction
// is rolled back and the operation is failed.
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

// DeleteProductsFromMerchant deletes a list of products from a merchant.
//
// The function takes the ID of the merchant and a slice of product IDs as input.
// It returns an error if the deletion fails. The function uses a transaction to
// ensure that if any error occurs during the deletion process, the transaction
// is rolled back and the operation is failed.
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

// EditProductPrice edits the price of a product in a merchant.
//
// The function takes the merchant ID, product ID, and the new price as input.
// It returns an error if the update fails. The function uses a transaction to
// ensure that if any error occurs during the update process, the transaction
// is rolled back and the operation is failed.
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

// EditVariantPrice edits the price of a variant in a merchant.
//
// The function takes the merchant ID, variant ID, and the new price as input.
// It returns an error if the update fails. The function uses a transaction to
// ensure that if any error occurs during the update process, the transaction
// is rolled back and the operation is failed.
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

// GetProductAvailableByMerchant retrieves the availability of products in an order request for a specific merchant.
//
// The function takes a merchant and an order request as input. It iterates over the order request items and
// checks the availability of each item against the merchant stock. If the item is out of stock, the item status
// is marked as "OUT_OF_STOCK". Otherwise, the item status is marked as "AVAILABLE". The function also calculates
// the total discount amount for the merchant.
//
// The function returns a MerchantAvailableProduct object that contains the merchant ID, name, items, sub total,
// sub total before discount, order request ID, and total discount amount.
//
// If any error occurs during the operation, the function returns an error.
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

// GetPhoneNumberFromMerchantID retrieves a list of phone numbers associated with a specific merchant.
//
// It takes a merchant ID as a string and returns a slice of phone numbers and an error if the operation fails.
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

// GetPushTokenFromMerchantID retrieves a list of push tokens associated with a specific merchant.
//
// It takes a merchant ID as a string and returns a slice of push tokens and an error if the operation fails.
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

// CreateMerchantType creates a new merchant type.
//
// It takes a merchant type model as a parameter and returns an error if the operation fails.
func (s *MerchantService) CreateMerchantType(data *models.MerchantTypeModel) error {

	err := s.db.Create(&data).Error
	if err != nil {
		return err
	}

	return nil
}

// GetAllMerchantType retrieves a list of all merchant types.
//
// It takes a request and a search string as parameters and returns a paginate.Page and an error if the operation fails.
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

// GetMerchantType retrieves a merchant type by its ID.
//
// It takes a merchant type ID as a parameter and returns a pointer to a merchant type model and an error if the operation fails.
func (s *MerchantService) GetMerchantType(id string) (*models.MerchantTypeModel, error) {
	var merchantType models.MerchantTypeModel
	err := s.db.Where("id = ?", id).First(&merchantType).Error
	if err != nil {
		return nil, err
	}

	return &merchantType, nil
}

// UpdateMerchantType updates a merchant type.
//
// It takes a merchant type model as a parameter and returns an error if the operation fails.
func (s *MerchantService) UpdateMerchantType(data *models.MerchantTypeModel) error {

	err := s.db.Save(data).Error
	if err != nil {
		return err
	}

	return nil
}

// DeleteMerchantType deletes a merchant type.
//
// It takes a merchant type ID as a parameter and returns an error if the operation fails.
func (s *MerchantService) DeleteMerchantType(id string) error {
	err := s.db.Delete(&models.MerchantTypeModel{}, "id = ?", id).Error
	if err != nil {
		return err
	}

	return nil
}

// UpdateProduct updates a product for a merchant.
//
// It takes a product ID, merchant ID, company ID, and a product model as parameters and returns an error if the operation fails.
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

// GetProductByID retrieves a product by its ID for a specific merchant and company.
//
// It takes a product ID and an http.Request as parameters and returns a pointer to a ProductModel if the product is found,
// otherwise an error is returned if the retrieval fails.
//
// The function also populates the product with its price and discount information.
// If the warehouse ID is not provided, the total stock of the product is calculated without considering the warehouse ID.
// The total stock is calculated by calling the GetStock function, which returns the total stock of the product with the given
// merchant ID and warehouse ID.
// The function also updates the product with the discounted price if there is an active discount.
// The discounted price is calculated by calling the CalculateDiscountedPrice function, which returns the discounted price of the
// product with the given merchant ID and price.
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

// GetSalesCountByBrand retrieves the sales count of all products by their brand and filters by various criteria.
//
// Args:
//
//	request: an optional HTTP request containing additional query parameters.
//	merchantID: an optional ID of the merchant to filter the sales count by.
//	warehouseID: an optional ID of the warehouse to filter the sales count by.
//	companyID: an optional ID of the company to filter the sales count by.
//	distributorID: an optional ID of the distributor to filter the sales count by.
//	startDate: an optional start date to filter the sales count by.
//	endDate: an optional end date to filter the sales count by.
//
// Returns:
//
//	a slice of maps, where each map contains the brand ID, name, and total sales quantity of the brand.
//	The total sales quantity is negated as the sales are stored as negative stock movements.
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

// GetSalesCountByBrandAndOrderType retrieves the total sales count grouped by brand and order type.
//
// This function takes an HTTP request and optional identifiers for merchant, warehouse,
// company, and distributor to filter the stock movements. It queries the stock_movements
// table to calculate the aggregated sales count for each brand and order type combination.
// The function joins the stock movements with products, brands, and pos_sales to gather
// necessary information and groups the results by brand ID, brand name, and order type.
// It returns a map where keys are order types, and values are lists of sales count data.
//
// Args:
//
//	request: An HTTP request that may contain additional filter parameters.
//	merchantID: An optional pointer to the merchant ID to filter stock movements.
//	warehouseID: An optional pointer to the warehouse ID to filter stock movements.
//	companyID: An optional pointer to the company ID to filter stock movements.
//	distributorID: An optional pointer to the distributor ID to filter stock movements.
//
// Returns:
//
//	A map with order types as keys and lists of sales count data as values, or an error if
//	any occurs during the database query.
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

// GetMerchantUsers retrieves a paginated list of users associated with a specific merchant.
//
// It takes an HTTP request, search query string, merchant ID, and company ID as input. The method uses
// GORM to query the database for users linked to the specified merchant ID, applying the search query to
// the user's full name, email, and phone. The function utilizes pagination to manage the result set and
// applies any necessary request modifications using the utils.FixRequest utility.
//
// The function returns a paginated page of UserModel and an error if the operation fails.
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

// AddMerchantUser adds a list of users to a merchant.
//
// The function takes the ID of the merchant and a slice of user IDs as input.
// It returns an error if the addition fails. The function uses a transaction to
// ensure that if any error occurs during the addition process, the transaction
// is rolled back and the operation is failed.
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

// DeleteUserFromMerchant deletes a list of users from a merchant.
//
// The function takes the ID of the merchant and a slice of user IDs as input.
// It returns an error if the deletion fails. The function uses a transaction to
// ensure that if any error occurs during the deletion process, the transaction
// is rolled back and the operation is failed.
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

// ListPricesOfProduct lists all prices of a product.
//
// The function takes the ID of a product as input. It returns a list of prices
// of the product and an error if the query fails.
func (s *MerchantService) ListPricesOfProduct(productID string) ([]models.PriceModel, error) {
	var prices []models.PriceModel
	err := s.db.Preload("PriceCategory").Where("product_id = ?", productID).Find(&prices).Error
	return prices, err
}

// GetDesksFromID lists all desks of a merchant.
//
// The function takes the ID of a merchant and a request object as input. It
// returns a paginate.Page object and an error if the query fails. The
// paginate.Page object contains the list of desks and the pagination
// information.
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

// AddDeskToMerchant adds a desk to a merchant.
//
// The function takes the ID of the merchant and a desk object as input. It
// returns an error if the addition fails. The function uses a transaction to
// ensure that if any error occurs during the addition process, the transaction
// is rolled back and the operation is failed.
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

// UpdateMerchantDesk updates a desk of a merchant.
//
// The function takes the ID of the merchant, the ID of the desk and a desk
// object as input. It returns an error if the update fails. The function uses a
// transaction to ensure that if any error occurs during the update process, the
// transaction is rolled back and the operation is failed.
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

// DeleteDeskFromMerchant deletes a desk from a merchant.
//
// The function takes the ID of the merchant and the ID of the desk as input. It
// returns an error if the deletion fails. The function uses a transaction to
// ensure that if any error occurs during the deletion process, the transaction
// is rolled back and the operation is failed.
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

// GetLayoutDetailFromID retrieves a layout of a merchant by its ID.
//
// The function takes the ID of the merchant and the ID of the layout as input. It
// returns a MerchantDeskLayout model and an error if the query fails. The
// function also populates the layout with its desks and contacts.
func (s *MerchantService) GetLayoutDetailFromID(merchantID string, layoutID string) (*models.MerchantDeskLayout, error) {
	var layout models.MerchantDeskLayout
	if err := s.db.Preload("MerchantDesks.Contact").Where("merchant_id = ? AND id = ?", merchantID, layoutID).First(&layout).Error; err != nil {
		return nil, err
	}
	return &layout, nil
}

// GetLayoutsFromID lists all layouts of a merchant.
//
// The function takes the ID of the merchant and a request object as input. It
// returns a paginate.Page object and an error if the query fails. The
// paginate.Page object contains the list of layouts and the pagination
// information.
func (s *MerchantService) GetLayoutsFromID(request http.Request, merchantID string) (paginate.Page, error) {
	pg := paginate.New()
	var layouts []models.MerchantDeskLayout
	stmt := s.db.Preload("MerchantDesks.Contact").Where("merchant_id = ?", merchantID)
	if request.URL.Query().Get("search") != "" {
		stmt = stmt.Where("layout_name LIKE ?", "%"+request.URL.Query().Get("search")+"%")
	}
	stmt = stmt.Order("created_at").Model(&models.MerchantDeskLayout{})
	utils.FixRequest(&request)
	page := pg.With(stmt).Request(request).Response(&layouts)
	page.Page = page.Page + 1
	return page, nil
}

// AddLayoutToMerchant adds a new layout to a merchant.
//
// The function takes the ID of the merchant and a layout object as input. It
// returns an error if the creation fails. The function uses a transaction to
// ensure that if any error occurs during the creation process, the transaction
// is rolled back and the operation is failed.
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

// UpdateLayoutMerchant updates a layout of a merchant.
//
// The function takes the ID of the merchant, the ID of the layout and a layout
// object as input. It returns an error if the update fails. The function uses a
// transaction to ensure that if any error occurs during the update process, the
// transaction is rolled back and the operation is failed.
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

// DeleteLayoutMerchant deletes a layout of a merchant.
//
// The function takes the ID of the merchant and the ID of the layout as input.
// It returns an error if the deletion fails. The function uses a transaction
// to ensure that if any error occurs during the deletion process, the transaction
// is rolled back and the operation is failed.
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

// UpdateTableContact updates the contact information for a table in a merchant's layout.
//
// The function takes the ID of the merchant, the ID of the table, the contact name,
// phone number, and contact ID as input. It returns an error if the update fails.
// The function uses a transaction to ensure that if any error occurs during the update
// process, the transaction is rolled back and the operation is failed. It also creates
// a new contact if the phone number is provided and the contact does not exist.
func (s *MerchantService) UpdateTableContact(merchantID string, tableID string, contactName string, contactPhone string, contactID string) error {
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	var phoneNumber string
	if contactPhone != "" {
		phoneNumber = utils.ParsePhoneNumber(contactPhone, "ID")
	}

	if err := tx.Model(&models.MerchantDesk{}).Where("merchant_id = ? AND id = ?", merchantID, tableID).
		Updates(map[string]any{
			"contact_name":  contactName,
			"contact_phone": phoneNumber,
			"contact_id":    contactID,
		}).Error; err != nil {
		tx.Rollback()
		return err
	}
	var merchant models.MerchantModel

	if phoneNumber != "" {
		if err := tx.Model(&models.MerchantModel{}).Where("id = ?", merchantID).First(&merchant).Error; err != nil {
			tx.Rollback()
			return err
		}
		var contact models.ContactModel
		err := tx.Model(&models.ContactModel{}).Where("phone = ? AND company_id = ?", phoneNumber, merchant.CompanyID).First(&contact).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			contact.Name = contactName
			contact.Phone = &phoneNumber
			contact.CompanyID = merchant.CompanyID
			contact.IsCustomer = true
			if err := tx.Model(&models.ContactModel{}).Create(&contact).Error; err != nil {
				tx.Rollback()
				return err
			}
		}
	}

	if contactID != "" {
		if err := tx.Model(&models.MerchantDesk{}).Where("merchant_id = ? AND id = ?", merchantID, tableID).
			Updates(map[string]any{
				"contact_id": contactID,
			}).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit().Error
}

// UpdateTableStatus updates the status of a table in a merchant's layout.
//
// The function takes the ID of the merchant, the ID of the table, and the new status as input.
// It returns an error if the update fails. If the table is already occupied, an error is returned.
// The function uses a transaction to ensure that if any error occurs during the update process,
// the transaction is rolled back and the operation is failed.
func (s *MerchantService) UpdateTableStatus(merchantID string, tableID string, status string) error {
	var table models.MerchantDesk
	if err := s.db.Model(&table).Where("merchant_id = ? AND id = ?", merchantID, tableID).First(&table).Error; err != nil {
		return err
	}
	if strings.ToUpper(*table.Status) == "OCCUPIED" {
		return errors.New("table is already occupied")
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

// GetMerchantStations retrieves a paginated list of merchant stations.
//
// The function takes an HTTP request and the ID of the merchant as input.
// It returns a paginated page of MerchantStation models and an error if the operation fails.
// The search query, if provided, filters the stations by name and description.
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

// GetMerchantStationDetail retrieves the details of a specific merchant station.
//
// The function takes the ID of the merchant and the ID of the station as input.
// It returns the MerchantStation model populated with its associated products or an error if the operation fails.
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

// GetOrdersFromStation retrieves a paginated list of orders from a specific merchant station.
//
// The function takes an HTTP request, the ID of the merchant, the ID of the station, and a status filter as input.
// It returns a paginated page of MerchantStationOrder models and an error if the operation fails.
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

// GetMerchantOrderStation retrieves a specific order from a merchant station.
//
// The function takes the ID of the order station and the ID of the station as input.
// It returns the MerchantStationOrder model or an error if the operation fails.
func (s *MerchantService) GetMerchantOrderStation(orderStationID string, stationID string) (*models.MerchantStationOrder, error) {
	var orderStation models.MerchantStationOrder
	if err := s.db.Where("id = ? AND  merchant_station_id = ?", orderStationID, stationID).First(&orderStation).Error; err != nil {
		return nil, err
	}
	return &orderStation, nil
}

// GetProductsFromMerchantStation retrieves products associated with a specific merchant station.
//
// The function takes the ID of the merchant and the ID of the station as input.
// It returns a slice of ProductModel populated with product images or an error if the operation fails.
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

// CreateMerchantStation creates a new station for a specific merchant.
//
// The function takes the ID of the merchant and the station model as input.
// It returns an error if the creation fails.
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

// UpdateMerchantStation updates the details of a specific merchant station.
//
// The function takes the ID of the merchant, the ID of the station, and the updated station model as input.
// It returns an error if the update fails.
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

// DeleteMerchantStation deletes a specific station from a merchant.
//
// The function takes the ID of the merchant and the ID of the station as input.
// It returns an error if the deletion fails.
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

// AddProductsToMerchantStation associates products with a specific merchant station.
//
// The function takes the ID of the merchant, the ID of the station, and a slice of product IDs as input.
// It returns an error if the operation fails.
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

// DeleteProductFromMerchantStation dissociates products from a specific merchant station.
//
// The function takes the ID of the merchant, the ID of the station, and a slice of product IDs as input.
// It returns an error if the operation fails.
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

// CreateOrder creates a new order for a specific merchant.
//
// The function takes the ID of the merchant and the order model as input.
// It returns an error if the creation fails.
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

// parseItems parses the given JSON into a slice of MerchantOrderItem.
func parseItems(orderItems json.RawMessage) []models.MerchantOrderItem {
	items := []models.MerchantOrderItem{}

	err := json.Unmarshal(orderItems, &items)
	if err != nil {
		return []models.MerchantOrderItem{}
	}
	return items
}

// DistributeOrder distributes the order to the stations.
//
// The function takes the ID of the merchant, the order model as input.
// It returns a slice of MerchantStationOrder and an error if the operation fails.
func (s *MerchantService) DistributeOrder(merchantID string, order *models.MerchantOrder) ([]models.MerchantStationOrder, error) {
	items := []models.MerchantOrderItem{}

	err := json.Unmarshal(order.Items, &items)
	if err != nil {
		return nil, err
	}
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

// GetOrderDetail returns the details of a specific order.
//
// The function takes the ID of the merchant, the ID of the order, and returns the order model and an error if the operation fails.
func (s *MerchantService) GetOrderDetail(merchantID string, orderID string) (*models.MerchantOrder, error) {
	var order models.MerchantOrder
	if err := s.db.Preload("MerchantDesk").
		Preload("Cashier").
		Preload("Payments").
		Preload("Contact").
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

	if request.URL.Query().Get("status") != "" {
		stmt = stmt.Where("order_status IN (?)", strings.Split(request.URL.Query().Get("status"), ","))
	}
	if request.URL.Query().Get("start_date") != "" {
		stmt = stmt.Where("created_at >= ?", request.URL.Query().Get("start_date"))
	}
	if request.URL.Query().Get("end_date") != "" {
		stmt = stmt.Where("created_at <= ?", request.URL.Query().Get("end_date"))
	}
	if request.URL.Query().Get("order") != "" {
		stmt = stmt.Order(request.URL.Query().Get("order"))
	} else {
		stmt = stmt.Order("updated_at desc")

	}
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

// GetPrintReceipt generates a PDF receipt for a given order.
//
// The function takes an order model, a template path (optional), and a time format string (optional).
// If the template path is not provided, the default template will be used.
// If the time format string is not provided, the default format will be used (02/01/2006 15:04).
// The function returns the generated PDF as []byte and an error if the operation fails.
func (s *MerchantService) GetPrintReceipt(order *models.MerchantOrder, templatePath, timeFormatStr string) ([]byte, error) {
	if timeFormatStr == "" {
		timeFormatStr = "02/01/2006 15:04"
	}
	var orderItems []models.MerchantOrderItem
	json.Unmarshal(order.Items, &orderItems)
	items := []utils.ReceiptItem{}
	discTotal := 0.0
	customerName := ""
	if order.Contact != nil {
		customerName = order.Contact.Name
	}

	merchant, err := s.GetMerchantByID(*order.MerchantID)
	if err != nil {
		return nil, err
	}
	for _, v := range orderItems {
		disc := ""
		if v.DiscountPercent > 0 {
			disc = fmt.Sprintf("%v%%", v.DiscountPercent)
		}
		discTotal += v.DiscountAmount
		items = append(items, utils.ReceiptItem{
			Description:     v.Product.Name,
			Quantity:        utils.FormatRupiah(v.Quantity),
			Price:           utils.FormatRupiah(v.UnitPrice),
			Total:           utils.FormatRupiah(v.Subtotal),
			DiscountPercent: disc,
			Notes:           v.Notes,
		})

	}
	date := order.UpdatedAt.Format(timeFormatStr)
	var data utils.ReceiptData = utils.ReceiptData{
		Items:           items,
		SubTotalPrice:   utils.FormatRupiah(order.SubTotal),
		TotalPrice:      utils.FormatRupiah(order.Total),
		CashierName:     order.Cashier.FullName,
		Code:            order.Code,
		Date:            date,
		DiscountAmount:  utils.FormatRupiah(discTotal),
		CustomerName:    customerName,
		MerchantName:    merchant.Name,
		MerchantAddress: fmt.Sprintf("%s, %s", merchant.Address, merchant.Phone),
	}

	return utils.GenerateOrderReceipt(data, templatePath)
}

// SplitBill splits a bill into two separate orders.
//
// The function takes an order model, a contact model, and a new items slice.
// The function will create a new order with the new items and the same merchant, cashier, and desk as the original order.
// The function will also update the original order by subtracting the new items from the original order.
// The function returns the new order and an error if the operation fails.
func (s *MerchantService) SplitBill(order *models.MerchantOrder, contact *models.ContactModel, newItems []models.MerchantOrderItem) (*models.MerchantOrder, error) {

	var orderItems []models.MerchantOrderItem
	json.Unmarshal(order.Items, &orderItems)

	var newOrder models.MerchantOrder
	newOrder.ID = utils.Uuid()
	newOrder.MerchantID = order.MerchantID
	newOrder.OrderStatus = order.OrderStatus
	newOrder.CashierID = order.CashierID
	newOrder.Step = order.Step
	newOrder.MerchantDeskID = order.MerchantDeskID
	newOrder.ParentID = &order.ID
	newOrder.Code = strings.ToUpper(utils.GenerateRandomString(6))
	if contact != nil {
		newOrder.ContactID = &contact.ID
		b, _ := json.Marshal(contact)
		newOrder.ContactData = b
	}
	newOrder.SubTotal = 0
	newOrder.Total = 0

	err := s.ctx.DB.Transaction(func(tx *gorm.DB) error {
		newItems2 := []models.MerchantOrderItem{}
		for i, v := range newItems {
			s.countSubtotal(&v)
			newItems[i] = v
			if v.Quantity > 0 {
				newItems2 = append(newItems2, v)
			}
		}
		newItems = newItems2
		b, _ := json.Marshal(newItems)
		newOrder.Items = b

		err := tx.Create(&newOrder).Error
		if err != nil {
			return err
		}
		subtotal := 0.0
		total := 0.0
		for _, newItem := range newItems {
			subtotal += newItem.SubtotalBeforeDisc
			total += newItem.Subtotal
		}
		newOrder.SubTotal = subtotal
		newOrder.Total = total

		fmt.Println("UPDATE NEW ORDER SUBTOTAL", subtotal)
		fmt.Println("UPDATE NEW ORDER TOTAL", total)
		err = tx.Model(&models.MerchantOrder{}).Where("id = ?", newOrder.ID).Debug().Updates(map[string]any{
			"sub_total": subtotal,
			"total":     total,
		}).Error
		if err != nil {
			return err
		}

		for i, oldItem := range orderItems {
			for _, newItem := range newItems {
				if oldItem.ID == newItem.ID {
					oldItem.Quantity = oldItem.Quantity - newItem.Quantity
					s.countSubtotal(&oldItem)
					orderItems[i] = oldItem
				}
			}
		}

		for _, oldOrder := range order.MerchantStationOrders {
			var item models.MerchantOrderItem
			json.Unmarshal(oldOrder.Item, &item)
			for _, newItem := range newItems {
				if item.ID == newItem.ID {
					item.Quantity = item.Quantity - newItem.Quantity
					s.countSubtotal(&item)

					if newItem.Quantity > 0 {
						b, _ := json.Marshal(newItem)
						newStatonOrder := models.MerchantStationOrder{
							OrderID:           newOrder.ID,
							Item:              b,
							Status:            oldOrder.Status,
							MerchantDeskID:    oldOrder.MerchantDeskID,
							MerchantStationID: oldOrder.MerchantStationID,
						}

						newStatonOrder.ID = utils.Uuid()
						fmt.Println("CREATE NEW STATION ORDER", newStatonOrder)
						tx.Create(&newStatonOrder)
					}

				}
			}
			if item.Quantity > 0 {
				b, _ := json.Marshal(item)
				oldOrder.Item = b
				fmt.Println("UPDATE OLD STATION ORDER", oldOrder)
				tx.Save(&oldOrder)
			}

			if item.Quantity == 0 {
				tx.Delete(&oldOrder)
			}

		}

		old_subtotal := 0.0
		old_total := 0.0
		fmt.Println("UPDATE OLD ORDER", order)
		for _, v := range orderItems {
			utils.LogJson(v)
			old_subtotal += v.SubtotalBeforeDisc
			old_total += v.Subtotal
		}
		fmt.Println("UPDATE OLD ORDER SUBTOTAL", old_subtotal)
		fmt.Println("UPDATE OLD ORDER TOTAL", old_total)
		order.SubTotal = old_subtotal
		order.Total = old_total

		orderItems2 := []models.MerchantOrderItem{}
		for _, v := range orderItems {
			if v.Quantity > 0 {
				orderItems2 = append(orderItems2, v)
			}
		}
		orderItems = orderItems2
		c, _ := json.Marshal(orderItems)

		err = tx.Model(&models.MerchantOrder{}).Where("id = ?", order.ID).Debug().Updates(map[string]any{
			"sub_total": old_subtotal,
			"total":     old_total,
			"items":     c,
		}).Error
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &newOrder, nil
}

// countSubtotal recalculates and updates the subtotal before discount, discount amount,
// and subtotal of the given order item.
//
// If the item has a discount percent greater than 0, the function calculates the
// subtotal before discount as the product of the item's quantity and unit price.
// It then calculates the discount amount as the product of the subtotal before
// discount and the discount percent divided by 100. The subtotal is calculated
// as the difference between the subtotal before discount and the discount amount.
//
// If the item has no discount percent, the function sets the subtotal before
// discount as the product of the item's quantity and unit price, and sets the
// subtotal as the difference between the subtotal before discount and the
// discount amount.
func (s *MerchantService) countSubtotal(item *models.MerchantOrderItem) {
	var beforeDisc = item.Quantity * item.UnitPrice
	item.SubtotalBeforeDisc = beforeDisc
	if item.DiscountPercent > 0 {
		item.Subtotal = beforeDisc - (beforeDisc * item.DiscountPercent / 100)
		item.DiscountAmount = beforeDisc * item.DiscountPercent / 100
	} else {
		item.Subtotal = beforeDisc - item.DiscountAmount
	}
}

// GetMerchantTableDetail retrieves a table by its ID in a specific merchant.
//
// The function takes the ID of the merchant and the ID of the table as input.
// It returns a MerchantDesk model populated with its associated contact, layout, and active orders.
//
// The function returns an error if the query fails.
func (s *MerchantService) GetMerchantTableDetail(merchantID string, tableID string) (*models.MerchantDesk, error) {
	table := &models.MerchantDesk{}
	err := s.db.Preload("Contact").Preload("MerchantDeskLayout").Preload("ActiveOrders", "order_status = 'ACTIVE'").Where("merchant_id = ? AND id = ?", merchantID, tableID).First(table).Error
	if err != nil {
		return nil, err
	}
	return table, nil
}
