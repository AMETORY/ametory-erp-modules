package product

import (
	"errors"
	"net/http"

	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/file"
	"github.com/AMETORY/ametory-erp-modules/shared/models"
	"github.com/AMETORY/ametory-erp-modules/utils"
	"github.com/morkid/paginate"
	"gorm.io/gorm"
)

type ProductService struct {
	db          *gorm.DB
	ctx         *context.ERPContext
	fileService *file.FileService
}

func NewProductService(db *gorm.DB, ctx *context.ERPContext, fileService *file.FileService) *ProductService {
	return &ProductService{db: db, ctx: ctx, fileService: fileService}
}

func Migrate(db *gorm.DB) error {
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
	)
}

func (s *ProductService) CreateProduct(data *models.ProductModel) error {
	return s.db.Create(data).Error
}

func (s *ProductService) UpdateProduct(id string, data *models.ProductModel) error {
	return s.db.Where("id = ?", id).Updates(data).Error
}

func (s *ProductService) DeleteProduct(id string) error {
	return s.db.Where("id = ?", id).Delete(&models.ProductModel{}).Error
}

func (s *ProductService) GetProductByID(id string, request *http.Request) (*models.ProductModel, error) {
	var product models.ProductModel
	err := s.db.Preload("MasterProduct").Preload("Category", func(db *gorm.DB) *gorm.DB {
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
	return &product, err
}

func (s *ProductService) GetProductByCode(code string) (*models.ProductModel, error) {
	var product models.ProductModel
	err := s.db.Preload("Category", func(db *gorm.DB) *gorm.DB {
		return db.Select("id", "name")
	}).Preload("Brand", func(db *gorm.DB) *gorm.DB {
		return db.Select("id", "name")
	}).Where("sku = ?", code).First(&product).Error
	product.Prices, _ = s.ListPricesOfProduct(product.ID)
	product.ProductImages, _ = s.ListImagesOfProduct(product.ID)
	return &product, err
}

func (s *ProductService) GetProducts(request http.Request, search string) (paginate.Page, error) {
	pg := paginate.New()
	stmt := s.db.Preload("Category", func(db *gorm.DB) *gorm.DB {
		return db.Select("id", "name")
	}).Preload("Brand", func(db *gorm.DB) *gorm.DB {
		return db.Select("id", "name")
	})
	if search != "" {
		stmt = stmt.Where("products.description ILIKE ? OR products.sku ILIKE ? OR products.name ILIKE ? OR products.barcode ILIKE ?",
			"%"+search+"%",
			"%"+search+"%",
			"%"+search+"%",
			"%"+search+"%",
		)
	}
	if request.Header.Get("ID-Company") != "" {
		stmt = stmt.Where("company_id = ?", request.Header.Get("ID-Company"))
	}
	if request.Header.Get("ID-Distributor") != "" {
		stmt = stmt.Where("company_id = ?", request.Header.Get("ID-Distributor"))
	}
	stmt = stmt.Model(&models.ProductModel{})
	utils.FixRequest(&request)
	page := pg.With(stmt).Request(request).Response(&[]models.ProductModel{})
	page.Page = page.Page + 1
	items := page.Items.(*[]models.ProductModel)
	newItems := make([]models.ProductModel, 0)

	for _, v := range *items {
		img, err := s.ListImagesOfProduct(v.ID)
		if err == nil {
			v.ProductImages = img
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

func (s *ProductService) CreatePriceCategory(data *models.PriceCategoryModel) error {
	return s.db.Create(data).Error
}

func (s *ProductService) AddPriceToProduct(productID string, data *models.PriceModel) error {
	if data.PriceCategoryID == "" {
		return errors.New("price category id is required")
	}
	data.ProductID = productID
	return s.db.Create(data).Error
}

func (s *ProductService) ListPricesOfProduct(productID string) ([]models.PriceModel, error) {
	var prices []models.PriceModel
	err := s.db.Preload("PriceCategory").Where("product_id = ?", productID).Find(&prices).Error
	return prices, err
}

func (s *ProductService) ListImagesOfProduct(productID string) ([]models.FileModel, error) {
	var images []models.FileModel
	err := s.db.Where("ref_id = ? and ref_type = ?", productID, "product").Find(&images).Error
	return images, err
}

func (s *ProductService) DeletePriceOfProduct(productID string, priceID string) error {
	return s.db.Where("product_id = ? and id = ?", productID, priceID).Delete(&models.PriceModel{}).Error
}

func (s *ProductService) DeleteImageOfProduct(productID string, imageID string) error {
	return s.db.Where("ref_id = ? and ref_type = ? and id = ?", productID, "product", imageID).Delete(&models.FileModel{}).Error
}

func (s *ProductService) GetStock(productID string, request *http.Request, warehouseID *string) (float64, error) {

	var totalStock float64
	db := s.db.Table("stock_movements")
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
		Select("COALESCE(SUM(quantity), 0)").
		Scan(&totalStock).Error; err != nil {
		return 0, err
	}

	return totalStock, nil
}

func (s *ProductService) GetProductsByMerchant(merchantID string, productIDs []string) ([]models.ProductModel, error) {
	var products []models.ProductModel
	db := s.db.Where("merchant_id = ?", merchantID)
	if len(productIDs) > 0 {
		db = db.Where("id in (?)", productIDs)
	}
	err := db.Find(&products).Error
	return products, err
}
