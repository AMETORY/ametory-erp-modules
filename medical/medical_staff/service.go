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

// NewMedicalStaffService creates a new instance of MedicalStaffService with the given database connection and context.
//
// The ERP context is used for authentication and authorization, while
// the database is used for CRUD operations on medical staff-related data.
func NewMedicalStaffService(db *gorm.DB, ctx *context.ERPContext) *MedicalStaffService {
	return &MedicalStaffService{db: db, ctx: ctx}
}

// CreateMedicalStaff creates a new medical staff in the database.
//
// It takes a pointer to a MedicalStaffModel as an argument and returns an error
// if the creation fails.
func (r *MedicalStaffService) CreateMedicalStaff(staff *models.MedicalStaffModel) error {
	return r.db.Create(staff).Error
}

// GetMedicalStaff retrieves a medical staff member by their ID.
//
// It takes a string argument representing the medical staff member's ID and returns
// a pointer to the MedicalStaffModel and an error if the retrieval fails.
func (r *MedicalStaffService) GetMedicalStaff(id string) (*models.MedicalStaffModel, error) {
	var staff models.MedicalStaffModel
	err := r.db.First(&staff, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &staff, nil
}

// UpdateMedicalStaff updates an existing medical staff member's details.
//
// It takes a string argument representing the medical staff member's ID and a pointer
// to a MedicalStaffModel containing updated fields. It returns an error if the update fails.
func (r *MedicalStaffService) UpdateMedicalStaff(id string, updatedStaff *models.MedicalStaffModel) error {
	return r.db.Model(&models.MedicalStaffModel{}).Where("id = ?", id).Updates(updatedStaff).Error
}

// DeleteMedicalStaff deletes a medical staff member from the database.
//
// It takes a string argument representing the medical staff member's ID and returns an error
// if the deletion fails.
func (r *MedicalStaffService) DeleteMedicalStaff(id string) error {
	return r.db.Delete(&models.MedicalStaffModel{}, "id = ?", id).Error
}

// AssignToSubFacility assigns a medical staff member to a specified sub-facility.
//
// It takes two string arguments: subFacilityID and staffID, representing the IDs
// of the sub-facility and the medical staff member, respectively. It creates a new
// SubFacilityStaff record in the database and returns an error if the creation fails.

func (r *MedicalStaffService) AssignToSubFacility(subFacilityID, staffID string) error {
	assignment := &models.SubFacilityStaff{
		BaseModel:           shared.BaseModel{ID: uuid.NewString()},
		SubFacilityModelID:  subFacilityID,
		MedicalStaffModelID: staffID,
	}
	return r.db.Create(assignment).Error
}
