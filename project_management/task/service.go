package team

import (
	"errors"
	"net/http"
	"time"

	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/shared"
	"github.com/AMETORY/ametory-erp-modules/shared/models"
	"github.com/AMETORY/ametory-erp-modules/utils"
	"github.com/morkid/paginate"
	"gorm.io/gorm"
)

type TaskService struct {
	db              *gorm.DB
	ctx             *context.ERPContext
	queryConditions map[string][]interface{}
	joinConditions  map[string][]interface{}
}

func NewTaskService(ctx *context.ERPContext) *TaskService {
	return &TaskService{
		db:              ctx.DB,
		ctx:             ctx,
		queryConditions: make(map[string][]interface{}, 0),
		joinConditions:  make(map[string][]interface{}, 0),
	}
}

func (s *TaskService) CreateTask(data *models.TaskModel) error {
	return s.db.Create(data).Error
}

func (s *TaskService) UpdateTask(id string, data *models.TaskModel) error {
	return s.db.Where("id = ?", id).Updates(data).Error
}

func (s *TaskService) DeleteTask(id string) error {
	return s.db.Where("id = ?", id).Delete(&models.TaskModel{}).Error
}

func (s *TaskService) GetTaskByID(id string) (*models.TaskModel, error) {
	var invoice models.TaskModel
	err := s.db.Where("id = ?", id).First(&invoice).Error
	return &invoice, err
}

func (s *TaskService) SetQuery(query map[string][]interface{}) {
	s.queryConditions = query
}
func (s *TaskService) SetJoins(joins map[string][]interface{}) {
	s.joinConditions = joins
}
func (s *TaskService) GetTasks(request http.Request, search string) (paginate.Page, error) {
	pg := paginate.New()
	stmt := s.db
	if search != "" {
		stmt = stmt.Where("tasks.name ILIKE ? OR tasks.description ILIKE ?",
			"%"+search+"%",
			"%"+search+"%",
		)
	}
	if request.Header.Get("ID-Company") != "" {
		stmt = stmt.Where("company_id = ?", request.Header.Get("ID-Company"))
	}

	if request.URL.Query().Get("column_id") != "" {
		stmt = stmt.Where("column_id = ?", request.URL.Query().Get("column_id"))
	}

	if request.URL.Query().Get("project_id") != "" {
		stmt = stmt.Where("project_id = ?", request.URL.Query().Get("project_id"))
	}
	if request.URL.Query().Get("created_by_id") != "" {
		stmt = stmt.Where("created_by_id = ?", request.URL.Query().Get("created_by_id"))
	}
	if request.URL.Query().Get("assignee_id") != "" {
		stmt = stmt.Where("assignee_id = ?", request.URL.Query().Get("assignee_id"))
	}
	if request.URL.Query().Get("parent_id") != "" {
		stmt = stmt.Where("parent_id = ?", request.URL.Query().Get("parent_id"))
	}

	if startDate, endDate, err := s.GetDateRangeFromRequest(request); err == nil {
		stmt = stmt.Where("date BETWEEN ? AND ?", startDate, endDate)
	}

	for k, v := range s.queryConditions {
		stmt = stmt.Where(k, v...)
	}
	for k, v := range s.joinConditions {
		stmt = stmt.Joins(k, v...)
	}

	request.URL.Query().Get("page")
	stmt = stmt.Model(&models.TaskModel{})
	utils.FixRequest(&request)
	page := pg.With(stmt).Request(request).Response(&[]models.TaskModel{})
	page.Page = page.Page + 1
	return page, nil
}

func (s *TaskService) GetDateRangeFromRequest(request http.Request) (time.Time, time.Time, error) {
	startDateStr := request.URL.Query().Get("start_date")
	endDateStr := request.URL.Query().Get("end_date")
	if startDateStr == "" || endDateStr == "" {
		return time.Time{}, time.Time{}, errors.New("start-date and end-date are required")
	}
	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	return startDate, endDate, nil
}

