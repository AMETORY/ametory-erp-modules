package patient

import (
	"net/http"

	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/shared/models"
	"github.com/AMETORY/ametory-erp-modules/utils"
	"github.com/morkid/paginate"
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

// GetPatientByPhoneNumber retrieves a patient by their phone number from the database.
//
// It takes a string argument representing the patient's phone number and returns a pointer
// to a PatientModel and an error. If the retrieval fails, it returns an error.
func (s *PatientService) GetPatientByPhoneNumber(phoneNumber string) (*models.PatientModel, error) {
	p := &models.PatientModel{}
	err := s.db.Where("phone_number = ?", phoneNumber).First(p).Error
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

// GetPatients retrieves a paginated list of patients from the database.
//
// It accepts an HTTP request and a search query string as inputs. The function
// uses GORM to query the database for patients, applying the search query to
// relevant fields in the doctors and doctor_specializations tables. The function
// utilizes pagination to manage the result set and applies any necessary request
// modifications using the utils.FixRequest utility.
//
// The function returns a paginated page of PatientModel models and an error if the
// operation fails.

func (s *PatientService) GetPatients(request http.Request, search string) (paginate.Page, error) {
	pg := paginate.New()
	stmt := s.db
	if search != "" {
		stmt = stmt.Where("full_name ILIKE ? OR identity_card_number ILIKE ? OR social_security_number ILIKE ? OR address ILIKE ?",
			"%"+search+"%",
			"%"+search+"%",
			"%"+search+"%",
			"%"+search+"%",
		)
	}

	order := request.URL.Query().Get("order")
	if order != "" {
		stmt = stmt.Order(order)
	} else {
		stmt = stmt.Order("full_name ASC")
	}
	stmt = stmt.Model(&models.PatientModel{})
	utils.FixRequest(&request)

	page := pg.With(stmt).Request(request).Response(&[]models.PatientModel{})
	page.Page = page.Page + 1
	return page, nil
}
