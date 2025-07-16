package product

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/file"
	"github.com/AMETORY/ametory-erp-modules/shared/models"
	"github.com/AMETORY/ametory-erp-modules/utils"
	"github.com/morkid/paginate"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ProductService struct {
	db          *gorm.DB
	ctx         *context.ERPContext
	fileService *file.FileService
	tagService  *TagService
}

// NewProductService creates a new instance of ProductService with the given database connection, context, file service and tag service.
func NewProductService(db *gorm.DB, ctx *context.ERPContext, fileService *file.FileService, tagService *TagService) *ProductService {
	return &ProductService{db: db, ctx: ctx, fileService: fileService, tagService: tagService}
}

// Migrate runs the database migration to create the necessary tables for the product module.
func Migrate(db *gorm.DB) error {
	// db.Migrator().AlterColumn(&models.VariantModel{}, "product_id")
	return db.AutoMigrate(
		&models.ProductModel{},
		&models.ProductCategoryModel{},
		&models.MasterProductModel{},
		&models.PriceCategoryModel{},
		&models.PriceModel{},
		&models.MasterProductPriceModel{},
		&models.VariantModel{},
		&models.VariantProductAttributeModel{},
		&models.ProductAttributeModel{},
		&models.ProductMerchant{},
		&models.DiscountModel{},
		&models.TagModel{},
		&models.ProductTag{},
		&models.VariantTag{},
		&models.VarianMerchant{},
		&models.ProductFeedbackModel{},
	)
}

// CreateProduct creates a new product.
//
// The function takes a pointer to a ProductModel, which contains the data for the new product.
// The function returns an error if the creation of the product fails.
func (s *ProductService) CreateProduct(data *models.ProductModel) error {
	return s.db.Create(data).Error
}

func (s *ProductService) CreateOrUpdateProduct(data *models.ProductModel) error {
	var product models.ProductModel
	err := s.db.Where("id = ?", data.ID).First(&product).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return s.db.Create(data).Error
		}
		return err
	}
	data.ID = product.ID
	return s.db.Save(data).Error
}

// UpdateProduct updates an existing product.
//
// The function takes a pointer to a ProductModel, which contains the updated data of the product.
// The function returns an error if the update operation fails.
//
// Note that this function omits any associations of the product, so the associations will not be updated.
func (s *ProductService) UpdateProduct(data *models.ProductModel) error {
	return s.db.Omit(clause.Associations).Save(data).Error
}

// DeleteProduct deletes a product by its ID.
//
// The function takes the ID of the product to be deleted as a string.
// It returns an error if the deletion operation fails.
func (s *ProductService) DeleteProduct(id string) error {
	return s.db.Where("id = ?", id).Delete(&models.ProductModel{}).Error
}

// GetVariantByID retrieves a variant by its ID.
//
// The function takes the ID of the variant to be retrieved as a string and an http.Request.
// The request is used to get the merchant ID from the headers.
// The function returns a pointer to a VariantModel if the variant is found,
// otherwise an error is returned if the retrieval fails.
//
// The function also returns the price of the variant with the given merchant ID.
func (s *ProductService) GetVariantByID(variantID string, request *http.Request) (*models.VariantModel, error) {
	var variant models.VariantModel
	err := s.db.Where("id = ?", variantID).First(&variant).Error
	if err != nil {
		return nil, err
	}
	merchantID := request.Header.Get("ID-Merchant")
	if merchantID == "" {
		return nil, errors.New("merchant not found")
	}
	// price := s.GetVariantPrice(merchantID, &variant)
	// variant.Price = price
	variant.GetPriceAndDiscount(s.db)

	return &variant, nil
}

// GetVariantPrice retrieves the price of the variant with the given merchant ID.
//
// The function takes the merchant ID as a string and a pointer to a VariantModel.
// It returns the price of the variant with the given merchant ID.
// If the variant merchant is not found, the function returns the price of the variant.
// The function also updates the variant with the discounted price if there is an active discount.
func (s *ProductService) GetVariantPrice(merchantID string, variant *models.VariantModel) float64 {
	var variantMerchant models.VarianMerchant
	price := variant.Price
	err := s.db.Where("variant_id = ? AND merchant_id = ?", variant.ID, merchantID).First(&variantMerchant).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return price
	}

	price = variantMerchant.Price
	variant.LastUpdatedStock = variantMerchant.LastUpdatedStock
	variant.LastStock = variantMerchant.LastStock
	discountedPrice, _, _, _, err := s.CalculateDiscountedPrice(variant.ProductID, price)
	if err == nil {
		price = discountedPrice
	}
	return price
}

