package medical

import (
	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/medical/appointment"
	"github.com/AMETORY/ametory-erp-modules/medical/healh_facility"
	"github.com/AMETORY/ametory-erp-modules/medical/medical_record"
	"github.com/AMETORY/ametory-erp-modules/medical/medical_staff"
	"github.com/AMETORY/ametory-erp-modules/medical/patient"
	"github.com/AMETORY/ametory-erp-modules/medical/pharmacy"
	"github.com/AMETORY/ametory-erp-modules/medical/treatment_queue"
	"github.com/AMETORY/ametory-erp-modules/shared/models"
	"gorm.io/gorm"
)

type MedicalService struct {
	db                    *gorm.DB
	ctx                   *context.ERPContext
	patientService        *patient.PatientService
	healthFacilityService *healh_facility.HeathFacilityService
	medicalStaffService   *medical_staff.MedicalStaffService
	appointmentService    *appointment.AppointmentService
	medicalRecord         *medical_record.MedicalRecordService
	pharmacyService       *pharmacy.PharmacyService
	treatmentQueue        *treatment_queue.TreatmentQueueService
}

func NewMedicalService(db *gorm.DB, ctx *context.ERPContext) *MedicalService {
	service := MedicalService{
		db:                    db,
		ctx:                   ctx,
		patientService:        patient.NewPatientService(db, ctx),
		healthFacilityService: healh_facility.NewHeathFacilityService(db, ctx),
		medicalStaffService:   medical_staff.NewMedicalStaffService(db, ctx),
		appointmentService:    appointment.NewAppointmentService(db, ctx),
		medicalRecord:         medical_record.NewMedicalRecordService(db, ctx),
		pharmacyService:       pharmacy.NewPharmacyService(db, ctx),
		treatmentQueue:        treatment_queue.NewTreatmentQueueService(db, ctx),
	}
	service.Migrate()
	return &service
}

func (s *MedicalService) Migrate() error {
	if s.ctx.SkipMigration {
		return nil
	}
	s.db.AutoMigrate(
		&models.PatientModel{},
		&models.HealthFacilityModel{},
		&models.SubFacilityModel{},
		&models.MedicalStaffModel{},
		&models.DoctorModel{},
		&models.NurseModel{},
		&models.MedicalAppointmentModel{},
		&models.MedicalRecordModel{},
		&models.MedicineModel{},
		&models.PharmacyModel{},
	)
	return nil
}
