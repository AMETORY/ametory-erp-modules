package employee_business_trip

import (
	"net/http"
	"time"

	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/shared/models"
	"github.com/AMETORY/ametory-erp-modules/utils"
	"github.com/morkid/paginate"
	"gorm.io/gorm"
)

type EmployeeBusinessTripService struct {
	db  *gorm.DB
	ctx *context.ERPContext
}

func NewEmployeeBusinessTripService(ctx *context.ERPContext) *EmployeeBusinessTripService {
	return &EmployeeBusinessTripService{db: ctx.DB, ctx: ctx}
}

// Migrate applies the database schema changes needed for the EmployeeBusinessTrip model.
// It uses GORM's AutoMigrate function to ensure the database table structure
// matches the defined model. This includes creating or updating tables as needed.
func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&models.EmployeeBusinessTrip{},
		// &models.BusinessTripUsage{},
		// &models.BusinessTripRefund{},
	)
}

// CreateEmployeeBusinessTrip creates a new EmployeeBusinessTrip in the database.
// It takes a pointer to an EmployeeBusinessTrip as input and creates a new record in the database.
// The function returns an error if the operation fails.
func (e *EmployeeBusinessTripService) CreateEmployeeBusinessTrip(employeeBusinessTrip *models.EmployeeBusinessTrip) error {
	return e.db.Create(employeeBusinessTrip).Error
}

// GetEmployeeBusinessTripByID retrieves an employee business trip by ID from the database.
//
// The employee business trip is queried using the GORM First method, and any errors are
// returned to the caller. If the employee business trip is not found, a nil pointer is
// returned together with a gorm.ErrRecordNotFound error.
//
// The function also retrieves all relevant related data, such as the employee, approver,
// company, trip participants, transport booking files, and hotel booking files.
func (e *EmployeeBusinessTripService) GetEmployeeBusinessTripByID(id string) (*models.EmployeeBusinessTrip, error) {
	var employeeBusinessTrip models.EmployeeBusinessTrip
	err := e.db.
		Preload("Employee", func(tx *gorm.DB) *gorm.DB {
			return tx.Preload("User").
				Preload("JobTitle").
				Preload("WorkLocation").
				Preload("WorkShift").
				Preload("Branch")
		}).
		Preload("Approver.User").
		Preload("ApprovalByAdmin").
		Preload("Company").
		Preload("TripParticipants", func(db *gorm.DB) *gorm.DB {
			return db.Preload("User").Preload("JobTitle")
		}).
		Preload("Approver.User").
		Where("id = ?", id).First(&employeeBusinessTrip).Error
	if err != nil {
		return nil, err
	}

	ticketFiles := []models.FileModel{}
	e.ctx.DB.Find(&ticketFiles, "ref_id = ? AND ref_type = ?", id, "employee_business_trip_transport_ticket")

	employeeBusinessTrip.TransportBookingFiles = ticketFiles

	hotelFiles := []models.FileModel{}
	e.ctx.DB.Find(&hotelFiles, "ref_id = ? AND ref_type = ?", id, "employee_business_trip_hotel_ticket")

	employeeBusinessTrip.HotelBookingFiles = hotelFiles

	return &employeeBusinessTrip, nil
}

// UpdateEmployeeBusinessTrip updates an existing employee business trip record in the database.
//
// It takes an EmployeeBusinessTrip pointer as input, and returns an error if the
// operation fails. The function uses GORM to update the employee business trip data in
// the employee_business_trips table where the ID matches.
//
// If the update is successful, the error is nil. Otherwise, the error contains
// information about what went wrong.
func (e *EmployeeBusinessTripService) UpdateEmployeeBusinessTrip(employeeBusinessTrip *models.EmployeeBusinessTrip) error {
	return e.db.Save(employeeBusinessTrip).Error
}

// DeleteEmployeeBusinessTrip deletes an employee business trip record from the database by ID.
//
// It takes an ID as input and returns an error if the deletion operation fails.
// The function uses GORM to delete the employee business trip data from the employee_business_trips table.
// If the deletion is successful, the error is nil. Otherwise, the error contains
// information about what went wrong.
func (e *EmployeeBusinessTripService) DeleteEmployeeBusinessTrip(id string) error {
	return e.db.Delete(&models.EmployeeBusinessTrip{}, "id = ?", id).Error
}

// FindAllByEmployeeID retrieves a paginated list of employee business trips.
//
// The method uses GORM to query the database for employee business trips, preloading the
// associated Company, Employee, Approver, and ApprovalByAdmin models. It applies a filter based on
// the company ID provided in the HTTP request header, and another filter based
// on the search parameter if provided. The function utilizes pagination to
// manage the result set and applies any necessary request modifications using
// the utils.FixRequest utility.
//
// The function returns a paginated page of EmployeeBusinessTrip and an error if the
// operation fails.
func (e *EmployeeBusinessTripService) FindAllByEmployeeID(request *http.Request, employeeID string) (paginate.Page, error) {
	pg := paginate.New()
	stmt := e.db.
		Preload("Company").
		Preload("Employee", func(tx *gorm.DB) *gorm.DB {
			return tx.Preload("User").
				Preload("JobTitle").
				Preload("WorkLocation").
				Preload("WorkShift").
				Preload("Branch")
		}).
		Preload("Approver.User").
		Preload("ApprovalByAdmin").
		Where("employee_id = ?", employeeID).
		Model(&models.EmployeeBusinessTrip{})
	if request.Header.Get("ID-Company") != "" {
		stmt = stmt.Where("company_id = ?", request.Header.Get("ID-Company"))
	}

	if request.URL.Query().Get("search") != "" {
		stmt = stmt.Where("reason LIKE ?", "%"+request.URL.Query().Get("search")+"%")
	}
	if request.URL.Query().Get("start_date") != "" && request.URL.Query().Get("end_date") != "" {
		stmt = stmt.Where("date >= ? AND date <= ?", request.URL.Query().Get("start_date"), request.URL.Query().Get("end_date"))
	} else if request.URL.Query().Get("start_date") != "" {
		stmt = stmt.Where("date = ?", request.URL.Query().Get("start_date"))
	}
	if request.URL.Query().Get("date") != "" {
		stmt = stmt.Where("DATE(date) = ?", request.URL.Query().Get("date"))
	}
	if request.URL.Query().Get("approver_id") != "" {
		stmt = stmt.Where("approver_id = ?", request.URL.Query().Get("approver_id"))
	}

	if request.URL.Query().Get("order") != "" {
		stmt = stmt.Order(request.URL.Query().Get("order"))
	} else {
		stmt = stmt.Order("date DESC")
	}
	utils.FixRequest(request)
	page := pg.With(stmt).Request(request).Response(&[]models.EmployeeBusinessTrip{})
	page.Page = page.Page + 1
	return page, nil
}

