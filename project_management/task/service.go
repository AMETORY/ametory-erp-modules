package task

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

// TaskService provides methods to manage tasks.
//
// It includes operations such as creating, updating, deleting, and retrieving tasks.
// The service requires a Gorm database instance and an ERP context.
type TaskService struct {
	db              *gorm.DB
	ctx             *context.ERPContext
	queryConditions map[string][]any
	joinConditions  map[string][]any
}

// NewTaskService creates a new instance of TaskService.
//
// It initializes the service with the provided ERP context.
func NewTaskService(ctx *context.ERPContext) *TaskService {
	return &TaskService{
		db:              ctx.DB,
		ctx:             ctx,
		queryConditions: make(map[string][]any, 0),
		joinConditions:  make(map[string][]any, 0),
	}
}

// CreateTask creates a new task in the database.
//
// It takes a TaskModel pointer as input and returns an error if any.
func (s *TaskService) CreateTask(data *models.TaskModel) error {
	return s.db.Create(data).Error
}

// UpdateTask updates an existing task in the database.
//
// It takes a task ID and a TaskModel pointer as input and returns an error if any.
func (s *TaskService) UpdateTask(id string, data *models.TaskModel) error {
	return s.db.Where("id = ?", id).Updates(data).Error
}

// DeleteTask deletes a task from the database.
//
// It takes a task ID as input and returns an error if any.
func (s *TaskService) DeleteTask(id string) error {
	return s.db.Where("id = ?", id).Delete(&models.TaskModel{}).Error
}

// GetTaskByID retrieves a task from the database by its ID.
//
// It takes a task ID as input and returns a TaskModel pointer and an error if any.
// The function preloads associated Tags, TaskAttribute, FormResponse, Activities, Assignee.User,
// Watchers.User, and Comments data.
func (s *TaskService) GetTaskByID(id string) (*models.TaskModel, error) {
	var invoice models.TaskModel
	err := s.db.Preload("Tags").Preload("TaskAttribute").Preload("FormResponse").Preload("Activities", func(db *gorm.DB) *gorm.DB {
		return s.db.Preload("Member.User").Preload("Column").Preload("Task")
	}).Preload("Assignee.User").Preload("Watchers.User").Preload("Comments", func(db *gorm.DB) *gorm.DB {
		return db.Preload("Member.User").Order("published_at").Where("status = ?", "PUBLISHED")
	}).Where("id = ?", id).First(&invoice).Error
	return &invoice, err
}

// SetQuery sets the query conditions for the TaskService.
//
// It takes a map of query conditions as input.
func (s *TaskService) SetQuery(query map[string][]interface{}) {
	s.queryConditions = query
}

// SetJoins sets the join conditions for the TaskService.
//
// It takes a map of join conditions as input.
func (s *TaskService) SetJoins(joins map[string][]interface{}) {
	s.joinConditions = joins
}

