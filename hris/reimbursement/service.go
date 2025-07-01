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

func NewReimbursementService(ctx *context.ERPContext, employeeService *employee.EmployeeService) *ReimbursementService {
	return &ReimbursementService{db: ctx.DB, ctx: ctx, employeeService: employeeService}
}

func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&models.ReimbursementModel{},
		&models.ReimbursementItemModel{},
	)
}

func (s *ReimbursementService) CreateReimbursement(m *models.ReimbursementModel) error {
	if m.EmployeeID == nil {
		return errors.New("employee id is required")
	}
	return s.db.Create(m).Error
}

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

func (s *ReimbursementService) UpdateReimbursement(m *models.ReimbursementModel) error {
	return s.db.Save(m).Error
}

func (s *ReimbursementService) DeleteReimbursement(id string) error {
	return s.db.Where("id = ?", id).Delete(&models.ReimbursementModel{}).Error
}

func (s *ReimbursementService) Delete(id string) error {
	return s.db.Where("id = ?", id).Delete(&models.ReimbursementModel{}).Error
}

func (s *ReimbursementService) CreateReimbursementItem(m *models.ReimbursementItemModel) error {
	if m.ReimbursementID == nil {
		return errors.New("reimbursement id is required")
	}
	return s.db.Create(m).Error
}

func (s *ReimbursementService) UpdateReimbursementItem(id string, m *models.ReimbursementItemModel) error {
	return s.db.Where("id = ?", id).Save(m).Error
}

func (s *ReimbursementService) DeleteReimbursementItem(id string) error {
	return s.db.Where("id = ?", id).Delete(&models.ReimbursementItemModel{}).Error
}

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
