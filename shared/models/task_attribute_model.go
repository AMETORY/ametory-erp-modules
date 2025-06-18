package models

import (
	"encoding/json"

	"github.com/AMETORY/ametory-erp-modules/shared"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TaskAttributeModel struct {
	shared.BaseModel
	Title       string                    `json:"title"`
	Description string                    `json:"description"`
	Fields      []TaskAttributeFieldModel `json:"fields" gorm:"-"`
	Data        *string                   `json:"data,omitempty" gorm:"type:JSON"`
	CompanyID   *string                   `json:"company_id,omitempty"`
	Company     *CompanyModel             `gorm:"foreignKey:CompanyID;constraint:OnDelete:CASCADE" json:"company,omitempty"`
}

func (m *TaskAttributeModel) BeforeCreate(tx *gorm.DB) (err error) {
	if m.ID == "" {
		m.ID = uuid.New().String()
	}
	return
}

func (m *TaskAttributeModel) TableName() string {
	return "task_attributes"
}

func (m *TaskAttributeModel) AfterFind(tx *gorm.DB) (err error) {
	var data []TaskAttributeFieldModel
	if m.Data != nil {
		err = json.Unmarshal([]byte(*m.Data), &data)
		if err != nil {
			return err
		}
		m.Fields = data
		m.Data = nil
	}
	return
}

type TaskAttributeFieldModel struct {
	ID           string            `json:"id"`
	Label        string            `json:"label"`
	Type         FormFieldType     `json:"type"`
	Options      []FormFieldOption `json:"options"`
	Required     bool              `json:"required"`
	IsMulti      bool              `json:"is_multi"`
	IsPinned     bool              `json:"is_pinned"`
	Placeholder  string            `json:"placeholder"`
	DefaultValue string            `json:"default_value"`
	HelpText     string            `json:"help_text"`
	Disabled     bool              `json:"disabled"`
	Value        any               `json:"value"`
}