// GetTasks retrieves a paginated list of tasks, optionally filtering by search, project ID, and other
// conditions.
//
// It takes an HTTP request, a search string, and an optional project ID as input. The search string is
// applied to the task's name and description fields. If a project ID is present in the request query,
// the result is filtered by the project ID. The function supports ordering and pagination.
//
// The function preloads associated Tags, Assignee.User, and Watchers.User data.
//
// Returns a paginated page of TaskModel and an error, if any.
func (s *TaskService) GetTasks(request http.Request, search string, projectId *string) (paginate.Page, error) {
	pg := paginate.New()
	stmt := s.db.Preload("Tags").Preload("Assignee.User").Preload("Watchers.User")
	if search != "" {
		stmt = stmt.Where("tasks.name ILIKE ? OR tasks.description ILIKE ?",
			"%"+search+"%",
			"%"+search+"%",
		)
	}
	// if request.Header.Get("ID-Company") != "" {
	// 	stmt = stmt.Where("company_id = ?", request.Header.Get("ID-Company"))
	// }

	if request.URL.Query().Get("column_id") != "" {
		stmt = stmt.Where("column_id = ?", request.URL.Query().Get("column_id"))
	}

	if request.URL.Query().Get("column_id") != "" {
		stmt = stmt.Where("column_id = ?", request.URL.Query().Get("column_id"))
	}
	if projectId != nil {
		stmt = stmt.Where("project_id = ?", *projectId)
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

	stmt = stmt.Order("order_number")
	request.URL.Query().Get("page")
	stmt = stmt.Model(&models.TaskModel{})
	utils.FixRequest(&request)
	page := pg.With(stmt).Request(request).Response(&[]models.TaskModel{})
	page.Page = page.Page + 1
	return page, nil
}

// GetMyTask retrieves a paginated list of tasks that are assigned to the current user.
//
// It takes an HTTP request, a search string, and a member ID as input. The search string is
// applied to the task's name and description fields. The result is filtered to only include
// tasks that are assigned to the given member ID. The function supports ordering and pagination.
//
// The function preloads associated Watchers.User, Assignee.User, Column, Project, and Parent data.
//
// Returns a paginated page of TaskModel and an error, if any.
func (s *TaskService) GetMyTask(request http.Request, search string, memberID string) (paginate.Page, error) {
	pg := paginate.New()
	stmt := s.db.Preload("Watchers.User").Preload("Assignee.User").Preload("Column").Preload("Project").Preload("Parent")
	if search != "" {
		stmt = stmt.Where("tasks.name ILIKE ? OR tasks.description ILIKE ?",
			"%"+search+"%",
			"%"+search+"%",
		)
	}
	stmt = stmt.Where("assignee_id = ?", memberID)
	stmt = stmt.Order("order_number")
	request.URL.Query().Get("page")
	stmt = stmt.Model(&models.TaskModel{})
	utils.FixRequest(&request)
	page := pg.With(stmt).Request(request).Response(&[]models.TaskModel{})
	page.Page = page.Page + 1
	return page, nil
}

// GetWatchedTask retrieves a paginated list of tasks that are being watched by the current user.
//
// It takes an HTTP request, a search string, and a member ID as input. The search string is
// applied to the task's name and description fields. The result is filtered to only include
// tasks that are being watched by the given member ID. The function supports ordering and pagination.
//
// The function preloads associated Watchers.User, Assignee.User, Column, Project, and Parent data.
//
// Returns a paginated page of TaskModel and an error, if any.
func (s *TaskService) GetWatchedTask(request http.Request, search string, memberID string) (paginate.Page, error) {
	pg := paginate.New()
	stmt := s.db.Preload("Watchers.User").Preload("Assignee.User").Preload("Column").Preload("Project").Preload("Parent")
	if search != "" {
		stmt = stmt.Where("tasks.name ILIKE ? OR tasks.description ILIKE ?",
			"%"+search+"%",
			"%"+search+"%",
		)
	}
	stmt = stmt.Joins("JOIN task_watchers ON tasks.id = task_watchers.task_model_id").
		Where("task_watchers.member_model_id = ?", memberID)
	stmt = stmt.Order("order_number")
	request.URL.Query().Get("page")
	stmt = stmt.Model(&models.TaskModel{})
	utils.FixRequest(&request)
	page := pg.With(stmt).Request(request).Response(&[]models.TaskModel{})
	page.Page = page.Page + 1
	return page, nil
}

// GetDateRangeFromRequest parses the start-date and end-date query parameters from a given
// HTTP request into a start date and end date.
//
// Returns the start date, end date, and an error if the operation fails.
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

// CountTasksInColumn returns the number of tasks in a given column.
//
// Returns the number of tasks and an error, if any.
func (s *TaskService) CountTasksInColumn(columnID string) (int64, error) {
	stmt := s.db.Model(&models.TaskModel{}).Where("column_id = ?", columnID)
	var count int64
	if err := stmt.Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

// MoveTask moves a given task to a new column and updates the order number.
//
// Returns an error if the operation fails.
func (s *TaskService) MoveTask(columnID string, taskID string, sourceColumnID string, orderNumber int) error {

	var task models.TaskModel
	if err := s.db.Where("id = ?", taskID).First(&task).Error; err != nil {
		return err
	}
	var sourceColumn models.ColumnModel
	if err := s.db.Where("id = ?", sourceColumnID).First(&sourceColumn).Error; err != nil {
		return err
	}
	var targetColumn models.ColumnModel
	if err := s.db.Where("id = ?", columnID).First(&targetColumn).Error; err != nil {
		return err
	}
	task.ColumnID = &columnID
	task.OrderNumber = orderNumber
	if err := s.db.Save(&task).Error; err != nil {
		return err
	}
	sourceTasks := make([]models.TaskModel, 0)
	if err := s.db.Where("column_id = ? AND order_number > ?", sourceColumnID, task.OrderNumber).Find(&sourceTasks).Error; err != nil {
		return err
	}
	for _, t := range sourceTasks {
		t.OrderNumber = t.OrderNumber - 1
		if err := s.db.Save(&t).Error; err != nil {
			return err
		}
	}
	targetTasks := make([]models.TaskModel, 0)
	if err := s.db.Where("column_id = ? AND order_number >= ?", columnID, task.OrderNumber).Find(&targetTasks).Error; err != nil {
		return err
	}
	for _, t := range targetTasks {
		t.OrderNumber = t.OrderNumber + 1
		if err := s.db.Save(&t).Error; err != nil {
			return err
		}
	}
	return nil

}

// MarkCompleted marks a task as completed by setting its CompletedDate.
// It takes a task ID as input and returns an error if any.
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

// ReorderTask reorders a task in its column by updating its order number.
// It takes a task ID and a new order number as input and returns an error if any.
func (s *TaskService) ReorderTask(taskID string, order int) error {
	tx := s.db.Begin()
	defer tx.Rollback()
	var task models.TaskModel
	if err := tx.Where("id = ?", taskID).First(&task).Error; err != nil {
		return err
	}
	var tasks []models.TaskModel
	if err := tx.Where("column_id = ? AND order_number >= ?", task.ColumnID, order).Find(&tasks).Error; err != nil {
		return err
	}
	for _, t := range tasks {
		t.OrderNumber = t.OrderNumber + 1
		if err := tx.Save(&t).Error; err != nil {
			return err
		}
	}
	task.OrderNumber = order
	if err := tx.Save(&task).Error; err != nil {
		return err
	}
	return tx.Commit().Error
}

// AddWatchers adds watchers to a task.
// It takes a task ID and a list of watcher IDs as input and returns an error if any.
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

// RemoveWatcher removes a watcher from a task.
// It takes a task ID and a watcher ID as input and returns an error if any.
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

// isContains checks if a given string is present in a slice of MemberModel.
// It returns true if the string is found, otherwise false.
func isContains(arr []models.MemberModel, str string) bool {
	for _, a := range arr {
		if a.ID == str {
			return true
		}
	}
	return false
}

// CreateComment creates a new comment for a task.
// It takes a task ID, a TaskCommentModel pointer, and a flag for auto-publishing as input.
// Returns an error if any.
func (s *TaskService) CreateComment(taskID string, comment *models.TaskCommentModel, autoPublish bool) error {
	now := time.Now()

	var task models.TaskModel
	if err := s.db.Where("id = ?", taskID).First(&task).Error; err != nil {
		return err
	}

	comment.TaskID = taskID
	if autoPublish {
		comment.Status = "PUBLISHED"
		comment.PublishedAt = &now
	}

	return s.db.Create(comment).Error
}

// UpdateStatusComment updates the status of a comment and sets its PublishedAt date if necessary.
// It takes a comment ID and a new status as input and returns an error if any.
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