// FindAllEmployeeBusinessTrips retrieves a paginated list of employee business trips.
//
// The method uses GORM to query the database for employee business trips, preloading the
// associated Company, Employee, Approver, ApprovalByAdmin, and Reviewer models. It applies a
// filter based on the company ID if provided in the HTTP request headers, and a filter based on
// the start and end date if provided in the HTTP request query. Additionally, it applies a filter
// based on the search query if provided, and a filter based on the approver ID if provided. The
// function utilizes pagination to manage the result set and applies any necessary request
// modifications using the utils.FixRequest utility.
//
// The function returns a paginated page of EmployeeBusinessTripModel and an error if the
// operation fails.
func (e *EmployeeBusinessTripService) FindAllEmployeeBusinessTrips(request *http.Request) (paginate.Page, error) {
	pg := paginate.New()
	stmt := e.db.
		Preload("Company").
		Preload("Employee", func(tx *gorm.DB) *gorm.DB {
			return tx.Preload("User").
				Preload("JobTitle").
				Preload("WorkLocation").
				Preload("WorkShift").
				Preload("Branch")
		}).
		Preload("Approver.User").
		Preload("ApprovalByAdmin").
		Preload("Reviewer").
		Model(&models.EmployeeBusinessTrip{})
	if request.Header.Get("ID-Company") != "" {
		stmt = stmt.Where("company_id = ?", request.Header.Get("ID-Company"))
	}

	if request.URL.Query().Get("search") != "" {
		stmt = stmt.Where("reason LIKE ?", "%"+request.URL.Query().Get("search")+"%")
	}
	if request.URL.Query().Get("start_date") != "" && request.URL.Query().Get("end_date") != "" {
		stmt = stmt.Where("date >= ? AND date <= ?", request.URL.Query().Get("start_date"), request.URL.Query().Get("end_date"))
	} else if request.URL.Query().Get("start_date") != "" {
		stmt = stmt.Where("date = ?", request.URL.Query().Get("start_date"))
	}
	if request.URL.Query().Get("date") != "" {
		stmt = stmt.Where("DATE(date) = ?", request.URL.Query().Get("date"))
	}
	if request.URL.Query().Get("approver_id") != "" {
		stmt = stmt.Where("approver_id = ?", request.URL.Query().Get("approver_id"))
	}

	if request.URL.Query().Get("order") != "" {
		stmt = stmt.Order(request.URL.Query().Get("order"))
	} else {
		stmt = stmt.Order("date DESC")
	}
	utils.FixRequest(request)
	page := pg.With(stmt).Request(request).Response(&[]models.EmployeeBusinessTrip{})
	page.Page = page.Page + 1
	return page, nil
}

// CountByEmployeeID retrieves the count of employee business trips by status for a given employee
// within a specified date range.
//
// The function returns a map where the keys are the statuses "REQUESTED", "APPROVED", and "REJECTED",
// and the values are the counts of trips for each status. If the operation is successful, the error
// is nil; otherwise, it returns an error indicating what went wrong.
//
// Parameters:
//   - employeeID: The ID of the employee whose business trips are being counted.
//   - startDate: The start date of the date range for filtering business trips.
//   - endDate: The end date of the date range for filtering business trips.
//
// Returns:
//   - A map of trip counts by status.
//   - An error if the operation fails.

func (e *EmployeeBusinessTripService) CountByEmployeeID(employeeID string, startDate *time.Time, endDate *time.Time) (map[string]int64, error) {
	var countREQUESTED, countAPPROVED, countREJECTED int64
	counts := make(map[string]int64)
	e.db.Model(&models.EmployeeBusinessTrip{}).
		Where("employee_id = ? AND status = ? AND date >= ? AND date <= ?", employeeID, "REQUESTED", startDate, endDate).
		Count(&countREQUESTED)
	e.db.Model(&models.EmployeeBusinessTrip{}).
		Where("employee_id = ? AND status = ? AND date >= ? AND date <= ?", employeeID, "APPROVED", startDate, endDate).
		Count(&countAPPROVED)
	e.db.Model(&models.EmployeeBusinessTrip{}).
		Where("employee_id = ? AND status = ? AND date >= ? AND date <= ?", employeeID, "REJECTED", startDate, endDate).
		Count(&countREJECTED)

	counts["REQUESTED"] = countREQUESTED
	counts["APPROVED"] = countAPPROVED
	counts["REJECTED"] = countREJECTED

	return counts, nil
}
