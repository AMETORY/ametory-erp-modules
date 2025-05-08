package task_attribute

import (
	"net/http"

	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/shared/models"
	"github.com/AMETORY/ametory-erp-modules/utils"
	"github.com/morkid/paginate"
	"gorm.io/gorm"
)

type TaskAttributeService struct {
	db  *gorm.DB
	ctx *context.ERPContext
}

func NewTaskAttibuteService(ctx *context.ERPContext) *TaskAttributeService {
	return &TaskAttributeService{
		db:  ctx.DB,
		ctx: ctx,
	}
}

// CreateTaskAttribute creates a new task attribute
func (s *TaskAttributeService) CreateTaskAttribute(data *models.TaskAttributeModel) error {
	return s.db.Create(data).Error
}

// GetTaskAttributeByID retrieves a task attribute by its ID
func (s *TaskAttributeService) GetTaskAttributeByID(id string) (*models.TaskAttributeModel, error) {
	var taskAttribute models.TaskAttributeModel
	err := s.db.First(&taskAttribute, "id = ?", id).Error
	return &taskAttribute, err
}

// UpdateTaskAttribute updates an existing task attribute
func (s *TaskAttributeService) UpdateTaskAttribute(id string, data *models.TaskAttributeModel) error {
	return s.db.Model(&models.TaskAttributeModel{}).Where("id = ?", id).Updates(data).Error
}

// DeleteTaskAttribute deletes a task attribute by its ID
func (s *TaskAttributeService) DeleteTaskAttribute(id string) error {
	return s.db.Delete(&models.TaskAttributeModel{}, "id = ?", id).Error
}

// GetTaskAttributes retrieves a list of task attributes with pagination
func (s *TaskAttributeService) GetTaskAttributes(request *http.Request, search string) (paginate.Page, error) {
	pg := paginate.New()
	stmt := s.db.Model(&models.TaskAttributeModel{})
	if search != "" {
		stmt = stmt.Where("title ILIKE ? OR description ILIKE ?", "%"+search+"%", "%"+search+"%")
	}
	if request.Header.Get("ID-Company") != "" {
		stmt = stmt.Where("company_id = ?", request.Header.Get("ID-Company"))
	}
	stmt = stmt.Order("created_at asc")
	utils.FixRequest(request)
	page := pg.With(stmt).Request(request).Response(&[]models.TaskAttributeModel{})
	page.Page = page.Page + 1
	return page, nil
}
