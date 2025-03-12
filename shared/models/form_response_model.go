package models

import (
	"github.com/AMETORY/ametory-erp-modules/shared"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type FormResponseModel struct {
	shared.BaseModel
	FormID   string        `json:"form_id" gorm:"index"`
	Form     FormModel     `json:"form,omitempty" gorm:"foreignKey:FormID;constraint:OnDelete:CASCADE;"`
	Sections []FormSection `json:"sections" gorm:"-"`
	Metadata string        `json:"metadata,omitempty" gorm:"type:JSON"`
	Data     string        `json:"Data" gorm:"type:JSON"`
	RefID    string        `json:"ref_id,omitempty" gorm:"type:char(36);index"`
	RefType  string        `json:"ref_type,omitempty" gorm:"type:varchar(255);index"`
}

func (f *FormResponseModel) TableName() string {
	return "form_responses"
}

func (f *FormResponseModel) BeforeCreate(tx *gorm.DB) (err error) {
	// Add custom logic before creating a FormResponseModel record
	if f.ID == "" {
		tx.Statement.SetColumn("id", uuid.New().String())
	}
	return nil
}

type FormSectionResponse struct {
	ID     string              `json:"id"`
	Fields []FormFieldResponse `json:"fields"`
}

type FormFieldResponse struct {
	ID              string        `json:"id"`
	Value           interface{}   `json:"value"`
	Filename        interface{}   `json:"filename"`
	StorageProvider string        `json:"storage_provider"`
	URL             string        `json:"url"`
	IsMulti         bool          `json:"is_multi"`
	Type            FormFieldType `json:"type"`
}
