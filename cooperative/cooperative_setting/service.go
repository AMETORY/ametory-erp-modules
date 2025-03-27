package cooperative_setting

import (
	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/shared/models"
	"gorm.io/gorm"
)

type CooperativeSettingService struct {
	db  *gorm.DB
	ctx *context.ERPContext
}

func NewCooperativeSettingService(db *gorm.DB, ctx *context.ERPContext) *CooperativeSettingService {
	return &CooperativeSettingService{
		db:  db,
		ctx: ctx,
	}
}

func (s *CooperativeSettingService) GetSetting(companyID *string) (*models.CooperativeSettingModel, error) {
	var setting models.CooperativeSettingModel
	err := s.ctx.DB.Where("company_id = ?", companyID).First(&setting).Error
	if err != nil {
		return nil, err
	}
	return &setting, nil
}
