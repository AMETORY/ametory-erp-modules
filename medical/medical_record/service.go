package medical_record

import (
	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/shared/models"
	"gorm.io/gorm"
)

// MedicalRecordService is the service for MedicalRecord model
type MedicalRecordService struct {
	ctx *context.ERPContext
	db  *gorm.DB
}

// NewMedicalRecordService will create new MedicalRecordService
//
// This method will create a new instance of MedicalRecordService with the given database connection and context.
func NewMedicalRecordService(db *gorm.DB, ctx *context.ERPContext) *MedicalRecordService {
	return &MedicalRecordService{db: db, ctx: ctx}
}

// CreateMedicalRecord creates a new MedicalRecord in the database.
//
// It takes a pointer to a MedicalRecordModel as an argument and returns an error
// if the creation fails.
func (s *MedicalRecordService) CreateMedicalRecord(record *models.MedicalRecordModel) error {
	if err := s.db.Create(record).Error; err != nil {
		return err
	}
	return nil
}

// GetMedicalRecordByID retrieves a MedicalRecord by its ID from the database.
//
// It takes a string argument representing the ID of the MedicalRecord and returns
// a pointer to a MedicalRecordModel and an error. If the retrieval fails, it
// returns an error.
func (s *MedicalRecordService) GetMedicalRecordByID(ID string) (*models.MedicalRecordModel, error) {
	var record models.MedicalRecordModel
	if err := s.db.Where("id = ?", ID).First(&record).Error; err != nil {
		return nil, err
	}
	return &record, nil
}

// GetMedicalRecordByPatientID retrieves a list of MedicalRecords associated with the specified patient ID from the database.
//
// It takes a string argument representing the patient ID and returns a slice of
// MedicalRecordModel and an error. If the retrieval fails, it returns an error.
func (s *MedicalRecordService) GetMedicalRecordByPatientID(patientID string) ([]models.MedicalRecordModel, error) {
	var records []models.MedicalRecordModel
	if err := s.db.Where("patient_id = ?", patientID).Find(&records).Error; err != nil {
		return nil, err
	}
	return records, nil
}

// UpdateMedicalRecord updates a MedicalRecord in the database.
//
// It takes a pointer to a MedicalRecordModel as an argument and returns an error
// if the update fails.
func (s *MedicalRecordService) UpdateMedicalRecord(record *models.MedicalRecordModel) error {
	if err := s.db.Save(record).Error; err != nil {
		return err
	}
	return nil
}

// DeleteMedicalRecord deletes a MedicalRecord from the database.
//
// It takes a string argument representing the ID of the MedicalRecord to be
// deleted and returns an error if the deletion fails.
func (s *MedicalRecordService) DeleteMedicalRecord(ID string) error {
	if err := s.db.Delete(&models.MedicalRecordModel{}, "id = ?", ID).Error; err != nil {
		return err
	}
	return nil
}
