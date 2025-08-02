package consultation

import (
	"net/http"

	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/shared/models"
	"github.com/AMETORY/ametory-erp-modules/utils"
	"github.com/morkid/paginate"
	"gorm.io/gorm"
)

type ConsultationService struct {
	db  *gorm.DB
	ctx *context.ERPContext
}

// NewConsultationService creates a new instance of ConsultationService with the given
// database connection and ERP context. It initializes the service to handle
// operations related to consultations.
func NewConsultationService(db *gorm.DB, ctx *context.ERPContext) *ConsultationService {
	return &ConsultationService{db: db, ctx: ctx}
}

// CreateConsultation adds a new consultation record to the database.
// It takes a pointer to a Consultation model as an argument and returns
// an error if the creation process fails.
func (s *ConsultationService) CreateConsultation(consultation *models.Consultation) error {
	return s.db.Create(consultation).Error
}

// GetConsultation retrieves a consultation record with the given ID.
// It takes a string representing the consultation ID as an argument and
// returns a pointer to a Consultation model and an error if the retrieval
// process fails.
func (s *ConsultationService) GetConsultation(id string) (*models.Consultation, error) {
	consultation := models.Consultation{}
	err := s.db.Where("id = ?", id).First(&consultation).Error
	return &consultation, err
}

// UpdateConsultation updates a consultation record with the given
// Consultation model.
//
// It takes a pointer to a Consultation model as an argument and returns
// an error if the update process fails.
func (s *ConsultationService) UpdateConsultation(consultation *models.Consultation) error {
	return s.db.Save(consultation).Error
}

// DeleteConsultation removes a consultation record from the database.
//
// It takes a string representing the consultation ID as an argument and
// returns an error if the deletion process fails.
func (s *ConsultationService) DeleteConsultation(id string) error {
	return s.db.Delete(&models.Consultation{}, "id = ?", id).Error
}

// GetConsultations retrieves a list of consultations with pagination.
//
// It takes a pointer to http.Request and an optional doctor ID as arguments.
// If the doctor ID is not nil, it will filter the consultations by the doctor ID.
// If the search query parameter is not empty, it will search for consultations by
// doctor name or patient name.
//
// It returns a paginate.Page and an error if the retrieval process fails.
func (s *ConsultationService) GetConsultations(request http.Request, doctorID *string) (paginate.Page, error) {
	pg := paginate.New()
	stmt := s.db
	if doctorID != nil {
		stmt = stmt.Where("doctor_id = ?", *doctorID)
	}
	if request.URL.Query().Get("search") != "" {
		search := request.URL.Query().Get("search")
		stmt = stmt.
			Joins("LEFT JOIN patients ON patients.id = consultations.patient_id").
			Joins("LEFT JOIN doctors ON doctors.id = consultations.doctor_id").
			Where("patients.name ILIKE ? OR doctors.name ILIKE ? ",
				"%"+search+"%",
				"%"+search+"%",
			)
	}

	request.URL.Query().Get("page")
	stmt = stmt.Model(&models.Consultation{})
	utils.FixRequest(&request)
	page := pg.With(stmt).Request(request).Response(&[]models.Consultation{})
	page.Page = page.Page + 1
	return page, nil
}

// CreatePayment adds a new consultation payment record to the database.
//
// It takes a pointer to a ConsultationPayment model as an argument
// and returns an error if the creation process fails.
func (s *ConsultationService) CreatePayment(payment *models.ConsultationPayment) error {
	return s.db.Create(payment).Error
}

// DeletePayment deletes a payment transaction associated with a consultation record.
//
// The function takes the ID of the transaction to be deleted as input and
// attempts to delete the corresponding record in the database. It returns an
// error if the deletion fails.
func (s *ConsultationService) DeletePayment(paymentID string) error {
	return s.db.Delete(&models.ConsultationPayment{}, "id = ?", paymentID).Error
}

func (s *ConsultationService) CreateInitialScreening(screen *models.InitialScreening) error {
	return s.db.Create(screen).Error
}

// GetInitialScreening retrieves an initial screening record with the given ID.
//
// It takes a string representing the screening ID as an argument and
// returns a pointer to a InitialScreening model and an error if the retrieval
// process fails.
func (s *ConsultationService) GetInitialScreening(id string) (*models.InitialScreening, error) {
	screen := models.InitialScreening{}
	err := s.db.Where("id = ?", id).First(&screen).Error
	return &screen, err
}

// UpdateInitialScreening updates an existing initial screening record in the database.
//
// It takes a pointer to an InitialScreening model as an argument and returns
// an error if the update process fails.
func (s *ConsultationService) UpdateInitialScreening(screen *models.InitialScreening) error {
	return s.db.Save(screen).Error
}

// DeleteInitialScreening deletes an initial screening record from the database.
//
// It takes a string argument representing the screening ID and returns an error
// if the deletion process fails.
func (s *ConsultationService) DeleteInitialScreening(id string) error {
	return s.db.Delete(&models.InitialScreening{}, "id = ?", id).Error
}
