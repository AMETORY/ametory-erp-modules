package reimbursement

import (
	"errors"
	"net/http"
	"time"

	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/hris/employee"
	"github.com/AMETORY/ametory-erp-modules/shared/models"
	"github.com/AMETORY/ametory-erp-modules/utils"
	"github.com/morkid/paginate"
	"gorm.io/gorm"
)

type ReimbursementService struct {
	db              *gorm.DB
	ctx             *context.ERPContext
	employeeService *employee.EmployeeService
}

// NewReimbursementService creates a new instance of ReimbursementService.
//
// The service is created by providing a GORM database instance, an ERP context, and an EmployeeService instance.
// The ERP context is used for authentication and authorization purposes, while the database instance is used for CRUD (Create, Read, Update, Delete) operations.
// The EmployeeService instance is used to retrieve employee information.
func NewReimbursementService(ctx *context.ERPContext, employeeService *employee.EmployeeService) *ReimbursementService {
	return &ReimbursementService{db: ctx.DB, ctx: ctx, employeeService: employeeService}
}

func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&models.ReimbursementModel{},
		&models.ReimbursementItemModel{},
	)
}

// CreateReimbursement creates a new reimbursement record in the database.
//
// The function takes a ReimbursementModel as input and uses it to create a new
// reimbursement record in the database. Any errors are returned to the caller.
//
// The function requires a non-nil employee id in the input model, and will return
// an error if the employee id is not provided.
func (s *ReimbursementService) CreateReimbursement(m *models.ReimbursementModel) error {
	if m.EmployeeID == nil {
		return errors.New("employee id is required")
	}
	return s.db.Create(m).Error
}

// FindAllReimbursementByEmployeeID retrieves a paginated list of reimbursements for a given employee ID.
//
// The function takes an http.Request as input and applies various filters based on the request
// parameters such as search term, date range, and order. The results can be ordered based on a
// specified order parameter or defaults to ordering by date in descending order.
//
// It utilizes pagination to manage the result set and applies any necessary request modifications
// using the utils.FixRequest utility. The function returns a paginated page of ReimbursementModel
// and an error if the operation fails.
func (s *ReimbursementService) FindAllReimbursementByEmployeeID(request *http.Request, employeeID string) (paginate.Page, error) {
	pg := paginate.New()
	stmt := s.db.Preload("Approver").Where("employee_id = ?", employeeID).Model(&models.ReimbursementModel{})
	if request.URL.Query().Get("search") != "" {
		stmt = stmt.Where("name LIKE ?", "%"+request.URL.Query().Get("search")+"%")
	}
	if request.URL.Query().Get("start_date") != "" && request.URL.Query().Get("end_date") != "" {
		stmt = stmt.Where("date BETWEEN ? AND ?", request.URL.Query().Get("start_date"), request.URL.Query().Get("end_date"))
	}
	if request.URL.Query().Get("order") != "" {
		stmt = stmt.Order(request.URL.Query().Get("order"))
	} else {
		stmt = stmt.Order("date DESC")
	}
	utils.FixRequest(request)
	page := pg.With(stmt).Request(request).Response(&[]models.ReimbursementModel{})
	page.Page = page.Page + 1
	return page, nil
}

// FindAllReimbursement retrieves a paginated list of all reimbursements.
//
// The function uses GORM to query the database for reimbursements, preloading associated
// Employee (with User, JobTitle, WorkLocation, WorkShift, Branch) and Approver (with User) models.
// It applies any necessary request modifications using the utils.FixRequest utility.
// The function returns a paginated page of ReimbursementModel and an error if the operation fails.
func (s *ReimbursementService) FindAllReimbursement(request *http.Request) (paginate.Page, error) {
	pg := paginate.New()
	stmt := s.db.
		Preload("Employee", func(tx *gorm.DB) *gorm.DB {
			return tx.Preload("User").
				Preload("JobTitle").
				Preload("WorkLocation").
				Preload("WorkShift").
				Preload("Branch")
		}).
		Preload("Approver.User").
		Model(&models.ReimbursementModel{})
	utils.FixRequest(request)
	page := pg.With(stmt).Request(request).Response(&[]models.ReimbursementModel{})
	page.Page = page.Page + 1
	return page, nil
}

