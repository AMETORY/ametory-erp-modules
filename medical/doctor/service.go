package doctor

import (
	"net/http"

	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/shared/models"
	"github.com/AMETORY/ametory-erp-modules/utils"
	"github.com/morkid/paginate"
	"gorm.io/gorm"
)

type DoctorService struct {
	db  *gorm.DB
	ctx *context.ERPContext
}

// NewDoctorService creates a new instance of DoctorService with the specified
// database connection and ERP context. It initializes the service to handle
// operations related to doctors and their schedules.
func NewDoctorService(db *gorm.DB, ctx *context.ERPContext) *DoctorService {
	return &DoctorService{db: db, ctx: ctx}
}

// CreateDoctor creates a new doctor in the database.
//
// It takes a pointer to a models.Doctor struct as an argument and returns
// an error if the doctor could not be created.
func (ds *DoctorService) CreateDoctor(doctor *models.Doctor) error {
	return ds.db.Create(doctor).Error
}

// GetDoctorByID returns a doctor by ID.
//
// It takes a doctor ID as a parameter and returns the doctor associated
// with that ID. If the doctor does not exist, it returns an error.
func (ds *DoctorService) GetDoctorByID(id string) (*models.Doctor, error) {
	var doctor models.Doctor

	err := ds.db.Where("id = ?", id).Find(&doctor).Error
	if err != nil {
		return nil, err
	}

	return &doctor, nil
}

// GetDoctors retrieves a paginated list of doctors from the database.
//
// It takes an HTTP request and a search query string as input. The method
// uses GORM to query the database for doctors, applying the search query
// to the name, STR number, and SIP number fields. The function utilizes
// pagination to manage the result set and applies any necessary request
// modifications using the utils.FixRequest utility.
//
// The function returns a paginated page of DoctorModel and an error if the
// operation fails.
func (s *DoctorService) GetDoctors(request http.Request, search string) (paginate.Page, error) {
	pg := paginate.New()
	stmt := s.db
	if search != "" {
		stmt = stmt.Where("name ILIKE ? OR str_number ILIKE ? OR sip_number ILIKE ?",
			"%"+search+"%",
			"%"+search+"%",
			"%"+search+"%",
		)
	}

	request.URL.Query().Get("page")
	stmt = stmt.Model(&models.Doctor{})
	utils.FixRequest(&request)
	page := pg.With(stmt).Request(request).Response(&[]models.Doctor{})
	page.Page = page.Page + 1
	return page, nil
}

// UpdateDoctor updates the details of an existing doctor in the database.
//
// It takes a string representing the doctor's ID and a pointer to a Doctor model
// containing the updated information as arguments. The function returns an error if
// the update operation fails.
func (ds *DoctorService) UpdateDoctor(id string, doctor *models.Doctor) error {
	return ds.db.Where("id = ?", id).Updates(doctor).Error
}

// DeleteDoctor removes a doctor from the database.
//
// It takes a string representing the doctor's ID as a parameter and returns
// an error if the deletion process fails.
func (ds *DoctorService) DeleteDoctor(id string) error {
	return ds.db.Where("id = ?", id).Delete(&models.Doctor{}).Error
}

// CreateDoctorSchedule adds a new doctor schedule to the database.
//
// It takes a pointer to a DoctorSchedule model as an argument and returns
// an error if the schedule could not be created.
func (ds *DoctorService) CreateDoctorSchedule(schedule *models.DoctorSchedule) error {
	return ds.db.Create(schedule).Error
}

// GetDoctorScheduleByID retrieves a doctor schedule by its ID.
//
// It takes a string argument representing the schedule ID and returns a pointer
// to a DoctorSchedule model and an error. If the retrieval fails, it returns an
// error.
func (ds *DoctorService) GetDoctorScheduleByID(id string) (*models.DoctorSchedule, error) {
	var schedule models.DoctorSchedule

	err := ds.db.Where("id = ?", id).Find(&schedule).Error
	if err != nil {
		return nil, err
	}

	return &schedule, nil
}

// GetDoctorSchedules retrieves a paginated list of doctor schedules from the database.
//
// It takes an HTTP request and a search query string as input. The method uses GORM
// to query the database for doctor schedules, applying the search query to the doctor ID,
// start time, end time, and price fields. The function utilizes pagination to manage the
// result set and applies any necessary request modifications using the utils.FixRequest utility.
//
// The function returns a paginated page of DoctorSchedule models and an error if the
// operation fails.
func (ds *DoctorService) GetDoctorSchedules(request http.Request, search string) (paginate.Page, error) {
	pg := paginate.New()
	stmt := ds.db
	if search != "" {
		stmt = stmt.Where("doctor_id ILIKE ? OR start_time ILIKE ? OR end_time ILIKE ? OR price ILIKE ?",
			"%"+search+"%",
			"%"+search+"%",
			"%"+search+"%",
			"%"+search+"%",
		)
	}

	request.URL.Query().Get("page")
	stmt = stmt.Model(&models.DoctorSchedule{})
	utils.FixRequest(&request)
	page := pg.With(stmt).Request(request).Response(&[]models.DoctorSchedule{})
	page.Page = page.Page + 1
	return page, nil
}

// UpdateDoctorSchedule updates an existing doctor schedule in the database.
//
// It takes a string representing the schedule's ID and a pointer to a DoctorSchedule
// model containing the updated information as arguments. The function returns an
// error if the update operation fails.
func (ds *DoctorService) UpdateDoctorSchedule(id string, schedule *models.DoctorSchedule) error {
	return ds.db.Where("id = ?", id).Updates(schedule).Error
}

// DeleteDoctorSchedule removes a doctor schedule from the database.
//
// It takes a string representing the doctor schedule's ID as a parameter and returns
// an error if the deletion process fails.
func (ds *DoctorService) DeleteDoctorSchedule(id string) error {
	return ds.db.Where("id = ?", id).Delete(&models.DoctorSchedule{}).Error
}
