package medical

import (
	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/medical/appointment"
	"github.com/AMETORY/ametory-erp-modules/medical/doctor"
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
	PatientService        *patient.PatientService
	HealthFacilityService *healh_facility.HeathFacilityService
	MedicalStaffService   *medical_staff.MedicalStaffService
	AppointmentService    *appointment.AppointmentService
	MedicalRecord         *medical_record.MedicalRecordService
	PharmacyService       *pharmacy.PharmacyService
	TreatmentQueue        *treatment_queue.TreatmentQueueService
	DoctorService         *doctor.DoctorService
}

// NewMedicalService creates a new instance of MedicalService with the given database connection and context.
//
// It returns a pointer to the newly created MedicalService.
func NewMedicalService(db *gorm.DB, ctx *context.ERPContext) *MedicalService {
	service := MedicalService{
		db:                    db,
		ctx:                   ctx,
		PatientService:        patient.NewPatientService(db, ctx),
		HealthFacilityService: healh_facility.NewHeathFacilityService(db, ctx),
		MedicalStaffService:   medical_staff.NewMedicalStaffService(db, ctx),
		AppointmentService:    appointment.NewAppointmentService(db, ctx),
		MedicalRecord:         medical_record.NewMedicalRecordService(db, ctx),
		PharmacyService:       pharmacy.NewPharmacyService(db, ctx),
		TreatmentQueue:        treatment_queue.NewTreatmentQueueService(db, ctx),
		DoctorService:         doctor.NewDoctorService(db, ctx),
	}
	service.Migrate()
	return &service
}

// Migrate performs automatic migrations for the medical service models.
//
// This method checks if migrations should be skipped and, if not, it uses
// Gorm's AutoMigrate to ensure that the database schema is up-to-date with
// the models defined in the system. It migrates patient, health facility,
// sub-facility, medical staff, doctor, nurse, medical appointment, medical
// record, medicine, and pharmacy models. It returns an error if the migration
// process fails.

func (s *MedicalService) Migrate() error {
	if s.ctx.SkipMigration {
		return nil
	}
	s.db.AutoMigrate(
		&models.PatientModel{},
		&models.HealthFacilityModel{},
		&models.SubFacilityModel{},
		&models.MedicalStaffModel{},
		&models.Doctor{},
		&models.DoctorSchedule{},
		&models.NurseModel{},
		&models.MedicalAppointmentModel{},
		&models.MedicalRecordModel{},
		&models.MedicineModel{},
		&models.PharmacyModel{},
		&models.Prescription{},
		&models.MedicationDetail{},
		&models.Consultation{},
		&models.ConsultationPayment{},
		&models.InitialScreening{},
	)
	return nil
}
