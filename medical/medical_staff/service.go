package medical_staff

import (
	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/shared"
	"github.com/AMETORY/ametory-erp-modules/shared/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type MedicalStaffService struct {
	db  *gorm.DB
	ctx *context.ERPContext
}

func NewMedicalStaffService(db *gorm.DB, ctx *context.ERPContext) *MedicalStaffService {
	return &MedicalStaffService{db: db, ctx: ctx}
}

func (r *MedicalStaffService) Create(staff *models.MedicalStaffModel) error {
	return r.db.Create(staff).Error
}

func (r *MedicalStaffService) AssignToSubFacility(subFacilityID, staffID string) error {
	assignment := &models.SubFacilityStaff{
		BaseModel:           shared.BaseModel{ID: uuid.NewString()},
		SubFacilityModelID:  subFacilityID,
		MedicalStaffModelID: staffID,
	}
	return r.db.Create(assignment).Error
}
