package leave

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/hris/employee"
	"github.com/AMETORY/ametory-erp-modules/shared/models"
	"github.com/AMETORY/ametory-erp-modules/utils"
	"github.com/morkid/paginate"
	"gorm.io/gorm"
)

type LeaveService struct {
	db              *gorm.DB
	ctx             *context.ERPContext
	employeeService *employee.EmployeeService
}

// NewLeaveService creates a new instance of LeaveService.
//
// The service is initialized with a GORM database from the ERP context
// and an EmployeeService for handling employee-related operations.

func NewLeaveService(ctx *context.ERPContext, employeeService *employee.EmployeeService) *LeaveService {
	return &LeaveService{db: ctx.DB, ctx: ctx, employeeService: employeeService}
}

// Migrate runs the database migration for the Leave module.
//
// The function takes a GORM database and runs the AutoMigrate method
// on the LeaveModel and LeaveCategory models.
//
// The function returns an error if the migration fails.
func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&models.LeaveModel{},
		&models.LeaveCategory{},
	)
}

// CreateLeave creates a new leave record in the database.
//
// The function takes a LeaveModel as input and creates a new leave record in the database.
// The function returns an error if the creation fails, or if the EmployeeID field of the input
// LeaveModel is nil.
func (s *LeaveService) CreateLeave(m *models.LeaveModel) error {
	if m.EmployeeID == nil {
		return errors.New("employee id is required")
	}
	return s.db.Create(m).Error
}

// FindAllLeave retrieves a paginated list of leave records from the database.
//
// The function takes an http.Request as input and applies various filters based on the request
// parameters such as company ID, search term, date range, and employee IDs. The results can be
// ordered based on a specified order parameter or defaults to ordering by start_date in descending
// order.
//
// The function returns a paginated page of LeaveModel and an error if the operation fails.
func (s *LeaveService) FindAllLeave(request *http.Request) (paginate.Page, error) {
	pg := paginate.New()
	stmt := s.db.
		Preload("Employee", func(tx *gorm.DB) *gorm.DB {
			return tx.Preload("User").Preload("JobTitle")
		}).
		Preload("Approver.User").
		Preload("LeaveCategory").
		Model(&models.LeaveModel{})
	if request.Header.Get("ID-Company") != "" {
		stmt = stmt.Where("company_id = ?", request.Header.Get("ID-Company"))
	}
	if request.URL.Query().Get("search") != "" {
		stmt = stmt.Where("name ilike ? or description ilike ?",
			"%"+request.URL.Query().Get("search")+"%",
			"%"+request.URL.Query().Get("search")+"%",
		)

	}
	if request.URL.Query().Get("start_date") != "" {
		stmt = stmt.Where("start_date >= ?", request.URL.Query().Get("start_date"))
	}

	if request.URL.Query().Get("end_date") != "" {
		stmt = stmt.Where("end_date <= ?", request.URL.Query().Get("end_date"))
	}

	if request.URL.Query().Get("employee_ids") != "" {
		stmt = stmt.Where("employee_id in (?)", strings.Split(request.URL.Query().Get("employee_ids"), ","))
	}

	utils.FixRequest(request)
	page := pg.With(stmt).Request(request).Response(&[]models.LeaveModel{})
	page.Page = page.Page + 1
	return page, nil
}