// GetProductByID retrieves a product by its ID.
//
// The function takes the ID of the product to be retrieved as a string and an http.Request.
// The request is used to get the merchant ID from the headers.
// The function returns a pointer to a ProductModel if the product is found,
// otherwise an error is returned if the retrieval fails.
//
// The function also returns the total stock of the product, which is the sum of the stock of all its variants.
// The total stock is calculated by calling the GetStock function, which returns the total stock of the product with the given merchant ID and warehouse ID.
// If the merchant ID is not provided, the total stock of the product is calculated without considering the merchant ID.
// If the warehouse ID is not provided, the total stock of the product is calculated without considering the warehouse ID.
//
// The function also updates the product with the discounted price if there is an active discount.
// The discounted price is calculated by calling the CalculateDiscountedPrice function, which returns the discounted price of the product with the given merchant ID and price.
// If the merchant ID is not provided, the discounted price is calculated without considering the merchant ID.
func (s *ProductService) GetProductByID(id string, request *http.Request) (*models.ProductModel, error) {
	idMerchant := ""
	if request != nil {
		if request.Header != nil {
			idMerchant = request.Header.Get("ID-Merchant")
		}
	} else {
		return nil, errors.New("request is nil")
	}
	var product models.ProductModel
	err := s.db.Preload("Tags").Preload("Company").Preload("Variants").Preload("Tax").Preload("MasterProduct").Preload("Category", func(db *gorm.DB) *gorm.DB {
		return db.Select("id", "name")
	}).Preload("Brand", func(db *gorm.DB) *gorm.DB {
		return db.Select("id", "name")
	}).Where("id = ?", id).First(&product).Error
	product.Prices, _ = s.ListPricesOfProduct(product.ID)
	product.ProductImages, _ = s.ListImagesOfProduct(product.ID)
	var warehouseID *string
	if request != nil {
		warehouseIDStr := request.Header.Get("ID-Warehouse")
		if warehouseIDStr != "" {
			warehouseID = &warehouseIDStr
		}
	}
	stock, _ := s.GetStock(product.ID, request, warehouseID)

	product.TotalStock = stock
	for i, v := range product.Variants {
		if idMerchant != "" {
			v.MerchantID = &idMerchant
		}
		v.GetPriceAndDiscount(s.db)
		variantStock, _ := s.GetVariantStock(product.ID, v.ID, request, warehouseID)
		v.TotalStock = variantStock
		product.Variants[i] = v
		fmt.Println("VARIANT STOCK", v.ID, variantStock)
	}
	if idMerchant != "" {
		product.MerchantID = &idMerchant
	}
	product.GetPriceAndDiscount(s.db)
	return &product, err
}

// GetProductByCode retrieves a product by its code.
//
// Args:
//
//	code: the code of the product to retrieve.
//
// Returns:
//
//	the product if found, and an error if any error occurs.
func (s *ProductService) GetProductByCode(code string) (*models.ProductModel, error) {
	var product models.ProductModel
	err := s.db.Preload("Category", func(db *gorm.DB) *gorm.DB {
		return db.Select("id", "name")
	}).Preload("Brand", func(db *gorm.DB) *gorm.DB {
		return db.Select("id", "name")
	}).Where("sku = ?", code).First(&product).Error
	product.Prices, _ = s.ListPricesOfProduct(product.ID)
	product.ProductImages, _ = s.ListImagesOfProduct(product.ID)
	product.GetPriceAndDiscount(s.db)
	return &product, err
}

// GetProductBySku retrieves a product by its SKU.
//
// Args:
//
//	sku: the SKU of the product to retrieve.
//
// Returns:
//
//	a pointer to a ProductModel if the product is found, and an error if any error occurs.
//
// The function preloads the Category and Brand associations and retrieves
// the product prices and images. It also calculates the product's discounted price, if applicable.
func (s *ProductService) GetProductBySku(sku string) (*models.ProductModel, error) {
	var product models.ProductModel
	err := s.db.Preload("Category", func(db *gorm.DB) *gorm.DB {
		return db.Select("id", "name")
	}).Preload("Brand", func(db *gorm.DB) *gorm.DB {
		return db.Select("id", "name")
	}).Where("sku = ?", sku).First(&product).Error
	if err != nil {
		return nil, err
	}
	product.Prices, _ = s.ListPricesOfProduct(product.ID)
	product.ProductImages, _ = s.ListImagesOfProduct(product.ID)
	product.GetPriceAndDiscount(s.db)
	return &product, nil
}