func (s *TaskService) MoveTask(columnID string, taskID string, sourceColumnID string) error {
	tx := s.db.Begin()
	defer tx.Rollback()
	var task models.TaskModel
	if err := tx.Where("id = ?", taskID).First(&task).Error; err != nil {
		return err
	}
	var sourceColumn models.ColumnModel
	if err := tx.Where("id = ?", sourceColumnID).First(&sourceColumn).Error; err != nil {
		return err
	}
	var targetColumn models.ColumnModel
	if err := tx.Where("id = ?", columnID).First(&targetColumn).Error; err != nil {
		return err
	}
	task.ColumnID = &columnID
	task.Order = 0
	if err := tx.Save(&task).Error; err != nil {
		return err
	}
	sourceTasks := make([]models.TaskModel, 0)
	if err := tx.Where("column_id = ? AND order > ?", sourceColumnID, task.Order).Find(&sourceTasks).Error; err != nil {
		return err
	}
	for _, t := range sourceTasks {
		t.Order = t.Order - 1
		if err := tx.Save(&t).Error; err != nil {
			return err
		}
	}
	targetTasks := make([]models.TaskModel, 0)
	if err := tx.Where("column_id = ? AND order >= ?", columnID, task.Order).Find(&targetTasks).Error; err != nil {
		return err
	}
	for _, t := range targetTasks {
		t.Order = t.Order + 1
		if err := tx.Save(&t).Error; err != nil {
			return err
		}
	}
	return tx.Commit().Error
}

func (s *TaskService) MarkCompleted(id string) error {
	tx := s.db.Begin()
	defer tx.Rollback()
	var task models.TaskModel
	if err := tx.Where("id = ?", id).First(&task).Error; err != nil {
		return err
	}
	task.CompletedDate = &time.Time{}
	if err := tx.Save(&task).Error; err != nil {
		return err
	}
	return tx.Commit().Error
}

func (s *TaskService) ReorderTask(taskID string, order int) error {
	tx := s.db.Begin()
	defer tx.Rollback()
	var task models.TaskModel
	if err := tx.Where("id = ?", taskID).First(&task).Error; err != nil {
		return err
	}
	var tasks []models.TaskModel
	if err := tx.Where("column_id = ? AND order >= ?", task.ColumnID, order).Find(&tasks).Error; err != nil {
		return err
	}
	for _, t := range tasks {
		t.Order = t.Order + 1
		if err := tx.Save(&t).Error; err != nil {
			return err
		}
	}
	task.Order = order
	if err := tx.Save(&task).Error; err != nil {
		return err
	}
	return tx.Commit().Error
}

func (s *TaskService) AddWatchers(id string, watchers []string) error {
	tx := s.db.Begin()
	defer tx.Rollback()
	var task models.TaskModel
	if err := tx.Where("id = ?", id).First(&task).Error; err != nil {
		return err
	}
	for _, watcher := range watchers {
		if !isContains(task.Watchers, watcher) {
			task.Watchers = append(task.Watchers, models.MemberModel{BaseModel: shared.BaseModel{ID: watcher}})
		}
	}
	if err := tx.Save(&task).Error; err != nil {
		return err
	}
	return tx.Commit().Error
}

func (s *TaskService) RemoveWatcher(taskID string, watcherID string) error {
	tx := s.db.Begin()
	defer tx.Rollback()
	var task models.TaskModel
	if err := tx.Where("id = ?", taskID).First(&task).Error; err != nil {
		return err
	}
	for i, watcher := range task.Watchers {
		if watcher.ID == watcherID {
			task.Watchers = append(task.Watchers[:i], task.Watchers[i+1:]...)
			break
		}
	}
	if err := tx.Save(&task).Error; err != nil {
		return err
	}
	return tx.Commit().Error
}

func isContains(arr []models.MemberModel, str string) bool {
	for _, a := range arr {
		if a.ID == str {
			return true
		}
	}
	return false
}

func (s *TaskService) CreateComment(taskID string, comment *models.TaskCommentModel, autoPublish bool) error {
	now := time.Now()
	tx := s.db.Begin()
	defer tx.Rollback()

	var task models.TaskModel
	if err := tx.Where("id = ?", taskID).First(&task).Error; err != nil {
		return err
	}

	comment.TaskID = taskID
	if autoPublish {
		comment.Status = "PUBLISHED"
		comment.PublishedAt = &now
	}
	if err := tx.Create(comment).Error; err != nil {
		return err
	}

	return tx.Commit().Error
}

func (s *TaskService) UpdateStatusComment(commentID string, status string) error {
	now := time.Now()
	tx := s.db.Begin()
	defer tx.Rollback()

	var comment models.TaskCommentModel
	if err := tx.Where("id = ?", commentID).First(&comment).Error; err != nil {
		return err
	}

	comment.Status = status
	comment.PublishedAt = &now
	if err := tx.Save(&comment).Error; err != nil {
		return err
	}

	return tx.Commit().Error
}
