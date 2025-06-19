package deduction_setting

import (
	"net/http"

	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/shared/models"
	"github.com/AMETORY/ametory-erp-modules/utils"
	"github.com/morkid/paginate"
	"gorm.io/gorm"
)

type DeductionSettingService struct {
	db  *gorm.DB
	ctx *context.ERPContext
}

func NewDeductionSettingService(ctx *context.ERPContext) *DeductionSettingService {
	return &DeductionSettingService{db: ctx.DB, ctx: ctx}
}

func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&models.DeductionSettingModel{},
	)
}

func (s *DeductionSettingService) Create(deductionSetting *models.DeductionSettingModel) error {
	return s.db.Create(deductionSetting).Error
}

func (a *DeductionSettingService) FindAll(request *http.Request) (paginate.Page, error) {
	pg := paginate.New()
	stmt := a.db.Model(&models.DeductionSettingModel{})
	utils.FixRequest(request)
	page := pg.With(stmt).Request(request).Response(&[]models.DeductionSettingModel{})
	page.Page = page.Page + 1
	return page, nil
}

func (s *DeductionSettingService) FindOne(id string) (*models.DeductionSettingModel, error) {
	deductionSetting := &models.DeductionSettingModel{}

	db := s.db.Model(&models.DeductionSettingModel{})

	if err := db.Where("id = ?", id).First(deductionSetting).Error; err != nil {
		return nil, err
	}

	return deductionSetting, nil
}

func (s *DeductionSettingService) Update(deductionSetting *models.DeductionSettingModel) error {
	return s.db.Save(deductionSetting).Error
}

func (s *DeductionSettingService) Delete(id string) error {
	return s.db.Delete(&models.DeductionSettingModel{}, "id = ?", id).Error
}
