package appointment

import (
	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/shared/models"
	"gorm.io/gorm"
)

type AppointmentService struct {
	db  *gorm.DB
	ctx *context.ERPContext
}

func NewAppointmentService(db *gorm.DB, ctx *context.ERPContext) *AppointmentService {
	return &AppointmentService{db: db, ctx: ctx}
}

func (s *AppointmentService) CreateAppointment(appointment *models.MedicalAppointmentModel) error {
	if err := s.db.Create(appointment).Error; err != nil {
		return err
	}
	return nil
}

func (s *AppointmentService) GetAppointmentsByPatientID(patientID string) ([]models.MedicalAppointmentModel, error) {
	var appointments []models.MedicalAppointmentModel
	err := s.db.Where("patient_id = ?", patientID).Find(&appointments).Error
	if err != nil {
		return nil, err
	}
	return appointments, nil
}

func (s *AppointmentService) GetAppointmentsByHealthFacilityID(facilityID string) ([]models.MedicalAppointmentModel, error) {
	var appointments []models.MedicalAppointmentModel
	err := s.db.Joins("JOIN sub_facilities ON sub_facilities.id = medical_appointments.sub_facility_id").
		Where("sub_facilities.facility_id = ?", facilityID).Find(&appointments).Error
	if err != nil {
		return nil, err
	}
	return appointments, nil
}

func (s *AppointmentService) GetAppointmentsBySubFacilityID(subFacilityID string) ([]models.MedicalAppointmentModel, error) {
	var appointments []models.MedicalAppointmentModel
	err := s.db.Where("sub_facility_id = ?", subFacilityID).Find(&appointments).Error
	if err != nil {
		return nil, err
	}
	return appointments, nil
}