// GetProducts retrieves a paginated list of products from the database.
//
// It takes an http.Request, a search query string, and a status string as input.
// The method uses GORM to query the database for products, applying the search
// query to various product fields. If the request contains a company ID header,
// the method also filters the result by the company ID. The function utilizes
// pagination to manage the result set and applies any necessary request
// modifications using the utils.FixRequest utility.
//
// The function returns a paginated page of ProductModel and an error if the
// operation fails.
func (s *ProductService) GetProducts(request http.Request, search string, status *string) (paginate.Page, error) {
	pg := paginate.New()
	stmt := s.db.Preload("Tags").Preload("Variants.Attributes.Attribute").Preload("Company").Preload("Category", func(db *gorm.DB) *gorm.DB {
		return db.Select("id", "name")
	}).Preload("Brand", func(db *gorm.DB) *gorm.DB {
		return db.Select("id", "name")
	}).Preload("Tax")
	stmt = stmt.Joins("LEFT JOIN brands ON brands.id = products.brand_id")
	stmt = stmt.Joins("LEFT JOIN product_categories ON product_categories.id = products.category_id")
	stmt = stmt.Joins("LEFT JOIN product_variants ON product_variants.product_id = products.id")
	stmt = stmt.Joins("LEFT JOIN product_tags ON product_tags.product_model_id = products.id")
	stmt = stmt.Joins("LEFT JOIN tags ON product_tags.tag_model_id = tags.id")
	if search != "" {
		stmt = stmt.Where("products.description ILIKE ? OR products.sku ILIKE ? OR products.name ILIKE ? OR products.barcode ILIKE ? OR brands.name ILIKE ? OR product_categories.name ILIKE ? OR product_variants.display_name ILIKE ?",
			"%"+search+"%",
			"%"+search+"%",
			"%"+search+"%",
			"%"+search+"%",
			"%"+search+"%",
			"%"+search+"%",
			"%"+search+"%",
		)
	}
	if request.Header.Get("ID-Company") != "" {
		if request.Header.Get("ID-Company") == "nil" || request.Header.Get("ID-Company") == "null" {
			stmt = stmt.Where("products.company_id is null")
		} else {
			stmt = stmt.Where("products.company_id = ?", request.Header.Get("ID-Company"))

		}
	}
	if request.Header.Get("ID-Distributor") != "" {
		stmt = stmt.Where("products.company_id = ?", request.Header.Get("ID-Distributor"))
	}
	if request.Header.Get("status") != "" {
		stmt = stmt.Where("products.status = ?", request.Header.Get("status"))
	}

	if request.URL.Query().Get("brand_id") != "" {
		stmt = stmt.Where("products.brand_id = ?", request.URL.Query().Get("brand_id"))
	}
	if request.URL.Query().Get("category_id") != "" {
		stmt = stmt.Where("products.category_id = ?", request.URL.Query().Get("category_id"))
	}
	if request.URL.Query().Get("order") != "" {
		stmt = stmt.Order(request.URL.Query().Get("order"))
	}
	if request.URL.Query().Get("status") != "" {
		stmt = stmt.Where("products.status = ?", request.URL.Query().Get("status"))
	}
	if request.URL.Query().Get("category_ids") != "" {
		stmt = stmt.Where("products.category_id IN (?)", strings.Split(request.URL.Query().Get("category_ids"), ","))
	}

	if status != nil {
		stmt = stmt.Where("products.status = ?", status)
	}

	stmt = stmt.Distinct("products.id")
	stmt = stmt.Select("products.*")
	stmt = stmt.Model(&models.ProductModel{})
	utils.FixRequest(&request)
	page := pg.With(stmt).Request(request).Response(&[]models.ProductModel{})
	page.Page = page.Page + 1
	items := page.Items.(*[]models.ProductModel)
	newItems := make([]models.ProductModel, 0)
	var warehouseID *string
	warehouseIDStr := request.Header.Get("ID-Warehouse")
	if warehouseIDStr != "" {
		warehouseID = &warehouseIDStr
	}

	for _, item := range *items {
		item.GetPriceAndDiscount(s.db)
		activeDiscount, _ := s.GetFirstActiveDiscount(item.ID)
		if activeDiscount.ID != "" {
			item.ActiveDiscount = activeDiscount
		}
		img, err := s.ListImagesOfProduct(item.ID)
		if err == nil {
			item.ProductImages = img
		}
		prices, err := s.ListPricesOfProduct(item.ID)
		if err == nil {
			item.Prices = prices
		}

		salesCount, _ := s.GetSalesCount(item.ID, &request, warehouseID)
		item.SalesCount = salesCount
		stock, _ := s.GetStock(item.ID, &request, warehouseID)
		item.TotalStock = stock

		item.TotalStock = stock
		for i, variant := range item.Variants {
			variant.GetPriceAndDiscount(s.db)
			variantStock, _ := s.GetVariantStock(item.ID, variant.ID, &request, warehouseID)
			variant.TotalStock = variantStock
			salesCount, _ := s.GetSalesVariantCount(item.ID, variant.ID, &request, warehouseID)
			variant.SalesCount = salesCount
			// variant.Price = s.GetVariantPrice(merchantID, &variant)
			item.Variants[i] = variant
			fmt.Println("VARIANT STOCK", variant.ID, variant.TotalStock)
		}
		newItems = append(newItems, item)
	}
	page.Items = &newItems
	return page, nil
}

// GetProductFeedbacks retrieves a paginated list of product feedbacks from the database.
//
// The function takes a product ID, an optional variant ID, a request, and a list of status strings as
// input. The method uses GORM to query the database for product feedbacks, applying the search query
// to the product ID and variant ID fields, and the status strings to the status field. The function
// utilizes pagination to manage the result set and applies any necessary request modifications using
// the utils.FixRequest utility.
//
// The function returns a paginated page of ProductFeedbackModel and an error if the operation fails.
func (s *ProductService) GetProductFeedbacks(productID string, variantID *string, request *http.Request, status []string) (paginate.Page, error) {
	pg := paginate.New()
	stmt := s.db
	if variantID == nil {
		stmt = stmt.Where("product_id = ?", productID)
	} else {
		stmt = stmt.Where("product_id = ? AND variant_id = ?", productID, *variantID)
	}
	if len(status) > 0 {
		stmt = stmt.Where("status IN (?)", status)
	}
	stmt = stmt.Model(&models.ProductFeedbackModel{})
	utils.FixRequest(request)
	page := pg.With(stmt).Request(request).Response(&[]models.ProductFeedbackModel{})
	page.Page = page.Page + 1
	return page, nil
}

// CountProductByMerchantID counts the number of products a merchant has.
//
// The function takes a merchant ID as a string and returns the count of products
// associated with the merchant and an error if the operation fails.
func (s *ProductService) CountProductByMerchantID(merchantID string) (int64, error) {
	var count int64
	err := s.db.Model(&models.ProductMerchant{}).Where("merchant_model_id = ?", merchantID).Count(&count).Error
	return count, err
}

