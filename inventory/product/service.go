package product

import (
	"net/http"

	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/shared"
	"github.com/AMETORY/ametory-erp-modules/utils"
	"github.com/morkid/paginate"
	"gorm.io/gorm"
)

type ProductService struct {
	db          *gorm.DB
	ctx         *context.ERPContext
	fileService *shared.FileService
}
type ProductCategoryService struct {
	db  *gorm.DB
	ctx *context.ERPContext
}

func NewProductService(db *gorm.DB, ctx *context.ERPContext, fileService *shared.FileService) *ProductService {
	return &ProductService{db: db, ctx: ctx, fileService: fileService}
}
func NewProductCategoryService(db *gorm.DB, ctx *context.ERPContext) *ProductCategoryService {
	return &ProductCategoryService{db: db, ctx: ctx}
}

func (s *ProductService) CreateProduct(data *ProductModel) error {
	return s.db.Create(data).Error
}

func (s *ProductService) UpdateProduct(id string, data *ProductModel) error {
	return s.db.Where("id = ?", id).Updates(data).Error
}

func (s *ProductService) DeleteProduct(id string) error {
	return s.db.Where("id = ?", id).Delete(&ProductModel{}).Error
}

func (s *ProductService) GetProductByID(id string) (*ProductModel, error) {
	var product ProductModel
	err := s.db.Where("id = ?", id).First(&product).Error
	product.Prices, _ = s.ListPricesOfProduct(product.ID)
	product.ProductImages, _ = s.ListImagesOfProduct(product.ID)
	return &product, err
}

func (s *ProductService) GetProductByCode(code string) (*ProductModel, error) {
	var product ProductModel
	err := s.db.Where("sku = ?", code).First(&product).Error
	product.Prices, _ = s.ListPricesOfProduct(product.ID)
	product.ProductImages, _ = s.ListImagesOfProduct(product.ID)
	return &product, err
}

func (s *ProductService) GetProducts(request http.Request, search string) (paginate.Page, error) {
	pg := paginate.New()
	stmt := s.db
	if search != "" {
		stmt = stmt.Where("products.description LIKE ? OR products.sku LIKE ? OR products.name LIKE ? OR products.barcode LIKE ?",
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
	stmt = stmt.Model(&ProductModel{})
	utils.FixRequest(&request)
	page := pg.With(stmt).Request(request).Response(&[]ProductModel{})
	page.Page = page.Page + 1
	return page, nil
}

func (s *ProductService) CreatePriceCategory(data *PriceCategoryModel) error {
	return s.db.Create(data).Error
}

func (s *ProductService) AddPriceToProduct(data *PriceModel) error {
	return s.db.Create(data).Error
}

func (s *ProductService) ListPricesOfProduct(productID string) ([]PriceModel, error) {
	var prices []PriceModel
	err := s.db.Where("product_id = ?", productID).Find(&prices).Error
	return prices, err
}

func (s *ProductService) ListImagesOfProduct(productID string) ([]shared.FileModel, error) {
	var images []shared.FileModel
	err := s.db.Where("ref_id = ?", productID).Find(&images).Error
	return images, err
}
