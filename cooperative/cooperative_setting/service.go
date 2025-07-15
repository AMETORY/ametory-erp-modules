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

// NewCooperativeSettingService creates a new instance of CooperativeSettingService.
//
// The service is created by providing a GORM database instance and an ERP context.
// The ERP context is used for authentication and authorization purposes, while the
// database instance is used for CRUD (Create, Read, Update, Delete) operations.
func NewCooperativeSettingService(db *gorm.DB, ctx *context.ERPContext) *CooperativeSettingService {
	return &CooperativeSettingService{
		db:  db,
		ctx: ctx,
	}
}

// GetSetting retrieves the cooperative setting for a given company ID.
//
// If the setting does not exist, the function returns a gorm.ErrRecordNotFound error.
// Otherwise, the function returns the setting and a nil error.
func (s *CooperativeSettingService) GetSetting(companyID *string) (*models.CooperativeSettingModel, error) {
	var setting models.CooperativeSettingModel
	err := s.ctx.DB.Where("company_id = ?", companyID).First(&setting).Error
	if err != nil {
		return nil, err
	}
	return &setting, nil
}