// CreatePriceCategory creates a new price category in the database.
//
// It takes a pointer to a PriceCategoryModel as parameter and returns an error.
// The error will be nil if the price category was created successfully.
func (s *ProductService) CreatePriceCategory(data *models.PriceCategoryModel) error {
	return s.db.Create(data).Error
}

// AddPriceToProduct adds a new price to a product in the database.
//
// It takes a product ID and a pointer to a PriceModel as parameters and returns an error.
// The error will be nil if the price was added successfully.
// The price category ID is required in the price data.
func (s *ProductService) AddPriceToProduct(productID string, data *models.PriceModel) error {
	if data.PriceCategoryID == "" {
		return errors.New("price category id is required")
	}
	data.ProductID = productID
	return s.db.Create(data).Error
}

// ListPricesOfProduct retrieves all the prices of a product.
//
// Args:
//
//	productID: the id of the product whose prices to retrieve.
//
// Returns:
//
//	all the prices of the product if found, and an error if any error occurs.
func (s *ProductService) ListPricesOfProduct(productID string) ([]models.PriceModel, error) {
	var prices []models.PriceModel
	err := s.db.Preload("PriceCategory").Where("product_id = ?", productID).Find(&prices).Error
	return prices, err
}

// ListImagesOfProduct retrieves all the images of a product.
//
// Args:
//
//	productID: the id of the product whose images to retrieve.
//
// Returns:
//
//	all the images of the product if found, and an error if any error occurs.
func (s *ProductService) ListImagesOfProduct(productID string) ([]models.FileModel, error) {
	var images []models.FileModel
	err := s.db.Where("ref_id = ? and ref_type = ?", productID, "product").Find(&images).Error
	return images, err
}

func (s *ProductService) DeletePriceOfProduct(productID string, priceID string) error {
	return s.db.Where("product_id = ? and id = ?", productID, priceID).Delete(&models.PriceModel{}).Error
}

// DeleteImageOfProduct deletes a specific image from a product in the database.
//
// Args:
//
//	productID: the ID of the product whose image to delete.
//	imageID: the ID of the image to delete.
//
// Returns:
//
//	an error if the deletion fails, otherwise returns nil.
func (s *ProductService) DeleteImageOfProduct(productID string, imageID string) error {
	return s.db.Where("ref_id = ? and ref_type = ? and id = ?", productID, "product", imageID).Delete(&models.FileModel{}).Error
}

// GetStock retrieves the total stock of a product in the database.
//
// Args:
//
//	productID: the ID of the product whose stock to retrieve.
//	request: an optional HTTP request containing additional query parameters.
//	warehouseID: an optional ID of the warehouse to filter the stock by.
//
// Returns:
//
//	the total stock quantity of the product if found, and an error if any error occurs.
func (s *ProductService) GetStock(productID string, request *http.Request, warehouseID *string) (float64, error) {

	var totalStock float64
	db := s.db.Model(&models.StockMovementModel{})
	if request != nil {
		if request.Header.Get("ID-Company") != "" {
			db = db.Where("company_id = ?", request.Header.Get("ID-Company"))
		}
		if request.Header.Get("ID-Distributor") != "" {
			db = db.Where("company_id = ?", request.Header.Get("ID-Distributor"))
		}
	}

	if warehouseID != nil {
		db = db.Where("warehouse_id = ?", *warehouseID)
	}

	if err := db.
		Where("product_id = ?", productID).
		Select("COALESCE(SUM(quantity * value), 0)").
		Scan(&totalStock).Error; err != nil {
		return 0, err
	}

	return totalStock, nil
}

// GetSalesCount retrieves the total sales of a product in the database.
//
// Args:
//
//	productID: the ID of the product whose sales to retrieve.
//	request: an optional HTTP request containing additional query parameters.
//	warehouseID: an optional ID of the warehouse to filter the stock by.
//
// Returns:
//
//	the total sales quantity of the product if found, and an error if any error occurs.
//	The result is negated as the sales are stored as negative stock movements.
func (s *ProductService) GetSalesCount(productID string, request *http.Request, warehouseID *string) (float64, error) {

	var totalStock float64
	db := s.db.Model(&models.StockMovementModel{})
	if request != nil {
		if request.Header.Get("ID-Company") != "" {
			if request.Header.Get("ID-Company") == "nil" || request.Header.Get("ID-Company") == "null" {
				db = db.Where("company_id is null")
			} else {
				db = db.Where("company_id = ?", request.Header.Get("ID-Company"))

			}
		}
		if request.Header.Get("ID-Distributor") != "" {
			db = db.Where("company_id = ?", request.Header.Get("ID-Distributor"))
		}
	}

	if warehouseID != nil {
		db = db.Where("warehouse_id = ?", *warehouseID)
	}
	db = db.Where("type in (?)", []models.MovementType{models.MovementTypeReturn, models.MovementTypeSale})

	if err := db.
		Where("product_id = ?", productID).
		Select("COALESCE(SUM(quantity * value), 0)").
		Scan(&totalStock).Error; err != nil {
		return 0, err
	}

	return -totalStock, nil
}

