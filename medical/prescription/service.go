package prescription

import (
	"net/http"

	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/shared/models"
	"github.com/AMETORY/ametory-erp-modules/utils"
	"github.com/morkid/paginate"
	"gorm.io/gorm"
)

type PrescriptionService struct {
	db  *gorm.DB
	ctx *context.ERPContext
}

// NewPrescriptionService creates a new instance of PrescriptionService with the given database connection and context.
// // It initializes the service to handle operations related to prescriptions.
func NewPrescriptionService(db *gorm.DB, ctx *context.ERPContext) *PrescriptionService {
	return &PrescriptionService{db: db, ctx: ctx}
}

func (s *PrescriptionService) CreatePrescription(prescription *models.Prescription) error {
	return s.db.Create(prescription).Error
}

func (s *PrescriptionService) GetPrescription(id string) (*models.Prescription, error) {
	var prescription models.Prescription
	err := s.db.Where("id = ?", id).First(&prescription).Error
	return &prescription, err
}

func (s *PrescriptionService) GetPrescriptions(request http.Request, search string) (paginate.Page, error) {
	pg := paginate.New()
	stmt := s.db
	if search != "" {
		stmt = stmt.
			Joins("LEFT JOIN patients ON patients.id = prescriptions.patient_id").
			Joins("LEFT JOIN doctors ON doctors.id = prescriptions.doctor_id").
			Where("patients.name ILIKE ? OR doctors.name ILIKE ? OR str_number ILIKE ? OR sip_number ILIKE ?",
				"%"+search+"%",
				"%"+search+"%",
				"%"+search+"%",
				"%"+search+"%",
			)
	}

	request.URL.Query().Get("page")
	stmt = stmt.Model(&models.Prescription{})
	utils.FixRequest(&request)
	page := pg.With(stmt).Request(request).Response(&[]models.Prescription{})
	page.Page = page.Page + 1
	return page, nil
}

func (s *PrescriptionService) UpdatePrescription(prescription *models.Prescription) error {
	return s.db.Save(prescription).Error
}

func (s *PrescriptionService) DeletePrescription(id string) error {
	return s.db.Where("id = ?", id).Delete(&models.Prescription{}).Error
}

func (s *PrescriptionService) AddMedicationDetail(prescriptionID string, medicationDetail *models.MedicationDetail) error {
	medicationDetail.PrescriptionID = prescriptionID
	return s.db.Create(medicationDetail).Error
}

func (s *PrescriptionService) GetMedicationDetails(prescriptionID string) ([]models.MedicationDetail, error) {
	var medicationDetails []models.MedicationDetail
	err := s.db.Where("prescription_id = ?", prescriptionID).Find(&medicationDetails).Error
	return medicationDetails, err
}

func (s *PrescriptionService) UpdateMedicationDetail(medicationDetail *models.MedicationDetail) error {
	return s.db.Save(medicationDetail).Error
}

func (s *PrescriptionService) DeleteMedicationDetail(id string) error {
	return s.db.Where("id = ?", id).Delete(&models.MedicationDetail{}).Error
}