// FindReimbursementByID retrieves a reimbursement record by its ID from the database.
//
// The function preloads the associated Employee (with User, JobTitle, WorkLocation, WorkShift, Branch),
// Approver (with User), and Items models. It also loads any file attachments for both the
// reimbursement and its items. The function returns the populated ReimbursementModel and an error
// if the operation fails. If the reimbursement record is not found, a nil pointer is returned
// together with a gorm.ErrRecordNotFound error.
func (s *ReimbursementService) FindReimbursementByID(id string) (*models.ReimbursementModel, error) {
	var m models.ReimbursementModel
	if err := s.db.
		Preload("Employee", func(tx *gorm.DB) *gorm.DB {
			return tx.Preload("User").
				Preload("JobTitle").
				Preload("WorkLocation").
				Preload("WorkShift").
				Preload("Branch")
		}).
		Preload("Approver.User").
		Preload("Items").Where("id = ?", id).First(&m).Error; err != nil {
		return nil, err
	}
	for i, v := range m.Items {
		files := []models.FileModel{}
		s.db.Find(&files, "ref_id = ? AND ref_type = ?", v.ID, "reimbursement_item")
		v.Attachments = files
		m.Items[i] = v
	}

	var files []models.FileModel
	s.db.Find(&files, "ref_id = ? AND ref_type = ?", m.ID, "reimbursement")
	m.Files = files

	return &m, nil
}

// UpdateReimbursement updates an existing reimbursement record in the database.
//
// The function takes a pointer to a ReimbursementModel as input and attempts to update
// the corresponding record in the database. If the update fails, it returns an error.
func (s *ReimbursementService) UpdateReimbursement(m *models.ReimbursementModel) error {
	return s.db.Save(m).Error
}

// DeleteReimbursement deletes a reimbursement record from the database by its ID.
//
// The function takes a reimbursement ID as input and attempts to delete the corresponding record
// from the database. If the deletion fails, it returns an error.
func (s *ReimbursementService) DeleteReimbursement(id string) error {
	return s.db.Where("id = ?", id).Delete(&models.ReimbursementModel{}).Error
}

// Delete removes a reimbursement record from the database by its ID.
//
// The function takes a reimbursement ID as input and attempts to delete the
// corresponding record from the database. If the deletion operation fails,
// it returns an error, otherwise it returns nil.
func (s *ReimbursementService) Delete(id string) error {
	return s.db.Where("id = ?", id).Delete(&models.ReimbursementModel{}).Error
}

// CreateReimbursementItem creates a new reimbursement item record in the database.
//
// The function takes a pointer to a ReimbursementItemModel as input and attempts to
// create a new reimbursement item record in the database. If the creation fails,
// it returns an error.
//
// The function requires a non-nil reimbursement id in the input model, and will return
// an error if the reimbursement id is not provided.
func (s *ReimbursementService) CreateReimbursementItem(m *models.ReimbursementItemModel) error {
	if m.ReimbursementID == nil {
		return errors.New("reimbursement id is required")
	}
	return s.db.Create(m).Error
}

// UpdateReimbursementItem updates an existing reimbursement item record in the database.
//
// The function takes an ID and a pointer to a ReimbursementItemModel as input and attempts to
// update the corresponding record in the database. If the update fails, it returns an error.
func (s *ReimbursementService) UpdateReimbursementItem(id string, m *models.ReimbursementItemModel) error {
	return s.db.Where("id = ?", id).Save(m).Error
}

// DeleteReimbursementItem deletes a reimbursement item record from the database by its ID.
//
// The function takes a reimbursement item ID as input and attempts to delete the
// corresponding record from the database. If the deletion operation fails,
// it returns an error, otherwise it returns nil.
func (s *ReimbursementService) DeleteReimbursementItem(id string) error {
	return s.db.Where("id = ?", id).Delete(&models.ReimbursementItemModel{}).Error
}

// CountByStatusAndEmployeeID returns a count of reimbursement records that match the given status, employee id, and date range.
//
// The function takes a status string, an employee id string, and start and end date pointers as input.
// It generates a GORM query that filters reimbursement records by the given status and employee id,
// and if given, the start and end dates. The function then executes the query and returns the count
// of matching records and an error if the operation fails.
func (s *ReimbursementService) CountByStatusAndEmployeeID(status string, employeeID string, startDate *time.Time, endDate *time.Time) (int64, error) {
	var count int64
	stmt := s.db.Model(&models.ReimbursementModel{}).
		Where("employee_id = ?", employeeID).
		Where("status = ?", status)
	if startDate != nil && endDate != nil {
		stmt = stmt.Where("created_at BETWEEN ? AND ?", startDate, endDate)
	}
	err := stmt.
		Count(&count).Error
	if err != nil {
		return 0, err
	}
	return count, nil
}