// GetSalesVariantCount retrieves the total sales of a product variant in the database.
//
// Args:
//
//	productID: the ID of the product whose sales to retrieve.
//	variantID: the ID of the variant whose sales to retrieve.
//	request: an optional HTTP request containing additional query parameters.
//	warehouseID: an optional ID of the warehouse to filter the stock by.
//
// Returns:
//
//	the total sales quantity of the product variant if found, and an error if any error occurs.
//	The result is negated as the sales are stored as negative stock movements.
func (s *ProductService) GetSalesVariantCount(productID, variantID string, request *http.Request, warehouseID *string) (float64, error) {

	var totalStock float64
	db := s.db.Model(&models.StockMovementModel{})
	if request != nil {
		if request.Header.Get("ID-Company") != "" {
			if request.Header.Get("ID-Company") == "nil" || request.Header.Get("ID-Company") == "null" {
				db = db.Where("company_id is null")
			} else {
				db = db.Where("company_id = ?", request.Header.Get("ID-Company"))

			}
		}
		if request.Header.Get("ID-Distributor") != "" {
			db = db.Where("company_id = ?", request.Header.Get("ID-Distributor"))
		}
	}

	if warehouseID != nil {
		db = db.Where("warehouse_id = ?", *warehouseID)
	}
	db = db.Where("type in (?)", []models.MovementType{models.MovementTypeReturn, models.MovementTypeSale})

	if err := db.
		Where("product_id = ? and variant_id = ?", productID, variantID).
		Select("COALESCE(SUM(quantity * value), 0)").
		Scan(&totalStock).Error; err != nil {
		return 0, err
	}

	return -totalStock, nil
}

// GetVariantStock retrieves the total stock of a product variant in the database.
//
// Args:
//
//	productID: the ID of the product whose stock to retrieve.
//	variantID: the ID of the variant whose stock to retrieve.
//	request: an optional HTTP request containing additional query parameters.
//	warehouseID: an optional ID of the warehouse to filter the stock by.
//
// Returns:
//
//	the total stock quantity of the product variant if found, and an error if any error occurs.
func (s *ProductService) GetVariantStock(productID string, variantID string, request *http.Request, warehouseID *string) (float64, error) {

	var totalStock float64
	db := s.db.Model(&models.StockMovementModel{})
	if request != nil {
		if request.Header.Get("ID-Company") != "" {
			if request.Header.Get("ID-Company") == "nil" || request.Header.Get("ID-Company") == "null" {
				db = db.Where("company_id is null")
			} else {
				db = db.Where("company_id = ?", request.Header.Get("ID-Company"))

			}
		}
		if request.Header.Get("ID-Distributor") != "" {
			db = db.Where("company_id = ?", request.Header.Get("ID-Distributor"))
		}
	}

	if warehouseID != nil {
		db = db.Where("warehouse_id = ?", *warehouseID)
	}

	if err := db.
		Where("product_id = ? AND variant_id = ?", productID, variantID).
		Select("COALESCE(SUM(quantity * value), 0)").
		Scan(&totalStock).Error; err != nil {
		return 0, err
	}

	return totalStock, nil
}

// GetProductsByMerchant retrieves a list of products associated with a specific merchant.
//
// It takes a merchant ID and a slice of product IDs as input. If product IDs are provided,
// the function filters the products by these IDs in addition to the merchant ID. It returns
// a slice of ProductModel and an error if the operation fails. The function also populates
// each product with its price and discount information.
func (s *ProductService) GetProductsByMerchant(merchantID string, productIDs []string) ([]models.ProductModel, error) {
	var products []models.ProductModel
	db := s.db.Where("merchant_id = ?", merchantID)
	if len(productIDs) > 0 {
		db = db.Where("id in (?)", productIDs)
	}
	err := db.Find(&products).Error
	for i, v := range products {
		v.GetPriceAndDiscount(s.db)
		products[i] = v
	}
	return products, err
}

// CreateProductVariant creates a new product variant in the database.
//
// The function takes a pointer to a VariantModel, which contains the data
// for the new product variant. It returns an error if the creation
// of the product variant fails.
func (s *ProductService) CreateProductVariant(data *models.VariantModel) error {
	return s.db.Create(data).Error
}

// AddProductUnit adds a new product unit to the database.
//
// The function takes a pointer to a ProductUnitData, which contains the data
// for the new product unit. It returns an error if the creation
// of the product unit fails.
func (s *ProductService) AddProductUnit(data *models.ProductUnitData) error {
	return s.db.Create(data).Error
}

// DeleteProductUnit deletes a specific unit from a product in the database.
//
// Args:
//
//	productID: the ID of the product whose unit is to be deleted.
//	unitID: the ID of the unit to be deleted.
//
// Returns:
//
//	an error if the deletion fails, otherwise returns nil.
func (s *ProductService) DeleteProductUnit(productID, unitID string) error {
	return s.db.Where("product_model_id = ? and unit_model_id = ?", productID, unitID).Unscoped().Delete(&models.ProductUnitData{}).Error
}

