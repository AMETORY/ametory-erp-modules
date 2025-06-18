package models

import (
	"encoding/json"
	"time"

	"github.com/AMETORY/ametory-erp-modules/shared"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ProjectModel struct {
	shared.BaseModel
	Name        string        `gorm:"type:varchar(255)" json:"name,omitempty"`
	Description string        `json:"description,omitempty"`
	Deadline    *time.Time    `json:"deadline,omitempty"`
	Status      string        `json:"status,omitempty"` // e.g., "ongoing", "completed"
	Columns     []ColumnModel `json:"columns,omitempty" gorm:"foreignKey:ProjectID;constraint:OnDelete:CASCADE"`
	Tasks       []TaskModel   `json:"tasks,omitempty" gorm:"foreignKey:ProjectID;constraint:OnDelete:CASCADE"`
	Members     []MemberModel `json:"members,omitempty" gorm:"many2many:project_members;"`
	CreatedByID *string       `gorm:"type:char(36);index" json:"created_by_id,omitempty"`
	CreatedBy   *UserModel    `json:"created_by,omitempty"`
	CompanyID   *string       `gorm:"type:char(36);index" json:"company_id,omitempty"`
	Company     *CompanyModel `json:"company,omitempty" gorm:"foreignKey:CompanyID;constraint:OnDelete:CASCADE;"`
}

func (ProjectModel) TableName() string {
	return "projects"
}

func (p *ProjectModel) BeforeCreate(tx *gorm.DB) (err error) {
	// Add any custom logic before creating a ProjectModel

	if p.ID == "" {
		tx.Statement.SetColumn("id", uuid.New().String())
	}
	return nil
}

func (ColumnModel) TableName() string {
	return "columns"
}

func (c *ColumnModel) BeforeCreate(tx *gorm.DB) (err error) {
	// Add any custom logic before creating a ColumnModel

	if c.ID == "" {
		tx.Statement.SetColumn("id", uuid.New().String())
	}
	return nil
}

type ColumnModel struct {
	shared.BaseModel
	ProjectID  string         `gorm:"type:char(36)" json:"project_id,omitempty"`
	Name       string         `gorm:"type:varchar(255)" json:"name,omitempty"`
	Icon       *string        `json:"icon,omitempty"`
	Order      int            `json:"order,omitempty"` // Urutan kolom
	Color      *string        `json:"color,omitempty"`
	Tasks      []TaskModel    `json:"tasks,omitempty" gorm:"foreignKey:ColumnID"`
	CountTasks int64          `gorm:"-" json:"count_tasks,omitempty"`
	Actions    []ColumnAction `json:"actions,omitempty" gorm:"foreignKey:ColumnID"`
}

type ColumnAction struct {
	shared.BaseModel
	Name            string           `gorm:"type:varchar(255)" json:"name,omitempty"`
	ColumnID        string           `gorm:"type:char(36);index" json:"column_id,omitempty"`
	Column          *ColumnModel     `gorm:"foreignKey:ColumnID;constraint:OnDelete:CASCADE;" json:"column,omitempty"`
	Action          string           `json:"action,omitempty"`
	ActionTrigger   string           `json:"action_trigger,omitempty"`               // ActionTriggers the event that triggers the action. ex: move_in, move_out
	ActionData      *json.RawMessage `gorm:"type:JSON" json:"action_data,omitempty"` // ActionData the data that will be passed to the action. ex: task_id, task_name
	ActionValue     string           `json:"action_value,omitempty"`
	ActionValueType string           `json:"action_value_type,omitempty"`
	Status          string           `gorm:"type:varchar(50);default:DRAFT" json:"status,omitempty"`
	Files           []FileModel      `json:"files,omitempty" gorm:"-"`
	ActionHour      *string          `gorm:"type:varchar(50)" json:"action_hour,omitempty"`
	ActionStatus    string           `gorm:"type:varchar(50);default:READY" json:"action_status,omitempty"`
}

func (m *ColumnAction) AfterFind(tx *gorm.DB) error {
	tx.Model(&FileModel{}).Where("ref_id = ? AND ref_type = ?", m.ID, "column_action").Find(&m.Files)
	return nil
}

func (c *ColumnAction) BeforeCreate(tx *gorm.DB) (err error) {
	// Add any custom logic before creating a ColumnAction

	if c.ID == "" {
		tx.Statement.SetColumn("id", uuid.New().String())
	}
	return nil
}

type ScheduledMessage struct {
	To       string               `json:"to"`
	Message  string               `json:"message"`
	Duration time.Duration        `json:"duration"`
	Files    []FileModel          `json:"files"`
	Data     WhatsappMessageModel `json:"data"`
	Action   *ColumnAction        `json:"action"`
	Task     *TaskModel           `json:"task"`
}
