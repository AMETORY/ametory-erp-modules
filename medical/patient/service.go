package patient

import (
	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/shared/models"
	"gorm.io/gorm"
)

type PatientService struct {
	db  *gorm.DB
	ctx *context.ERPContext
}

func NewPatientService(db *gorm.DB, ctx *context.ERPContext) *PatientService {
	return &PatientService{db: db, ctx: ctx}
}

// CreatePatient creates a new patient in the database.
//
// It takes a pointer to a PatientModel as an argument and returns an error
// if the creation fails.
func (s *PatientService) CreatePatient(patient *models.PatientModel) error {
	return s.db.Create(patient).Error
}

// GetPatientByID retrieves a patient by its ID from the database.
//
// It takes a string argument representing the patient ID and returns a pointer
// to a PatientModel and an error. If the retrieval fails, it returns an error.
func (s *PatientService) GetPatientByID(ID string) (*models.PatientModel, error) {
	p := &models.PatientModel{}
	err := s.db.Where("id = ?", ID).First(p).Error
	if err != nil {
		return nil, err
	}

	return p, nil
}

// UpdatePatient updates a patient in the database.
//
// It takes a pointer to a PatientModel as an argument and returns an error
// if the update fails.
func (s *PatientService) UpdatePatient(patient *models.PatientModel) error {
	return s.db.Save(patient).Error
}

// DeletePatient deletes a patient from the database.
//
// It takes a string argument representing the patient ID and returns an error
// if the deletion fails.
func (s *PatientService) DeletePatient(ID string) error {
	return s.db.Where("id = ?", ID).Delete(&models.PatientModel{}).Error
}