// FindAllByEmployeeID retrieves a paginated list of leave records for a given employee ID.
//
// This function takes an HTTP request and an employee ID as inputs. It uses GORM to query
// the database for leave records, preloading the associated Employee (with User and JobTitle),
// Approver (with User), and LeaveCategory models. It applies various filters based on the
// request parameters such as search term, date range, and order. The results can be ordered
// based on a specified order parameter or defaults to ordering by start_date in descending order.
//
// Pagination is utilized to manage the result set, and any necessary request modifications
// are applied using the utils.FixRequest utility. The function returns a paginated page of
// LeaveModel and an error if the operation fails.
func (s *LeaveService) FindAllByEmployeeID(request *http.Request, employeeID string) (paginate.Page, error) {
	pg := paginate.New()
	stmt := s.db.
		Preload("Employee", func(tx *gorm.DB) *gorm.DB {
			return tx.Preload("User").Preload("JobTitle")

		}).
		Preload("Approver.User").
		Preload("LeaveCategory").Where("employee_id = ?", employeeID).Model(&models.LeaveModel{})
	if request.URL.Query().Get("search") != "" {
		stmt = stmt.Where("name ilike ? or description ilike ?",
			"%"+request.URL.Query().Get("search")+"%",
			"%"+request.URL.Query().Get("search")+"%",
		)
	}
	if request.URL.Query().Get("search") != "" {
		stmt = stmt.Where("name LIKE ?", "%"+request.URL.Query().Get("search")+"%")
	}
	if request.URL.Query().Get("start_date") != "" && request.URL.Query().Get("end_date") != "" {
		stmt = stmt.Where("start_date BETWEEN ? AND ?", request.URL.Query().Get("start_date"), request.URL.Query().Get("end_date"))
	}
	if request.URL.Query().Get("order") != "" {
		stmt = stmt.Order(request.URL.Query().Get("order"))
	} else {
		stmt = stmt.Order("start_date DESC")
	}
	utils.FixRequest(request)
	page := pg.With(stmt).Request(request).Response(&[]models.LeaveModel{})
	page.Page = page.Page + 1
	return page, nil
}

// FindLeaveByID retrieves a leave record by ID from the database.
//
// The function takes a leave ID as input and uses it to query the database for the leave record.
// The associated Employee (with User, JobTitle, WorkLocation, WorkShift, and Branch), Approver (with User),
// ApprovalByAdmin, and LeaveCategory models are preloaded.
//
// The function returns the leave record and an error if the operation fails. If the leave record is not found,
// a nil pointer is returned together with a gorm.ErrRecordNotFound error.
func (s *LeaveService) FindLeaveByID(id string) (*models.LeaveModel, error) {
	var m models.LeaveModel
	if err := s.db.
		Preload("Employee", func(tx *gorm.DB) *gorm.DB {
			return tx.Preload("User").
				Preload("JobTitle").
				Preload("WorkLocation").
				Preload("WorkShift").
				Preload("Branch")
		}).
		Preload("Approver.User").
		Preload("ApprovalByAdmin").
		Preload("LeaveCategory").Where("id = ?", id).First(&m).Error; err != nil {
		return nil, err
	}

	files := []models.FileModel{}
	s.db.Find(&files, "ref_id = ? AND ref_type = ?", m.ID, "leave")
	m.Files = files

	return &m, nil
}

func (s *LeaveService) CountLeaveSummary(employee *models.EmployeeModel, startDate time.Time, endDate time.Time) (int64, error) {
	var count int64
	err := s.db.
		Select("sum(diff)").
		Table("(?) as t", s.db.Model(&models.LeaveModel{}).
			Select("DATE_PART('Day', end_date::timestamp - start_date::timestamp) + 1 as diff").
			Where("employee_id = ? AND start_date >= ? AND start_date <= ? and status in (?)", employee.ID, startDate.Format(time.RFC3339), endDate.Format(time.RFC3339), []string{"APPROVED", "FINISHED", "DONE"}).
			Where("deleted_at IS NULL")).
		Scan(&count).Error

	if err != nil {
		return int64(employee.AnnualLeaveDays), err
	}

	return int64(employee.AnnualLeaveDays) - count, nil
}

// UpdateLeave updates a leave record in the database.
//
// The function takes a LeaveModel object as input and attempts to
// update the corresponding leave record in the database. If the operation
// is successful, it returns nil; otherwise, it returns an error indicating
// what went wrong.
//
// Parameters:
//
//	m (*models.LeaveModel): The leave model instance to be updated.
//
// Returns:
//
//	error: An error object if the operation fails, or nil if it is successful.
func (s *LeaveService) UpdateLeave(m *models.LeaveModel) error {
	return s.db.Save(m).Error
}