// GetProductVariants retrieves a list of product variants for a given product ID.
//
// It takes a string product ID and an http.Request as input. The method uses
// GORM to query the database for the product variants. The result is preloaded with
// the variant attributes and tags. The function also retrieves the total stock for
// each variant and adds it to the variant model. The total stock is retrieved by
// calling the GetVariantStock method.
//
// The function returns a slice of VariantModel and an error if the operation
// fails. If the request contains a warehouse ID header, the method also filters
// the result by the warehouse ID.
func (s *ProductService) GetProductVariants(productID string, request http.Request) ([]models.VariantModel, error) {
	var variants []models.VariantModel
	err := s.db.Preload("Attributes.Attribute").Preload("Tags").Where("product_id = ?", productID).Find(&variants).Error
	var warehouseID *string
	warehouseIDStr := request.Header.Get("ID-Warehouse")
	if warehouseIDStr != "" {
		warehouseID = &warehouseIDStr
	}
	for i, v := range variants {
		v.GetPriceAndDiscount(s.db)
		variantStock, _ := s.GetVariantStock(productID, v.ID, &request, warehouseID)
		v.TotalStock = variantStock
		variants[i] = v
		fmt.Println("VARIANT STOCK", v.ID, variantStock)
	}
	return variants, err
}

// UpdateProductVariant updates an existing product variant in the database.
//
// It takes a pointer to a VariantModel, which contains the updated data for the product variant.
// The function returns an error if the update operation fails.
func (s *ProductService) UpdateProductVariant(data *models.VariantModel) error {
	return s.db.Save(data).Error
}

// DeleteProductVariant deletes a product variant from the database by its ID.
//
// It takes a string id as a parameter and attempts to delete the corresponding
// VariantModel record. The function returns an error if the deletion fails,
// otherwise it returns nil indicating the operation was successful.
func (s *ProductService) DeleteProductVariant(id string) error {
	return s.db.Where("id = ?", id).Unscoped().Delete(&models.VariantModel{}).Error
}

// AddDiscount creates a new discount for a given product.
//
// Args:
//
//	productID: the ID of the product for which the discount is to be created.
//	discountType: the type of the discount, which can be either DiscountPercentage or DiscountAmount.
//	value: the value of the discount, which can be either a percentage or a nominal value.
//	startDate: the start date of the discount, which must be before the end date if the end date is not nil.
//	endDate: the end date of the discount, which can be nil if the discount has no end date.
//
// Returns:
//
//	a pointer to the newly created DiscountModel, or an error if the creation of the discount fails.
func (s *ProductService) AddDiscount(productID string, discountType models.DiscountType, value float64, startDate time.Time, endDate *time.Time) (*models.DiscountModel, error) {
	// Validasi tanggal jika endDate tidak nil
	if endDate != nil && startDate.After(*endDate) {
		return nil, errors.New("start date must be before end date")
	}

	// Buat diskon
	discount := models.DiscountModel{
		ProductID: productID,
		Type:      discountType,
		Value:     value,
		StartDate: startDate,
		EndDate:   endDate,
		IsActive:  true,
	}
	if err := s.db.Create(&discount).Error; err != nil {
		return nil, err
	}

	return &discount, nil
}

// DeleteDiscount removes a discount from a product.
//
// This function deletes a discount associated with a specific product using
// the provided product ID and discount ID. It permanently deletes the record
// without soft deleting it. If the operation is successful, it returns nil;
// otherwise, it returns an error detailing what went wrong.
//
// Parameters:
//
//	productID: the ID of the product to which the discount belongs.
//	discountID: the ID of the discount to be deleted.
//
// Returns:
//
//	An error if the deletion fails, otherwise returns nil.
func (s *ProductService) DeleteDiscount(productID, discountID string) error {
	return s.db.Where("id = ? and product_id =?", discountID, productID).Unscoped().Delete(&models.DiscountModel{}).Error
}

// GetFirstActiveDiscount retrieves the first active discount for a given product.
//
// It takes a string product ID as a parameter and returns a pointer to a DiscountModel
// and an error. The function returns the first discount (in descending order of creation date)
// which is active and has a start date before the current time, and an end date either
// after the current time or nil. If the operation fails, an error is returned.
func (s *ProductService) GetFirstActiveDiscount(productID string) (*models.DiscountModel, error) {
	var discount *models.DiscountModel
	err := s.db.Where("product_id = ? AND is_active = ? AND start_date <= ?", productID, true, time.Now()).
		Where("end_date IS NULL OR end_date >= ?", time.Now()).Order("created_at DESC").
		First(&discount).Error
	return discount, err
}

// GetActiveDiscounts retrieves all active discounts for a given product.
//
// It takes a product ID as a parameter and returns a slice of DiscountModel
// and an error. The function returns all discounts that are active and have
// a start date before the current time, and an end date either after the
// current time or nil. Discounts are ordered by their creation date in
// descending order. If the operation fails, an error is returned.
func (s *ProductService) GetActiveDiscounts(productID string) ([]models.DiscountModel, error) {
	var discounts []models.DiscountModel
	err := s.db.Where("product_id = ? AND is_active = ? AND start_date <= ?", productID, true, time.Now()).
		Where("end_date IS NULL OR end_date >= ?", time.Now()).Order("created_at DESC").
		Find(&discounts).Error
	return discounts, err
}

// GetAllDiscountByProductID retrieves all discounts for a given product.
//
// It takes a product ID as a parameter and returns a slice of DiscountModel
// and an error. The function returns all discounts associated with the
// given product, ordered by their creation date in descending order.
// If the operation fails, an error is returned.
func (s *ProductService) GetAllDiscountByProductID(productID string) ([]models.DiscountModel, error) {
	var discounts []models.DiscountModel
	err := s.db.Where("product_id = ?", productID).Find(&discounts).Error
	return discounts, err
}

