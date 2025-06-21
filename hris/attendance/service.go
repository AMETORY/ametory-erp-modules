package attendance

import (
	"net/http"

	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/shared/models"
	"github.com/AMETORY/ametory-erp-modules/utils"
	"github.com/morkid/paginate"
	"gorm.io/gorm"
)

type AttendanceService struct {
	db  *gorm.DB
	ctx *context.ERPContext
}

func NewAttendanceService(ctx *context.ERPContext) *AttendanceService {
	return &AttendanceService{db: ctx.DB, ctx: ctx}
}

func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&models.AttendanceModel{},
	)
}

func (a *AttendanceService) Create(m *models.AttendanceModel) error {
	return a.db.Create(m).Error
}

func (a *AttendanceService) FindOne(id string) (*models.AttendanceModel, error) {
	m := &models.AttendanceModel{}
	if err := a.db.Where("id = ?", id).First(m).Error; err != nil {
		return nil, err
	}

	return m, nil
}

func (a *AttendanceService) FindAll(request *http.Request) (paginate.Page, error) {
	pg := paginate.New()
	stmt := a.db.Model(&models.AttendanceModel{})
	if request.Header.Get("ID-Company") != "" {
		stmt = stmt.Where("company_id = ?", request.Header.Get("ID-Company"))
	}
	utils.FixRequest(request)
	page := pg.With(stmt).Request(request).Response(&[]models.AttendanceModel{})
	page.Page = page.Page + 1
	return page, nil
}

func (a *AttendanceService) Update(id string, m *models.AttendanceModel) error {
	return a.db.Where("id = ?", id).Updates(m).Error
}

func (a *AttendanceService) Delete(id string) error {
	return a.db.Where("id = ?", id).Delete(&models.AttendanceModel{}).Error
}
