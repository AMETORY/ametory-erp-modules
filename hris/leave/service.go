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

func NewLeaveService(ctx *context.ERPContext, employeeService *employee.EmployeeService) *LeaveService {
	return &LeaveService{db: ctx.DB, ctx: ctx, employeeService: employeeService}
}

func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&models.LeaveModel{},
		&models.LeaveCategory{},
	)
}

func (s *LeaveService) CreateLeave(m *models.LeaveModel) error {
	if m.EmployeeID == nil {
		return errors.New("employee id is required")
	}
	return s.db.Create(m).Error
}

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

func (s *LeaveService) UpdateLeave(m *models.LeaveModel) error {
	return s.db.Save(m).Error
}

func (s *LeaveService) DeleteLeave(id string) error {
	return s.db.Where("id = ?", id).Delete(&models.LeaveModel{}).Error
}

func (s *LeaveService) Delete(id string) error {
	return s.db.Where("id = ?", id).Delete(&models.LeaveModel{}).Error
}

func (s *LeaveService) CreateLeaveCategory(c *models.LeaveCategory) error {
	return s.db.Create(c).Error
}

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

func (s *LeaveService) FindLeaveCategoryByID(id string) (*models.LeaveCategory, error) {
	var category models.LeaveCategory
	if err := s.db.Where("id = ?", id).First(&category).Error; err != nil {
		return nil, err
	}
	return &category, nil
}

func (s *LeaveService) UpdateLeaveCategory(c *models.LeaveCategory) error {
	return s.db.Save(c).Error
}

func (s *LeaveService) DeleteLeaveCategory(id string) error {
	return s.db.Where("id = ?", id).Delete(&models.LeaveCategory{}).Error
}

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