// GetBestDealByPercentage retrieves a list of products with the best discounts
// in terms of percentage.
//
// Args:
//
//	limit: the number of products to return.
//
// Returns:
//
//	a slice of ProductModel, or an error if the operation fails.
func (s *ProductService) GetBestDealByPercentage(limit int) ([]models.ProductModel, error) {
	var products []models.ProductModel
	err := s.db.Joins("JOIN discounts ON products.id = discounts.product_id").
		Where("discounts.type = ? AND discounts.is_active = ? AND discounts.start_date <= ?", models.DiscountPercentage, true, time.Now()).
		Where("discounts.end_date IS NULL OR discounts.end_date >= ?", time.Now()).
		Order("discounts.value DESC").
		Limit(limit).
		Find(&products).Error
	return products, err
}

// GetBestDealByAmount retrieves a list of products with the best discounts
// in terms of amount.
//
// Args:
//
//	limit: the number of products to return.
//
// Returns:
//
//	a slice of ProductModel, or an error if the operation fails.
func (s *ProductService) GetBestDealByAmount(limit int) ([]models.ProductModel, error) {
	var products []models.ProductModel
	err := s.db.Joins("JOIN discounts ON products.id = discounts.product_id").
		Where("discounts.type = ? AND discounts.is_active = ? AND discounts.start_date <= ?", models.DiscountAmount, true, time.Now()).
		Where("discounts.end_date IS NULL OR discounts.end_date >= ?", time.Now()).
		Order("discounts.value DESC").
		Limit(limit).
		Find(&products).Error
	return products, err
}

// GetBestDealByDiscountedPrice retrieves a list of products with the best
// discounts based on the actual discounted price.
//
// It takes an integer limit as a parameter, which specifies the maximum number
// of products to return. The method uses GORM to perform a join on the products
// and discounts tables, filtering for active discounts within the valid date
// range. It calculates the discounted price using either a percentage or
// amount discount and orders the products by the lowest discounted price.
//
// Returns:
//
// A slice of ProductModel containing the products with the best discounted
// prices, or an error if the operation fails.
func (s *ProductService) GetBestDealByDiscountedPrice(limit int) ([]models.ProductModel, error) {
	var products []models.ProductModel
	err := s.db.Joins("JOIN discounts ON products.id = discounts.product_id").
		Where("discounts.is_active = ? AND discounts.start_date <= ?", true, time.Now()).
		Where("discounts.end_date IS NULL OR discounts.end_date >= ?", time.Now()).
		Select("products.*, (products.price - CASE WHEN discounts.type = 'PERCENTAGE' THEN products.price * discounts.value / 100 ELSE discounts.value END) as discounted_price").
		Order("discounted_price ASC").
		Limit(limit).
		Find(&products).Error
	return products, err
}

// CalculateDiscountedPrice calculates the discounted price for a given product ID and original price.
//
// It takes a product ID and an original price as parameters and returns the discounted price, discount amount, discount percentage/amount, and discount type, as well as an error if the operation fails.
//
// The method first retrieves all active discounts for the given product ID using GetActiveDiscounts. If there are no active discounts, it returns the original price and a nil error.
//
// If there are active discounts, it takes the first discount in the list and calculates the discounted price by subtracting the discount amount from the original price. The discount amount is calculated based on the discount type, which can be a percentage or an amount.
//
// The function then pastes the discounted price to ensure it is not negative and returns it along with the discount amount, discount percentage/amount, discount type, and a nil error if the operation is successful.
func (s *ProductService) CalculateDiscountedPrice(productID string, originalPrice float64) (float64, float64, float64, string, error) {
	// fmt.Println("ORIGINAL_PRICE", originalPrice)
	// Dapatkan diskon aktif untuk produk
	discounts, err := s.GetActiveDiscounts(productID)
	if err != nil {
		return 0, 0, 0, "", err
	}

	if len(discounts) == 0 {
		return originalPrice, 0, 0, "", nil
	}

	// Hitung harga setelah diskon
	discountedPrice := originalPrice
	discount := discounts[0]
	discAmount := float64(0)
	switch discount.Type {
	case models.DiscountPercentage:
		discAmount = originalPrice * (discount.Value / 100)
		discountedPrice -= discAmount
	case models.DiscountAmount:
		discAmount = discount.Value
		discountedPrice -= discAmount
	}

	// Pastikan harga tidak negatif
	if discountedPrice < 0 {
		discountedPrice = 0
	}

	fmt.Println("DISCOUNT_AMOUNT", discAmount)

	return discountedPrice, discAmount, discount.Value, string(discount.Type), nil
}

// UpdateDiscount updates an existing discount record in the database.
//
// The function takes a string discount ID and a pointer to a DiscountModel
// as parameters and attempts to update the corresponding discount in the
// database. It returns an error if the update operation fails.
//
// It first validates the start and end dates of the discount. If the end date
// is before the start date, it returns an error. Otherwise, it proceeds with
// the update operation.
//
// Parameters:
//
//	discountID (string): The ID of the discount record to be updated.
//	data (*models.DiscountModel): The discount data to be updated.
//
// Returns:
//
//	error: An error object if the update operation fails, or nil if it is
//	  successful.
func (s *ProductService) UpdateDiscount(discountID string, data models.DiscountModel) error {
	// Validasi tanggal
	if data.EndDate != nil && data.StartDate.After(*data.EndDate) {
		return errors.New("start date must be before end date")
	}

	return s.db.Model(&models.DiscountModel{}).Where("id = ?", discountID).Save(&data).Error
}