// DeleteLeave deletes an existing leave record from the database.
//
// The function takes a leave ID as parameter and attempts to delete the
// leave record with the given ID from the database. If the operation is
// successful, it returns an error object indicating what went wrong.
//
// Parameters:
//
//	id (string): The ID of the leave record to be deleted.
//
// Returns:
//
//	error: An error object if the operation fails, or nil if it is successful.
func (s *LeaveService) DeleteLeave(id string) error {
	return s.db.Where("id = ?", id).Delete(&models.LeaveModel{}).Error
}

// Delete deletes an existing leave record from the database.
//
// The function takes a leave ID as parameter and attempts to delete the
// leave record with the given ID from the database. If the operation is
// successful, it returns an error object indicating what went wrong.
//
// Parameters:
//
//	id (string): The ID of the leave record to be deleted.
//
// Returns:
//
//	error: An error object if the operation fails, or nil if it is successful.
func (s *LeaveService) Delete(id string) error {
	return s.db.Where("id = ?", id).Delete(&models.LeaveModel{}).Error
}

// CreateLeaveCategory creates a new leave category record in the database.
//
// The function takes a pointer to a LeaveCategory as parameter and creates a new
// leave category record in the database with the given details.
//
// Parameters:
//
//	c (*models.LeaveCategory): The leave category model instance to be created.
//
// Returns:
//
//	error: An error object if the operation fails, or nil if it is successful.
func (s *LeaveService) CreateLeaveCategory(c *models.LeaveCategory) error {
	return s.db.Create(c).Error
}

// FindAllLeaveCategories retrieves a paginated list of leave categories from the database.
//
// The function takes an HTTP request as input and applies filters based on the request
// parameters such as company ID and search term. The results can be filtered by company ID,
// allowing for company-specific or global categories, and can be searched by name.
//
// Pagination is utilized to manage the result set, and any necessary request modifications
// are applied using the utils.FixRequest utility. The function returns a paginated page of
// LeaveCategory and an error if the operation fails.
func (s *LeaveService) FindAllLeaveCategories(request *http.Request) (paginate.Page, error) {
	pg := paginate.New()
	stmt := s.db.Model(&models.LeaveCategory{})
	if request.Header.Get("ID-Company") != "" {
		stmt = stmt.Where("company_id = ? or company_id is null", request.Header.Get("ID-Company"))
	}
	if request.URL.Query().Get("search") != "" {
		stmt = stmt.Where("name ilike ?", "%"+request.URL.Query().Get("search")+"%")
	}
	utils.FixRequest(request)
	page := pg.With(stmt).Request(request).Response(&[]models.LeaveCategory{})
	page.Page = page.Page + 1
	return page, nil
}

// FindLeaveCategoryByID retrieves a leave category by ID from the database.
//
// The function takes a leave category ID as input and uses it to query the database for the
// leave category record. The function returns the leave category record and an error if the
// operation fails.
//
// Parameters:
//
//	id (string): The ID of the leave category to be retrieved.
//
// Returns:
//
//	*models.LeaveCategory, error: The leave category model instance and an error object if the
//	operation fails, or nil if it is successful.
func (s *LeaveService) FindLeaveCategoryByID(id string) (*models.LeaveCategory, error) {
	var category models.LeaveCategory
	if err := s.db.Where("id = ?", id).First(&category).Error; err != nil {
		return nil, err
	}
	return &category, nil
}

