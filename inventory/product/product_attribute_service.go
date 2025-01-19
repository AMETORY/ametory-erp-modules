package product

import (
	"net/http"

	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/shared/models"
	"github.com/AMETORY/ametory-erp-modules/utils"
	"github.com/morkid/paginate"
	"gorm.io/gorm"
)

type ProductAttributeService struct {
	db  *gorm.DB
	ctx *context.ERPContext
}

func NewProductAttributeService(db *gorm.DB, ctx *context.ERPContext) *ProductAttributeService {
	return &ProductAttributeService{db: db, ctx: ctx}
}

func (s *ProductAttributeService) CreateProductAttribute(data *models.ProductAttributeModel) error {
	return s.db.Create(data).Error
}

func (s *ProductAttributeService) UpdateProductAttribute(id string, data *models.ProductAttributeModel) error {
	return s.db.Model(&models.ProductAttributeModel{}).Where("id = ?", id).Updates(data).Error
}

func (s *ProductAttributeService) DeleteProductAttribute(id string) error {
	return s.db.Delete(&models.ProductAttributeModel{}, "id = ?", id).Error
}

func (s *ProductAttributeService) GetProductAttributeByID(id string) (*models.ProductAttributeModel, error) {
	var attribute models.ProductAttributeModel
	err := s.db.First(&attribute, "id = ?", id).Error
	return &attribute, err
}

func (s *ProductAttributeService) GetProductAttributes(request http.Request, search string) (paginate.Page, error) {
	pg := paginate.New()
	stmt := s.db
	if search != "" {
		stmt = stmt.Where("product_attributes.name ILIKE ? ",
			"%"+search+"%",
		)
	}
	stmt = stmt.Model(&models.ProductAttributeModel{})
	utils.FixRequest(&request)
	page := pg.With(stmt).Request(request).Response(&[]models.ProductAttributeModel{})
	page.Page = page.Page + 1
	return page, nil
}
