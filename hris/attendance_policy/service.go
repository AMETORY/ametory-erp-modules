package attendance_policy

import (
	"errors"
	"net/http"

	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/shared/models"
	"github.com/AMETORY/ametory-erp-modules/utils"
	"github.com/morkid/paginate"
	"gorm.io/gorm"
)

type AttendancePolicyService struct {
	db  *gorm.DB
	ctx *context.ERPContext
}

func NewAttendancePolicyService(ctx *context.ERPContext) *AttendancePolicyService {
	return &AttendancePolicyService{db: ctx.DB, ctx: ctx}
}

func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&models.AttendancePolicy{},
	)
}

func (s *AttendancePolicyService) Create(input *models.AttendancePolicy) error {
	return s.db.Create(input).Error
}

func (s *AttendancePolicyService) FindOne(id string) (*models.AttendancePolicy, error) {
	var input models.AttendancePolicy
	err := s.db.Where("id = ?", id).First(&input).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		return nil, err
	}
	return &input, nil
}

func (a *AttendancePolicyService) FindAll(request *http.Request) (paginate.Page, error) {
	pg := paginate.New()
	stmt := a.db.Model(&models.AttendancePolicy{})
	if request.Header.Get("ID-Company") != "" {
		stmt = stmt.Where("company_id = ?", request.Header.Get("ID-Company"))
	}
	utils.FixRequest(request)
	page := pg.With(stmt).Request(request).Response(&[]models.AttendancePolicy{})
	page.Page = page.Page + 1
	return page, nil
}

func (s *AttendancePolicyService) Update(input *models.AttendancePolicy) error {
	return s.db.Save(input).Error
}

func (s *AttendancePolicyService) Delete(id string) error {
	return s.db.Delete(&models.AttendancePolicy{}, "id = ?", id).Error
}
