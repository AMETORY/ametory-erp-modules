package product

import (
	"errors"
	"net/http"

	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/shared"
	"github.com/AMETORY/ametory-erp-modules/utils"
	"github.com/morkid/paginate"
	"gorm.io/gorm"
)

type MasterProductService struct {
	db  *gorm.DB
	ctx *context.ERPContext
}

func NewMasterProductService(db *gorm.DB, ctx *context.ERPContext) *MasterProductService {
	return &MasterProductService{db: db, ctx: ctx}
}

func (s *MasterProductService) CreateMasterProduct(data *MasterProductModel) error {
	return s.db.Create(data).Error
}

func (s *MasterProductService) UpdateMasterProduct(id string, data *MasterProductModel) error {
	return s.db.Where("id = ?", id).Updates(data).Error
}

func (s *MasterProductService) DeleteMasterProduct(id string) error {
	return s.db.Where("id = ?", id).Delete(&MasterProductModel{}).Error
}

func (s *MasterProductService) GetMasterProductByID(id string) (*MasterProductModel, error) {
	var product MasterProductModel
	err := s.db.Where("id = ?", id).First(&product).Error
	product.Prices, _ = s.ListPricesOfProduct(product.ID)
	product.ProductImages, _ = s.ListImagesOfProduct(product.ID)
	return &product, err
}

func (s *MasterProductService) GetMasterProductByCode(code string) (*MasterProductModel, error) {
	var product MasterProductModel
	err := s.db.Where("code = ?", code).First(&product).Error
	return &product, err
}

func (s *MasterProductService) GetMasterProducts(request http.Request, search string) (paginate.Page, error) {
	pg := paginate.New()
	stmt := s.db
	if search != "" {
		stmt = stmt.Where("master_products.description ILIKE ? OR master_products.sku ILIKE ? OR master_products.name ILIKE ? OR master_products.barcode ILIKE ?",
			"%"+search+"%",
			"%"+search+"%",
			"%"+search+"%",
			"%"+search+"%",
		)
	}
	if request.Header.Get("ID-Company") != "" {
		stmt = stmt.Where("company_id = ?", request.Header.Get("ID-Company"))
	}
	stmt = stmt.Model(&MasterProductModel{})
	utils.FixRequest(&request)
	page := pg.With(stmt).Request(request).Response(&[]MasterProductModel{})
	page.Page = page.Page + 1
	items := page.Items.(*[]MasterProductModel)
	newItems := make([]MasterProductModel, 0)
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

func (s *MasterProductService) CreatePriceCategory(data *PriceCategoryModel) error {
	return s.db.Create(data).Error
}

func (s *MasterProductService) AddPriceToMasterProduct(productID string, data *MasterProductPriceModel) error {
	if data.PriceCategoryID == "" {
		return errors.New("price category id is required")
	}
	data.MasterProductID = productID
	return s.db.Create(data).Error
}

func (s *MasterProductService) ListPricesOfProduct(productID string) ([]MasterProductPriceModel, error) {
	var prices []MasterProductPriceModel
	err := s.db.Preload("PriceCategory").Where("master_product_id = ?", productID).Find(&prices).Error
	return prices, err
}

func (s *MasterProductService) ListImagesOfProduct(productID string) ([]shared.FileModel, error) {
	var images []shared.FileModel
	err := s.db.Where("ref_id = ? and ref_type = ?", productID, "master-product").Find(&images).Error
	return images, err
}

func (s *MasterProductService) DeletePriceFromMasterProduct(productID string, priceID string) error {
	return s.db.Where("master_product_id = ? and id = ?", productID, priceID).Delete(&MasterProductPriceModel{}).Error
}

func (s *MasterProductService) DeleteImageFromMasterProduct(productID string, imageID string) error {
	return s.db.Where("ref_id = ? and ref_type = ? and id = ?", productID, "master-product", imageID).Delete(&shared.FileModel{}).Error
}
