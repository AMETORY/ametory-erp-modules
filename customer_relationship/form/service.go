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

// NewFormService creates a new instance of FormService with the given database and ERP context.
// It initializes the FormService with the provided gorm database and ERP context, allowing
// interaction with form-related data within the specified context.

func NewFormService(db *gorm.DB, ctx *context.ERPContext) *FormService {
	return &FormService{db: db, ctx: ctx}
}

// CreateFormTemplate creates a new form template in the database.
// It takes a pointer to a `models.FormTemplate` and returns an error if the template
// could not be created.
func (s *FormService) CreateFormTemplate(formTemplate *models.FormTemplate) error {
	return s.db.Create(formTemplate).Error
}

// GetFormTemplate retrieves a form template with the given ID from the database.
// It returns a pointer to a `models.FormTemplate` and an error if the template
// could not be found.
func (s *FormService) GetFormTemplate(id string) (*models.FormTemplate, error) {
	var formTemplate models.FormTemplate

	if err := s.db.First(&formTemplate, "id = ?", id).Error; err != nil {
		return nil, err
	}

	return &formTemplate, nil
}

// GetFormTemplates retrieves a list of form templates from the database, filtered by the given search string if any.
// It takes a pointer to an http request and a search string, and returns a paginate.Page containing a list of
// form templates, and an error if the templates could not be retrieved.
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

// UpdateFormTemplate updates an existing form template.
//
// It takes a string id and a pointer to a FormTemplate as parameters and
// returns an error. It uses the gorm.DB connection to update a record in the
// form_templates table.
func (s *FormService) UpdateFormTemplate(id string, formTemplate *models.FormTemplate) error {
	return s.db.Where("id = ?", id).Omit("id").Updates(formTemplate).Error
}

// DeleteFormTemplate deletes a form template with the specified ID from the database.
// It takes a string ID as a parameter and returns an error if the deletion fails.

func (s *FormService) DeleteFormTemplate(id string) error {
	return s.db.Delete(&models.FormTemplate{}, id).Error
}

// CreateForm creates a new form in the database.
// It takes a pointer to a `models.FormModel` and returns an error if the form
// could not be created.
func (s *FormService) CreateForm(form *models.FormModel) error {
	return s.db.Create(form).Error
}

// GetForm retrieves a form by its ID from the database.
//
// It preloads associated data such as Responses, FormTemplate, CreatedBy, CreatedByMember.User, Project, and Column.
// Returns a pointer to a FormModel and an error if the form could not be found or any other database error occurs.

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

// GetFormByCode retrieves a form by its code from the database.
//
// It preloads associated data such as FormTemplate, CreatedBy, CreatedByMember.User, Project, and Column.
// Returns a pointer to a FormModel and an error if the form could not be found or any other database error occurs.
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

// GetForms retrieves a paginated list of forms from the database.
//
// It takes a pointer to an http request and a search query string as parameters.
// The search query is applied to the form title and description fields.
// The function uses pagination to manage the result set and includes any necessary
// request modifications using the utils.FixRequest utility.
//
// The function returns a paginated page of FormModel and an error if the
// operation fails.
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

// UpdateForm updates an existing form.
//
// It takes a string id and a pointer to a FormModel as parameters and
// returns an error. It uses the gorm.DB connection to update a record in the
// forms table.
func (s *FormService) UpdateForm(id string, form *models.FormModel) error {
	return s.db.Where("id = ?", id).Omit("id").Updates(form).Error
}

// DeleteForm deletes a form from the database.
//
// It takes a string id as a parameter and returns an error. It uses the gorm.DB
// connection to delete a record from the forms table.
func (s *FormService) DeleteForm(id string) error {
	return s.db.Delete(&models.FormModel{}, id).Error
}
