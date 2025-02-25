package models

import (
	"strings"
	"time"

	"github.com/AMETORY/ametory-erp-modules/shared"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TaskModel struct {
	shared.BaseModel
	Name        string        `gorm:"type:varchar(255);not null" json:"name,omitempty"`
	Description string        `gorm:"type:text" json:"description,omitempty"`
	ProjectID   string        `gorm:"type:char(36);not null" json:"project_id,omitempty"`
	Project     ProjectModel  `json:"project,omitempty"`
	CreatedByID *string       `gorm:"type:char(36);index" json:"created_by_id,omitempty"`
	CreatedBy   *MemberModel  `json:"created_by,omitempty"`
	AssigneeID  *string       `gorm:"type:char(36);index" json:"assignee_id,omitempty"`
	Assignee    *MemberModel  `json:"assignee,omitempty"`
	ParentID    *string       `gorm:"type:char(36);index" json:"parent_id,omitempty"`
	Parent      *TaskModel    `json:"parent,omitempty"`
	Children    []TaskModel   `gorm:"foreignKey:ParentID" json:"children,omitempty"`
	Order       int           `json:"order,omitempty"`
	Status      string        `gorm:"type:varchar(50);not null" json:"status,omitempty"`
	IsDone      bool          `json:"is_done,omitempty"`
	StartDate   *time.Time    `json:"start_date,omitempty"`
	EndDate     *time.Time    `json:"end_date,omitempty"`
	Files       []FileModel   `gorm:"-" json:"files,omitempty"`
	Cover       *FileModel    `json:"cover,omitempty"`
	Watchers    []MemberModel `gorm:"many2many:task_watchers" json:"watchers,omitempty"`
}

func (TaskModel) TableName() string {
	return "tasks"
}

func (t *TaskModel) BeforeCreate(tx *gorm.DB) error {
	t.ID = uuid.New().String()
	if t.Order == 0 {
		var lastTask TaskModel
		if err := tx.Where("project_id = ?", t.ProjectID).Order("order DESC").First(&lastTask).Error; err != nil {
			return err
		}
		t.Order = lastTask.Order + 1
	}
	return nil
}

func (t *TaskModel) AfterFind(tx *gorm.DB) error {
	var files []FileModel
	if err := tx.Where("ref_id = ? AND ref_type = ?", t.ID, "task").Find(&files).Error; err != nil {
		return err
	}
	t.Files = files
	if len(files) > 0 {
		if strings.HasPrefix(files[0].MimeType, "image/") {
			t.Cover = &files[0]
		}
	}
	return nil
}

func (t *TaskModel) GetRecursiveChildren(tx *gorm.DB) ([]TaskModel, error) {
	children := make([]TaskModel, 0)
	if err := tx.Where("parent_id = ?", t.ID).Find(&children).Error; err != nil {
		return nil, err
	}

	for _, child := range children {
		subChildren, err := child.GetRecursiveChildren(tx)
		if err != nil {
			return nil, err
		}
		children = append(children, subChildren...)
	}

	return children, nil
}
