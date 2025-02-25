package models

import (
	"github.com/AMETORY/ametory-erp-modules/shared"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ProjectModel struct {
	shared.BaseModel
	Name        string        `json:"name,omitempty"`
	Description string        `json:"description,omitempty"`
	Deadline    string        `json:"deadline,omitempty"`
	Status      string        `json:"status,omitempty"` // e.g., "ongoing", "completed"
	Columns     []ColumnModel `json:"columns,omitempty" gorm:"foreignKey:ProjectID"`
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
	ProjectID string  `gorm:"type:char(36)" json:"project_id,omitempty"`
	Name      string  `json:"name,omitempty"`
	Order     int     `json:"order,omitempty"` // Urutan kolom
	Color     *string `json:"color,omitempty"`
}
