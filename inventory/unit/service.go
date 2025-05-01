package unit

import (
	"errors"
	"net/http"

	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/shared/models"
	"github.com/AMETORY/ametory-erp-modules/utils"
	"github.com/morkid/paginate"
	"gorm.io/gorm"
)

type UnitService struct {
	db  *gorm.DB
	ctx *context.ERPContext
}

func NewUnitService(db *gorm.DB, ctx *context.ERPContext) *UnitService {
	return &UnitService{db: db, ctx: ctx}
}

func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&models.UnitModel{},
		&models.ProductUnits{},
	)
}

func (s *UnitService) CreateUnit(data *models.UnitModel) error {
	return s.db.Create(data).Error
}

func (s *UnitService) UpdateUnit(id string, data *models.UnitModel) error {
	return s.db.Where("id = ?", id).Updates(data).Error
}

func (s *UnitService) DeleteUnit(id string) error {
	return s.db.Where("id = ?", id).Delete(&models.UnitModel{}).Error
}

func (s *UnitService) GetUnitByID(id string) (*models.UnitModel, error) {
	var invoice models.UnitModel
	err := s.db.Where("id = ?", id).First(&invoice).Error
	return &invoice, err
}

func (s *UnitService) GetUnits(request http.Request, search string) (paginate.Page, error) {
	pg := paginate.New()
	stmt := s.db
	if search != "" {
		stmt = stmt.Where("name ILIKE ? OR code ILIKE ? OR description ILIKE ?",
			"%"+search+"%",
			"%"+search+"%",
			"%"+search+"%",
		)
	}
	if request.Header.Get("ID-Company") != "" {
		stmt = stmt.Where("company_id = ? or company_id is null", request.Header.Get("ID-Company"))
	}
	request.URL.Query().Get("page")
	stmt = stmt.Model(&models.UnitModel{})
	utils.FixRequest(&request)
	page := pg.With(stmt).Request(request).Response(&[]models.UnitModel{})
	page.Page = page.Page + 1
	return page, nil
}

func (s *UnitService) GetUnitByName(name string) (*models.UnitModel, error) {
	var brand models.UnitModel
	err := s.db.Where("name = ?", name).First(&brand).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			brand = models.UnitModel{
				Name: name,
			}
			err := s.db.Create(&brand).Error
			if err != nil {
				return nil, err
			}
			return &brand, nil
		}
		return nil, err
	}
	return &brand, nil
}
