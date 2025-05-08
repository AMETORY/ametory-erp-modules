package form

import (
	"net/http"

	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/shared/models"
	"github.com/AMETORY/ametory-erp-modules/utils"
	"github.com/morkid/paginate"
	"gorm.io/gorm"
)

type FormService struct {
	db  *gorm.DB
	ctx *context.ERPContext
}

func NewFormService(db *gorm.DB, ctx *context.ERPContext) *FormService {
	return &FormService{db: db, ctx: ctx}
}

func (s *FormService) CreateFormTemplate(formTemplate *models.FormTemplate) error {
	return s.db.Create(formTemplate).Error
}

func (s *FormService) GetFormTemplate(id string) (*models.FormTemplate, error) {
	var formTemplate models.FormTemplate

	if err := s.db.First(&formTemplate, "id = ?", id).Error; err != nil {
		return nil, err
	}

	return &formTemplate, nil
}

func (s *FormService) GetFormTemplates(request http.Request, search string) (paginate.Page, error) {
	pg := paginate.New()
	stmt := s.db.
		Preload("CreatedBy").
		Preload("CreatedByMember.User").
		Model(&models.FormTemplate{})
	if search != "" {
		stmt = stmt.Where("title LIKE ?", "%"+search+"%")
	}

	if request.Header.Get("ID-Company") != "" {
		stmt = stmt.Where("company_id = ? or company_id is null", request.Header.Get("ID-Company"))
	}

	utils.FixRequest(&request)
	page := pg.With(stmt).Request(request).Response(&[]models.FormTemplate{})
	page.Page = page.Page + 1

	return page, nil
}

func (s *FormService) UpdateFormTemplate(id string, formTemplate *models.FormTemplate) error {
	return s.db.Where("id = ?", id).Omit("id").Updates(formTemplate).Error
}

func (s *FormService) DeleteFormTemplate(id string) error {
	return s.db.Delete(&models.FormTemplate{}, id).Error
}

func (s *FormService) CreateForm(form *models.FormModel) error {
	return s.db.Create(form).Error
}

func (s *FormService) GetForm(id string) (*models.FormModel, error) {
	var form models.FormModel

	if err := s.db.
		Preload("Responses").
		Preload("FormTemplate").
		Preload("CreatedBy").
		Preload("CreatedByMember.User").
		Preload("Project").
		Preload("Column").
		First(&form, "id = ?", id).Error; err != nil {
		return nil, err
	}

	return &form, nil
}
func (s *FormService) GetFormByCode(code string) (*models.FormModel, error) {
	var form models.FormModel

	if err := s.db.
		Preload("FormTemplate").
		Preload("CreatedBy").
		Preload("CreatedByMember.User").
		Preload("Project").
		Preload("Column").
		First(&form, "code = ?", code).Error; err != nil {
		return nil, err
	}

	return &form, nil
}

func (s *FormService) GetForms(request http.Request, search string) (paginate.Page, error) {
	pg := paginate.New()
	stmt := s.db.
		Preload("FormTemplate").
		Preload("CreatedBy").
		Preload("CreatedByMember.User").
		Preload("Project").
		Preload("Column").
		Model(&models.FormModel{})
	if search != "" {
		stmt = stmt.Where("title LIKE ? or description LIKE ?",
			"%"+search+"%",
			"%"+search+"%",
		)
	}

	utils.FixRequest(&request)
	page := pg.With(stmt).Request(request).Response(&[]models.FormModel{})
	page.Page = page.Page + 1

	return page, nil
}

func (s *FormService) UpdateForm(id string, form *models.FormModel) error {
	return s.db.Where("id = ?", id).Omit("id").Updates(form).Error
}

func (s *FormService) DeleteForm(id string) error {
	return s.db.Delete(&models.FormModel{}, id).Error
}
