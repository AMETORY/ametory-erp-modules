package doctor

import (
	"net/http"
	"time"

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

	err := ds.db.
		Preload("Specialization").
		Where("id = ?", id).Find(&doctor).Error
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
	stmt = stmt.Preload("Specialization")
	if search != "" {
		stmt = stmt.Joins("LEFT JOIN doctor_specializations ON doctor_specializations.code = doctors.specialization_code")
		stmt = stmt.Where("doctors.name ILIKE ? OR doctors.str_number ILIKE ? OR doctors.s_ip_number ILIKE ? OR doctor_specializations.name ILIKE ? OR doctor_specializations.description ILIKE ?",
			"%"+search+"%",
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
		stmt = stmt.Order("name ASC")
	}
	stmt = stmt.Model(&models.Doctor{})
	utils.FixRequest(&request)

	page := pg.With(stmt).Request(request).Response(&[]models.Doctor{})
	page.Page = page.Page + 1
	return page, nil
}

// GetDoctorByName&SpecializationCode returns a doctor by name and specialization code.
//
// It takes a name and specialization code as parameters and returns the doctor
// associated with that name and specialization code. If the doctor does not exist,
// it returns an error.
func (ds *DoctorService) GetDoctorByNameAndSpecializationCode(name *string, specializationCode *string) ([]models.Doctor, error) {
	var doctors []models.Doctor
	stmt := ds.db.Preload("Specialization")
	if specializationCode != nil {
		stmt = stmt.Joins("JOIN doctor_specializations ON doctor_specializations.code = doctors.specialization_code").
			Where("doctor_specializations.code = ?", specializationCode)
	}
	if name != nil {
		stmt = stmt.Where("doctors.name ilike ?", "%"+*name+"%")
	}
	err := stmt.Debug().Find(&doctors).Error
	if err != nil {
		return nil, err
	}
	return doctors, nil
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
	return ds.db.Debug().Create(schedule).Error
}

// GetDoctorScheduleByID retrieves a doctor schedule by its ID.
//
// It takes a string argument representing the schedule ID and returns a pointer
// to a DoctorSchedule model and an error. If the retrieval fails, it returns an
// error.
func (ds *DoctorService) GetDoctorScheduleByID(id string) (*models.DoctorSchedule, error) {
	var schedule models.DoctorSchedule

	err := ds.db.Where("id = ?", id).
		Preload("Patient").Preload("Doctor.Specialization").
		Find(&schedule).Error
	if err != nil {
		return nil, err
	}

	return &schedule, nil
}
func (ds *DoctorService) GetDoctorScheduleByDate(doctorId string, date *string, timeStr *string) (*models.DoctorSchedule, error) {
	var schedule models.DoctorSchedule
	stmt := ds.db.Preload("Doctor.Specialization").Where("doctor_id = ?", doctorId)

	if date != nil && timeStr != nil {
		availableTime, err := time.ParseInLocation("2006-01-02 15:04", *date+" "+*timeStr, time.Local)
		if err == nil {
			stmt = stmt.Where("start_time = ?", availableTime)
		}
	}
	err := stmt.Debug().First(&schedule).Error
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
func (ds *DoctorService) GetDoctorSchedules(request http.Request, search string, doctorID *string) (paginate.Page, error) {
	pg := paginate.New()
	stmt := ds.db.Preload("Patient")
	if search != "" {
		stmt = stmt.Joins("LEFT JOIN doctors ON doctor_schedules.doctor_id = doctors.id").
			Where("doctors.name ILIKE ?",
				"%"+search+"%",
			)
	}

	order := request.URL.Query().Get("order")
	if order != "" {
		stmt = stmt.Order(order)
	} else {
		stmt = stmt.Order("created_at ASC")
	}

	if request.URL.Query().Get("doctor_id") != "" {
		stmt = stmt.Where("doctor_id = ?", request.URL.Query().Get("doctor_id"))
	}

	if doctorID != nil {
		stmt = stmt.Where("doctor_id = ?", *doctorID)
	}

	if request.URL.Query().Get("time") != "" {
		stmt = stmt.Where("start_time <= ? AND end_time >= ?", request.URL.Query().Get("time"), request.URL.Query().Get("time"))
	}

	if request.URL.Query().Get("date") != "" {
		stmt = stmt.Where("start_time <= ? AND end_time >= ?", request.URL.Query().Get("date"), request.URL.Query().Get("date"))
	}
	if request.URL.Query().Get("month") != "" {
		thisMonth, err := time.Parse("2006-01-02", request.URL.Query().Get("month"))
		if err == nil {
			firstDateOfMonth := time.Date(thisMonth.Year(), thisMonth.Month(), 1, 0, 0, 0, 0, time.UTC)
			lastDateOfMonth := firstDateOfMonth.AddDate(0, 1, -1)
			stmt = stmt.Where("start_time >= ? AND end_time <= ?", firstDateOfMonth, lastDateOfMonth)

		}

	}

	if request.URL.Query().Get("status") != "" {
		stmt = stmt.Where("status = ?", request.URL.Query().Get("status"))
	}
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

// CreateDoctorSpecialization adds a new doctor specialization to the database.
//
// It takes a pointer to a DoctorSpecialization model as an argument and returns
// an error if the specialization could not be created.
func (ds *DoctorService) CreateDoctorSpecialization(specialization *models.DoctorSpecialization) error {
	return ds.db.Create(specialization).Error
}

// GetDoctorSpecializations retrieves a paginated list of doctor specializations from the database.
//
// It takes a pointer to an HTTP request and a search query string as arguments. The method
// uses GORM to query the database for doctor specializations, applying the search query
// to the code, name, and description fields. The function utilizes pagination to manage
// the result set and applies any necessary request modifications using the utils.FixRequest
// utility.
//
// The function returns a paginated page of DoctorSpecialization models and an error if the
// operation fails.
func (ds *DoctorService) GetDoctorSpecializations(request http.Request, search string) (paginate.Page, error) {
	pg := paginate.New()
	stmt := ds.db
	if search != "" {
		stmt = stmt.Where("code ILIKE ? OR name ILIKE ? OR description ILIKE ?",
			"%"+search+"%",
			"%"+search+"%",
			"%"+search+"%",
		)
	}

	order := request.URL.Query().Get("order")
	if order != "" {
		stmt = stmt.Order(order)
	} else {
		stmt = stmt.Order("name ASC")
	}
	stmt = stmt.Model(&models.DoctorSpecialization{})
	utils.FixRequest(&request)
	page := pg.With(stmt).Request(request).Response(&[]models.DoctorSpecialization{})
	page.Page = page.Page + 1
	return page, nil
}

// GetDoctorSpecializationsCode retrieves a list of doctor specialization codes from the database.
//
// It takes no arguments and returns a slice of strings and an error if the operation fails.
func (ds *DoctorService) GetDoctorSpecializationsCode() ([]string, error) {
	var codes []string
	err := ds.db.Model(&models.DoctorSpecialization{}).Select("code").Find(&codes).Error
	if err != nil {
		return nil, err
	}
	return codes, nil
}

// GetDoctorSpecializationByID retrieves a doctor specialization by its ID.
//
// It takes a string argument representing the specialization ID and returns a pointer
// to a DoctorSpecialization model and an error. If the retrieval fails, it returns an
// error.
func (ds *DoctorService) GetDoctorSpecializationByID(id string) (*models.DoctorSpecialization, error) {
	var specialization models.DoctorSpecialization

	err := ds.db.Where("id = ?", id).Find(&specialization).Error
	if err != nil {
		return nil, err
	}

	return &specialization, nil
}

// UpdateDoctorSpecialization updates an existing doctor specialization in the database.
//
// It takes a string representing the specialization's ID and a pointer to a DoctorSpecialization
// model containing the updated information as arguments. The function returns an
// error if the update operation fails.
func (ds *DoctorService) UpdateDoctorSpecialization(id string, specialization *models.DoctorSpecialization) error {
	return ds.db.Where("id = ?", id).Updates(specialization).Error
}

// DeleteDoctorSpecialization removes a doctor specialization from the database.
//
// It takes a string representing the doctor specialization's ID as a parameter and returns
// an error if the deletion process fails.
func (ds *DoctorService) DeleteDoctorSpecialization(id string) error {
	return ds.db.Where("id = ?", id).Delete(&models.DoctorSpecialization{}).Error
}

// FindScheduleFromParams retrieves a doctor schedule by its parameters.
//
// It takes a set of parameters in the form of a gin.Context and returns a pointer
// to a DoctorSchedule model and an error. If the retrieval fails, it returns an
// error.
func (ds *DoctorService) FindScheduleFromParams(doctorID, date, specializationCode, timeStr, doctorName, status *string) []models.DoctorSchedule {
	var schedules []models.DoctorSchedule
	stmt := ds.db.Preload("Doctor.Specialization")
	if doctorID != nil {
		stmt = stmt.Where("doctor_id = ?", *doctorID)
	}
	if date != nil {
		stmt = stmt.Where("DATE(start_time) = ?", *date)
	}
	if specializationCode != nil || doctorName != nil {
		stmt = stmt.Joins("LEFT JOIN doctors ON doctor_schedules.doctor_id = doctors.id")
		if specializationCode != nil {
			stmt = stmt.Where("doctors.specialization_code = ?", *specializationCode)
		}
		if doctorName != nil {
			stmt = stmt.Where("doctors.name ilike ?", "%"+*doctorName+"%")
		}
	}
	if date != nil && timeStr != nil {
		availableTime, err := time.ParseInLocation("2006-01-02 15:04", *date+" "+*timeStr, time.Local)
		if err == nil {
			stmt = stmt.Where("start_time >= ?", availableTime)
		}
	}

	if status != nil {
		stmt = stmt.Where("status = ?", *status)
	}

	stmt.Where("start_time >= ?", time.Now().Format("2006-01-02")).Debug().Find(&schedules)

	return schedules
}

// FindUserSchedule retrieves user schedules based on provided parameters.
//
// It takes a set of parameters in the form of a gin.Context and returns a slice
// of UserSchedule models. If the retrieval fails, it returns an error.
func (ds *DoctorService) FindUserSchedule(patientId, phoneNumber, status, startDate, endDate *string) ([]models.DoctorSchedule, error) {
	var schedules []models.DoctorSchedule
	stmt := ds.db.Preload("Patient").Preload("Doctor.Specialization")
	if patientId != nil {
		stmt = stmt.Where("patient_id = ?", *patientId)
	}

	if phoneNumber != nil {
		stmt = stmt.Joins("LEFT JOIN patients ON doctor_schedules.patient_id = patients.id")
		stmt = stmt.Where("patients.phone_number = ?", *phoneNumber)
	}

	if status != nil {
		stmt = stmt.Where("status = ?", *status)
	}

	if startDate != nil {
		stmt = stmt.Where("start_time >= ?", *startDate)
	} else if endDate != nil {
		stmt = stmt.Where("start_time <= ?", *endDate)
	} else if startDate != nil && endDate != nil {
		stmt = stmt.Where("start_time >= ? AND start_time <= ?", *startDate, *endDate)
	}
	if err := stmt.Find(&schedules).Error; err != nil {
		return nil, err
	}

	return schedules, nil
}
