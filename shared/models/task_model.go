package models

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/AMETORY/ametory-erp-modules/shared"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TaskModel struct {
	shared.BaseModel
	Name                    string                 `gorm:"type:varchar(255);not null" json:"name,omitempty"`
	Description             string                 `gorm:"type:text" json:"description,omitempty"`
	ProjectID               string                 `gorm:"type:char(36);not null" json:"project_id,omitempty"`
	Project                 ProjectModel           `json:"project,omitempty" gorm:"foreignKey:ProjectID;constraint:OnDelete:CASCADE;"`
	ColumnID                *string                `gorm:"type:char(36);not null" json:"column_id,omitempty"`
	Column                  *ColumnModel           `json:"column,omitempty" gorm:"foreignKey:ColumnID;constraint:OnDelete:CASCADE;"`
	CreatedByID             *string                `gorm:"type:char(36);index" json:"created_by_id,omitempty"`
	CreatedBy               *MemberModel           `json:"created_by,omitempty" gorm:"foreignKey:CreatedByID;constraint:OnDelete:CASCADE;"`
	AssigneeID              *string                `gorm:"type:char(36);index" json:"assignee_id,omitempty"`
	Assignee                *MemberModel           `json:"assignee,omitempty" gorm:"foreignKey:AssigneeID;constraint:OnDelete:CASCADE;"`
	ParentID                *string                `gorm:"type:char(36);index" json:"parent_id,omitempty"`
	Parent                  *TaskModel             `json:"parent,omitempty" gorm:"foreignKey:ParentID;constraint:OnDelete:CASCADE;"`
	Children                []TaskModel            `gorm:"foreignKey:ParentID" json:"children,omitempty"`
	OrderNumber             int                    `json:"order_number,omitempty"`
	Status                  string                 `gorm:"type:varchar(50);not null" json:"status,omitempty"`
	Completed               bool                   `json:"completed,omitempty"`
	CompletedDate           *time.Time             `json:"completed_date,omitempty"`
	StartDate               *time.Time             `json:"start_date,omitempty"`
	EndDate                 *time.Time             `json:"end_date,omitempty"`
	Files                   []FileModel            `gorm:"-" json:"files,omitempty"`
	Cover                   *FileModel             `json:"cover,omitempty" gorm:"-"`
	Watchers                []MemberModel          `gorm:"many2many:task_watchers" json:"watchers,omitempty"`
	Comments                []TaskCommentModel     `gorm:"foreignKey:TaskID" json:"comments,omitempty"`
	Activities              []ProjectActivityModel `gorm:"foreignKey:TaskID" json:"activities,omitempty"`
	CommentCount            int                    `gorm:"-" json:"comment_count,omitempty"`
	Percentage              float64                `json:"percentage,omitempty" binding:"max=100"`
	Priority                string                 `gorm:"type:varchar(50)" json:"priority,omitempty"`
	Severity                string                 `gorm:"type:varchar(50)" json:"severity,omitempty"`
	FormResponseID          *string                `gorm:"type:char(36);index" json:"form_response_id,omitempty"`
	FormResponse            *FormResponseModel     `json:"form_response,omitempty" gorm:"foreignKey:FormResponseID;constraint:OnDelete:CASCADE;"`
	TaskAttributeID         *string                `gorm:"type:char(36);index" json:"task_attribute_id,omitempty"`
	TaskAttribute           *TaskAttributeModel    `json:"task_attribute,omitempty" gorm:"foreignKey:TaskAttributeID;constraint:OnDelete:CASCADE;"`
	TaskAttibuteData        *string                `json:"task_attribute_data,omitempty" gorm:"type:JSON"`
	TaskAttributeDataParsed any                    `gorm:"-" json:"task_attribute_data_parsed,omitempty" `
	RefID                   *string                `gorm:"type:char(36);index" json:"ref_id,omitempty"`
	RefType                 *string                `gorm:"type:varchar(255);index" json:"ref_type,omitempty"`
}

func (TaskModel) TableName() string {
	return "tasks"
}

func (t *TaskModel) BeforeCreate(tx *gorm.DB) error {

	if t.ID == "" {
		tx.Statement.SetColumn("id", uuid.New().String())
	}

	if t.Percentage == 100 || t.Completed {
		t.Status = "COMPLETED"
		t.Completed = true
		t.Percentage = 100
		now := time.Now()
		t.CompletedDate = &now
	}

	return nil
}

func (t *TaskModel) AfterFind(tx *gorm.DB) error {
	var files []FileModel
	if err := tx.Where("ref_id = ? AND ref_type = ?", t.ID, "task").Find(&files).Error; err == nil {
		t.Files = files
		if len(files) > 0 {
			if strings.HasPrefix(files[0].MimeType, "image/") {
				t.Cover = &files[0]
			}
		}
	}
	var count int64
	if err := tx.Model(&TaskCommentModel{}).Where("task_id = ?", t.ID).Count(&count).Error; err != nil {
		return err
	}

	t.CommentCount = int(count)

	if t.TaskAttibuteData != nil {
		var data TaskAttributeModel
		err := json.Unmarshal([]byte(*t.TaskAttibuteData), &data)
		if err != nil {
			return err
		}
		t.TaskAttribute = &data
		t.TaskAttibuteData = nil
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
