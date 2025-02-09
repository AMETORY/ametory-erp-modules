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
func NewMedicalRecordService(db *gorm.DB, ctx *context.ERPContext) *MedicalRecordService {
	return &MedicalRecordService{db: db, ctx: ctx}
}

// CreateMedicalRecord will create new MedicalRecord
func (s *MedicalRecordService) CreateMedicalRecord(record *models.MedicalRecordModel) error {
	if err := s.db.Create(record).Error; err != nil {
		return err
	}
	return nil
}

// GetMedicalRecordByID will find MedicalRecord by ID
func (s *MedicalRecordService) GetMedicalRecordByID(ID string) (*models.MedicalRecordModel, error) {
	var record models.MedicalRecordModel
	if err := s.db.Where("id = ?", ID).First(&record).Error; err != nil {
		return nil, err
	}
	return &record, nil
}

// GetMedicalRecordByPatientID will find MedicalRecord by PatientID
func (s *MedicalRecordService) GetMedicalRecordByPatientID(patientID string) ([]models.MedicalRecordModel, error) {
	var records []models.MedicalRecordModel
	if err := s.db.Where("patient_id = ?", patientID).Find(&records).Error; err != nil {
		return nil, err
	}
	return records, nil
}

// UpdateMedicalRecord will update MedicalRecord
func (s *MedicalRecordService) UpdateMedicalRecord(record *models.MedicalRecordModel) error {
	if err := s.db.Save(record).Error; err != nil {
		return err
	}
	return nil
}

// DeleteMedicalRecord will delete MedicalRecord
func (s *MedicalRecordService) DeleteMedicalRecord(ID string) error {
	if err := s.db.Delete(&models.MedicalRecordModel{}, "id = ?", ID).Error; err != nil {
		return err
	}
	return nil
}