// UpdateLeaveCategory updates an existing leave category record in the database.
//
// The function takes a pointer to a LeaveCategory as parameter and updates the
// corresponding record in the database with the given details.
//
// Parameters:
//
//	c (*models.LeaveCategory): The leave category model instance to be updated.
//
// Returns:
//
//	error: An error object if the operation fails, or nil if it is successful.
func (s *LeaveService) UpdateLeaveCategory(c *models.LeaveCategory) error {
	return s.db.Save(c).Error
}

// DeleteLeaveCategory removes a leave category from the database by its ID.
//
// The function takes a leave category ID as input and attempts to delete the
// corresponding record from the database. If the operation is successful, it
// returns nil; otherwise, it returns an error indicating what went wrong.
//
// Parameters:
//
//	id (string): The ID of the leave category to be deleted.
//
// Returns:
//
//	error: An error object if the operation fails, or nil if it is successful.
func (s *LeaveService) DeleteLeaveCategory(id string) error {
	return s.db.Where("id = ?", id).Delete(&models.LeaveCategory{}).Error
}

// GenLeaveCategories is a utility function that generates a set of standard leave categories.
//
// It inserts the following leave categories into the database:
//
// - Dinas Luar Kota
// - Cuti Menikah
// - Cuti Menikahkan Anak
// - Cuti Khitanan Anak
// - Cuti Baptis Anak
// - Cuti Istri Melahirkan atau Keguguran
// - Cuti Keluarga Meninggal
// - Cuti Anggota Keluarga Dalam Satu Rumah Meninggal
// - Cuti Ibadah Haji
// - Cuti Diluar Tanggungan
// - Pergantian Overtime
// - Pergantian Shift/Jadwal
// - Izin Lainnya
// - Izin Sakit
// - Sakit dengan Surat Dokter
// - Absen
//
// Note that this function is only intended to be called once, during the initial setup of the
// system. It is not intended to be called by the normal flow of the system.
func (s *LeaveService) GenLeaveCategories() {
	cats := []string{
		"Dinas Luar Kota",
		"Cuti Menikah",
		"Cuti Menikahkan Anak",
		"Cuti Khitanan Anak",
		"Cuti Baptis Anak",
		"Cuti Istri Melahirkan atau Keguguran",
		"Cuti Keluarga Meninggal",
		"Cuti Anggota Keluarga Dalam Satu Rumah Meninggal",
		"Cuti Ibadah Haji",
		"Cuti Diluar Tanggungan",
		"Pergantian Overtime",
		"Pergantian Shift/Jadwal",
		"Izin Lainnya",
	}

	for _, v := range cats {
		if s.ctx.DB.Where("name = ?", v).First(&models.LeaveCategory{}).Error == nil {
			continue
		}
		s.ctx.DB.Create(&models.LeaveCategory{
			Name: v,
		})
	}

	sicks := []string{
		"Izin Sakit",
		"Sakit dengan Surat Dokter",
	}
	for _, v := range sicks {
		s.ctx.DB.Create(&models.LeaveCategory{
			Name: v,
			Sick: true,
		})
	}

	s.ctx.DB.Create(&models.LeaveCategory{
		Name:   "Absen",
		Absent: true,
	})
}

// CountByEmployeeID returns the count of approved leaves for a given employee ID and date range.
//
// Parameters:
//
//	employeeID (string): The ID of the employee whose leaves are being counted.
//	startDate (*time.Time): The start date of the date range for filtering leave records.
//	endDate (*time.Time): The end date of the date range for filtering leave records.
//
// Returns:
//
//	int64, error: The count of approved leaves and an error object if the operation fails, or nil if it is successful.
func (s *LeaveService) CountByEmployeeID(employeeID string, startDate *time.Time, endDate *time.Time) (int64, error) {
	var countPending int64
	err := s.ctx.DB.Model(&models.LeaveModel{}).
		Where("status = ?", "APPROVED").
		Where("employee_id = ?", employeeID).
		Where("start_date >= ?", startDate).
		Where("start_date <= ?", endDate).
		Count(&countPending).Error

	return countPending, err
}