// DeactivateDiscount deactivates a discount record in the database.
//
// The function takes a string discount ID as parameter and attempts to
// deactivate the corresponding discount in the database. It returns an
// error if the operation fails.
//
// Parameters:
//
//	discountID (string): The ID of the discount record to be deactivated.
//
// Returns:
//
//	error: An error object if the update operation fails, or nil if it is
//	  successful.
func (s *ProductService) DeactivateDiscount(discountID string) error {
	return s.db.Model(&models.DiscountModel{}).Where("id = ?", discountID).Update("is_active", false).Error
}

// AddProductTagByName adds a tag to a product.
//
// If the tag does not exist, it is created. If it already exists, the existing
// tag is used. The tag is then added to the product's tags.
//
// Parameters:
//
//	productID (string): The ID of the product to add the tag to.
//	name (string): The name of the tag to add.
//
// Returns:
//
//	*models.TagModel: The tag that was added to the product.
//	error: An error object if the operation fails, or nil if it is successful.
func (s *ProductService) AddProductTagByName(productID string, name string) (*models.TagModel, error) {
	tag, err := s.tagService.GetTagByName(name)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			tag = &models.TagModel{
				Name:  name,
				Color: "#F5F5F5",
			}
			err = s.db.Create(tag).Error
			if err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}
	var product models.ProductModel
	s.db.Find(&product, "id = ?", productID)

	// Tambahkan tag ke produk
	err = s.db.Model(&product).Association("Tags").Append(tag)
	return tag, err
}

// AddVariantTagByName adds a tag to a product variant.
//
// If the tag does not exist, it is created. If it already exists, the existing
// tag is used. The tag is then added to the product variant's tags.
//
// Parameters:
//
//	productID (string): The ID of the product variant to add the tag to.
//	name (string): The name of the tag to add.
//
// Returns:
//
//	error: An error object if the operation fails, or nil if it is successful.
func (s *ProductService) AddVariantTagByName(productID string, name string) error {
	tag, err := s.tagService.GetTagByName(name)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			tag = &models.TagModel{
				Name: name,
			}
			err = s.db.Create(tag).Error
			if err != nil {
				return err
			}
		} else {
			return err
		}
	}

	// Tambahkan tag ke produk
	err = s.db.Model(&models.VariantModel{}).Where("id = ?", productID).Association("Tags").Append(tag)
	return err
}

// CreateProductFeedback creates a new product feedback.
//
// Parameters:
//
//	data (*models.ProductFeedbackModel): The data to create the new product feedback.
//
// Returns:
//
//	error: An error object if the operation fails, or nil if it is successful.
func (s *ProductService) CreateProductFeedback(data *models.ProductFeedbackModel) error {
	return s.db.Create(data).Error
}

// DeleteProductFeedback deletes a product feedback by its ID.
//
// Parameters:
//
//	id (string): The ID of the product feedback to be deleted.
//
// Returns:
//
//	error: An error if the deletion operation fails, or nil if it is successful.
func (s *ProductService) DeleteProductFeedback(id string) error {
	return s.db.Delete(&models.ProductFeedbackModel{}, "id = ?", id).Error
}

// GetBestSellingProduct retrieves a list of best selling products, sorted by total sale descending.
//
// Parameters:
//
//	request (*http.Request): The HTTP request object, used to filter the result based on the company ID in the headers.
//	limit (int): The maximum number of records to return.
//	warehouseID (*string): The ID of the warehouse to filter the result. If nil, the result is not filtered.
//
// Returns:
//
//	[]models.PopularProduct: A list of PopularProduct, each containing the product ID, display name, and total sale.
//	error: An error object if the operation fails, or nil if it is successful.
func (s *ProductService) GetBestSellingProduct(request *http.Request, limit int, warehouseID *string) ([]models.PopularProduct, error) {
	var results []models.PopularProduct
	db := s.db.Model(&models.StockMovementModel{})
	if request != nil {
		if request.Header.Get("ID-Company") != "" {
			if request.Header.Get("ID-Company") == "nil" || request.Header.Get("ID-Company") == "null" {
				db = db.Where("products.company_id is null")
			} else {
				db = db.Where("products.company_id = ?", request.Header.Get("ID-Company"))

			}
		}
		// if request.Header.Get("ID-Distributor") != "" {
		// 	db = db.Where("products.company_id = ?", request.Header.Get("ID-Distributor"))
		// }
	}

	if warehouseID != nil {
		db = db.Where("warehouse_id = ?", *warehouseID)
	}
	db = db.Where("type in (?)", []models.MovementType{models.MovementTypeReturn, models.MovementTypeSale})

	if err := db.
		Joins("JOIN products on stock_movements.product_id = products.id").
		Select("products.id", "products.display_name", "COALESCE(SUM(quantity * value), 0) * -1 as total_sale").
		Group("products.id").
		Order("COALESCE(SUM(quantity * value), 0) DESC").
		Limit(limit).
		Scan(&results).Error; err != nil {
		return nil, err
	}

	return results, nil
}
