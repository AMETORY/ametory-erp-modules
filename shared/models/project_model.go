package models

import (
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
	Members     []MemberModel `json:"members,omitempty" gorm:"many2many:project_members;constraint:OnDelete:CASCADE"`
	CreatedByID *string       `gorm:"type:char(36);index" json:"created_by_id,omitempty"`
	CreatedBy   *UserModel    `json:"created_by,omitempty" gorm:"foreignKey:CreatedByID;constraint:OnDelete:CASCADE"`
	CompanyID   *string       `gorm:"type:char(36);index" json:"company_id,omitempty"`
	Company     *CompanyModel `json:"company,omitempty" gorm:"foreignKey:CompanyID;constraint:OnDelete:CASCADE;"`
}

func (ProjectModel) TableName() string {
	return "projects"
}

func (p *ProjectModel) BeforeCreate(tx *gorm.DB) (err error) {
	// Add any custom logic before creating a ProjectModel
	p.ID = uuid.New().String()
	return nil
}

func (ColumnModel) TableName() string {
	return "columns"
}

func (c *ColumnModel) BeforeCreate(tx *gorm.DB) (err error) {
	// Add any custom logic before creating a ColumnModel
	c.ID = uuid.New().String()
	return nil
}

type ColumnModel struct {
	shared.BaseModel
	ProjectID string      `gorm:"type:char(36)" json:"project_id,omitempty"`
	Name      string      `gorm:"type:varchar(255)" json:"name,omitempty"`
	Icon      *string     `json:"icon,omitempty"`
	Order     int         `json:"order,omitempty"` // Urutan kolom
	Color     *string     `json:"color,omitempty"`
	Tasks     []TaskModel `json:"tasks,omitempty" gorm:"foreignKey:ColumnID"`
}
