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

func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&models.EmployeeBusinessTrip{},
		// &models.BusinessTripUsage{},
		// &models.BusinessTripRefund{},
	)
}

func (e *EmployeeBusinessTripService) CreateEmployeeBusinessTrip(employeeBusinessTrip *models.EmployeeBusinessTrip) error {
	return e.db.Create(employeeBusinessTrip).Error
}

func (e *EmployeeBusinessTripService) GetEmployeeBusinessTripByID(id string) (*models.EmployeeBusinessTrip, error) {
	var employeeBusinessTrip models.EmployeeBusinessTrip
	err := e.db.
		Preload("Employee").
		Preload("Company").
		Preload("BusinessTripUsages").
		Preload("Refunds").
		Preload("Approver.User").
		Where("id = ?", id).First(&employeeBusinessTrip).Error
	if err != nil {
		return nil, err
	}

	hotelBookings := []models.FileModel{}
	e.db.Find(&hotelBookings, "ref_id = ? AND ref_type = ?", employeeBusinessTrip.ID, "hotel_booking_receipt")
	employeeBusinessTrip.HotelBookingFiles = hotelBookings

	transportReceipt := []models.FileModel{}
	e.db.Find(&transportReceipt, "ref_id = ? AND ref_type = ?", employeeBusinessTrip.ID, "transport_receipt")
	employeeBusinessTrip.HotelBookingFiles = transportReceipt

	return &employeeBusinessTrip, nil
}

func (e *EmployeeBusinessTripService) UpdateEmployeeBusinessTrip(employeeBusinessTrip *models.EmployeeBusinessTrip) error {
	return e.db.Save(employeeBusinessTrip).Error
}

func (e *EmployeeBusinessTripService) DeleteEmployeeBusinessTrip(id string) error {
	return e.db.Delete(&models.EmployeeBusinessTrip{}, "id = ?", id).Error
}

func (e *EmployeeBusinessTripService) FindAllByEmployeeID(request *http.Request, employeeID string) (paginate.Page, error) {
	pg := paginate.New()
	stmt := e.db.
		Preload("Employee").
		Preload("Company").
		Preload("Approver.User").
		Where("employee_id = ?", employeeID).
		Model(&models.EmployeeBusinessTrip{})
	if request.Header.Get("ID-Company") != "" {
		stmt = stmt.Where("company_id = ?", request.Header.Get("ID-Company"))
	}

	if request.URL.Query().Get("search") != "" {
		stmt = stmt.Where("reason LIKE ?", "%"+request.URL.Query().Get("search")+"%")
	}
	if request.URL.Query().Get("start_date") != "" && request.URL.Query().Get("end_date") != "" {
		stmt = stmt.Where("date_requested >= ? AND date_requested <= ?", request.URL.Query().Get("start_date"), request.URL.Query().Get("end_date"))
	} else if request.URL.Query().Get("start_date") != "" {
		stmt = stmt.Where("date_requested = ?", request.URL.Query().Get("start_date"))
	}
	if request.URL.Query().Get("date") != "" {
		stmt = stmt.Where("DATE(date_requested) = ?", request.URL.Query().Get("date"))
	}
	if request.URL.Query().Get("approver_id") != "" {
		stmt = stmt.Where("approver_id = ?", request.URL.Query().Get("approver_id"))
	}

	if request.URL.Query().Get("order") != "" {
		stmt = stmt.Order(request.URL.Query().Get("order"))
	} else {
		stmt = stmt.Order("date_requested DESC")
	}
	utils.FixRequest(request)
	page := pg.With(stmt).Request(request).Response(&[]models.EmployeeBusinessTrip{})
	page.Page = page.Page + 1
	return page, nil
}
func (e *EmployeeBusinessTripService) FindAllEmployeeBusinessTrips(request *http.Request) (paginate.Page, error) {
	pg := paginate.New()
	stmt := e.db.
		Preload("Employee").
		Preload("Company").
		Preload("Approver").
		Preload("Reviewer").
		Model(&models.EmployeeBusinessTrip{})
	if request.Header.Get("ID-Company") != "" {
		stmt = stmt.Where("company_id = ?", request.Header.Get("ID-Company"))
	}

	if request.URL.Query().Get("search") != "" {
		stmt = stmt.Where("reason LIKE ?", "%"+request.URL.Query().Get("search")+"%")
	}
	if request.URL.Query().Get("start_date") != "" && request.URL.Query().Get("end_date") != "" {
		stmt = stmt.Where("date_requested >= ? AND date_requested <= ?", request.URL.Query().Get("start_date"), request.URL.Query().Get("end_date"))
	} else if request.URL.Query().Get("start_date") != "" {
		stmt = stmt.Where("date_requested = ?", request.URL.Query().Get("start_date"))
	}
	if request.URL.Query().Get("date") != "" {
		stmt = stmt.Where("DATE(date_requested) = ?", request.URL.Query().Get("date"))
	}
	if request.URL.Query().Get("approver_id") != "" {
		stmt = stmt.Where("approver_id = ?", request.URL.Query().Get("approver_id"))
	}

	if request.URL.Query().Get("order") != "" {
		stmt = stmt.Order(request.URL.Query().Get("order"))
	} else {
		stmt = stmt.Order("date_requested DESC")
	}
	utils.FixRequest(request)
	page := pg.With(stmt).Request(request).Response(&[]models.EmployeeBusinessTrip{})
	page.Page = page.Page + 1
	return page, nil
}

func (e *EmployeeBusinessTripService) CountByEmployeeID(employeeID string, startDate *time.Time, endDate *time.Time) (map[string]int64, error) {
	var countREQUESTED, countAPPROVED, countREJECTED int64
	counts := make(map[string]int64)
	e.db.Model(&models.EmployeeBusinessTrip{}).
		Where("employee_id = ? AND status = ? AND date_requested >= ? AND date_requested <= ?", employeeID, "REQUESTED", startDate, endDate).
		Count(&countREQUESTED)
	e.db.Model(&models.EmployeeBusinessTrip{}).
		Where("employee_id = ? AND status = ? AND date_requested >= ? AND date_requested <= ?", employeeID, "APPROVED", startDate, endDate).
		Count(&countAPPROVED)
	e.db.Model(&models.EmployeeBusinessTrip{}).
		Where("employee_id = ? AND status = ? AND date_requested >= ? AND date_requested <= ?", employeeID, "REJECTED", startDate, endDate).
		Count(&countREJECTED)

	counts["REQUESTED"] = countREQUESTED
	counts["APPROVED"] = countAPPROVED
	counts["REJECTED"] = countREJECTED

	return counts, nil
}
