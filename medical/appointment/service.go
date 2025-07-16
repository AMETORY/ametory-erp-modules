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

// NewAppointmentService creates a new instance of AppointmentService with the specified
// database connection and ERP context. It initializes the service to handle
// operations related to medical appointments.
func NewAppointmentService(db *gorm.DB, ctx *context.ERPContext) *AppointmentService {
	return &AppointmentService{db: db, ctx: ctx}
}

// CreateAppointment creates a new medical appointment for a patient in the database.
// It takes a pointer to a MedicalAppointmentModel as an argument and returns an error
// if the creation fails.
func (s *AppointmentService) CreateAppointment(appointment *models.MedicalAppointmentModel) error {
	if err := s.db.Create(appointment).Error; err != nil {
		return err
	}
	return nil
}

// GetAppointmentsByPatientID retrieves a list of medical appointments for a
// patient with the specified ID from the database. It takes a string
// argument representing the patient ID and returns a slice of
// MedicalAppointmentModel and an error. If the retrieval fails, it returns
// an error.
func (s *AppointmentService) GetAppointmentsByPatientID(patientID string) ([]models.MedicalAppointmentModel, error) {
	var appointments []models.MedicalAppointmentModel
	err := s.db.Where("patient_id = ?", patientID).Find(&appointments).Error
	if err != nil {
		return nil, err
	}
	return appointments, nil
}

// GetAppointmentsByHealthFacilityID retrieves a list of medical appointments
// associated with the specified health facility ID. It performs a join
// operation with the sub_facilities table to filter appointments based on
// the facility's ID. The function returns a slice of MedicalAppointmentModel
// and an error if the retrieval fails.
func (s *AppointmentService) GetAppointmentsByHealthFacilityID(facilityID string) ([]models.MedicalAppointmentModel, error) {
	var appointments []models.MedicalAppointmentModel
	err := s.db.Joins("JOIN sub_facilities ON sub_facilities.id = medical_appointments.sub_facility_id").
		Where("sub_facilities.facility_id = ?", facilityID).Find(&appointments).Error
	if err != nil {
		return nil, err
	}
	return appointments, nil
}

// GetAppointmentsBySubFacilityID retrieves a list of medical appointments associated with the
// specified sub-facility ID. It takes a string argument representing the sub-facility ID
// and returns a slice of MedicalAppointmentModel and an error. If the retrieval fails,
// it returns an error.
func (s *AppointmentService) GetAppointmentsBySubFacilityID(subFacilityID string) ([]models.MedicalAppointmentModel, error) {
	var appointments []models.MedicalAppointmentModel
	err := s.db.Where("sub_facility_id = ?", subFacilityID).Find(&appointments).Error
	if err != nil {
		return nil, err
	}
	return appointments, nil
}
