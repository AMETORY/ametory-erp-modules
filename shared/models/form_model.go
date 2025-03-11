package models

import (
	"encoding/json"
	"fmt"

	"github.com/AMETORY/ametory-erp-modules/shared"
	"github.com/AMETORY/ametory-erp-modules/utils"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type FormFieldType string

const (
	TextField       FormFieldType = "text"
	TextArea        FormFieldType = "textarea"
	RadioButton     FormFieldType = "radio"
	Checkbox        FormFieldType = "checkbox"
	Dropdown        FormFieldType = "dropdown"
	DatePicker      FormFieldType = "date"
	DateRangePicker FormFieldType = "date_range"
	NumberField     FormFieldType = "number"
	Currency        FormFieldType = "currency"
	EmailField      FormFieldType = "email"
	PasswordField   FormFieldType = "password"
	FileUpload      FormFieldType = "file"
	ToggleSwitch    FormFieldType = "toggle"
)

type FormModel struct {
	shared.BaseModel
	Code              string              `json:"code,omitempty"`
	Title             string              `json:"title,omitempty"`
	Description       string              `json:"description,omitempty"`
	Picture           *FileModel          `json:"picture,omitempty" gorm:"-"`
	Cover             *FileModel          `json:"cover,omitempty" gorm:"-"`
	Sections          []FormSection       `json:"sections,omitempty" gorm:"-"`
	SubmitURL         string              `json:"submit_url,omitempty"`
	Method            string              `json:"method,omitempty"`
	Headers           string              `json:"headers,omitempty" gorm:"type:JSON"`
	IsPublic          bool                `json:"is_public,omitempty"`
	Status            string              `json:"status,omitempty"`
	FormTemplateID    *string             `gorm:"type:char(36);index" json:"form_template_id,omitempty"`
	FormTemplate      *FormTemplate       `json:"form_template,omitempty" gorm:"foreignKey:FormTemplateID;constraint:OnDelete:CASCADE;"`
	CreatedByID       *string             `gorm:"type:char(36);index" json:"created_by_id,omitempty"`
	CreatedBy         *UserModel          `json:"created_by,omitempty" gorm:"foreignKey:CreatedByID;constraint:OnDelete:SET NULL;"`
	CreatedByMemberID *string             `gorm:"type:char(36);index" json:"created_by_member_id,omitempty"`
	CreatedByMember   *MemberModel        `json:"created_by_member,omitempty" gorm:"foreignKey:CreatedByMemberID;constraint:OnDelete:SET NULL;"`
	Responses         []FormResponseModel `json:"responses,omitempty" gorm:"foreignKey:FormID"`
	CompanyID         *string             `gorm:"type:char(36);index" json:"company_id,omitempty"`
	Company           *CompanyModel       `json:"company,omitempty" gorm:"foreignKey:CompanyID;constraint:OnDelete:CASCADE;"`
	ColumnID          *string             `gorm:"type:char(36);index" json:"column_id,omitempty"`
	Column            *ColumnModel        `json:"column,omitempty" gorm:"foreignKey:ColumnID;constraint:OnDelete:CASCADE;"`
	ProjectID         *string             `gorm:"type:char(36);index" json:"project_id,omitempty"`
	Project           *ProjectModel       `json:"project,omitempty" gorm:"foreignKey:ProjectID;constraint:OnDelete:CASCADE;"`
}

func (f *FormModel) BeforeCreate(tx *gorm.DB) error {
	if f.ID == "" {
		tx.Statement.SetColumn("id", uuid.New().String())
	}
	if f.Code == "" {
		f.Code = utils.RandString(7, true)
	}

	return nil
}

func (f *FormModel) AfterFind(tx *gorm.DB) (err error) {
	if f.FormTemplate != nil {
		if f.FormTemplate.Data != nil {
			var data []FormSection
			if err = json.Unmarshal([]byte(*f.FormTemplate.Data), &data); err != nil {
				return
			}
			f.Sections = data
		}
	}

	fmt.Println("GETTING FORM PICTURES")
	var formPicture FileModel
	if err := tx.Where("ref_id = ? AND ref_type = ?", f.ID, "form-picture").First(&formPicture).Error; err == nil {
		f.Picture = &formPicture
	}
	var formCover FileModel
	if err := tx.Where("ref_id = ? AND ref_type = ?", f.ID, "form-cover").First(&formCover).Error; err == nil {
		f.Cover = &formCover
	}

	return
}

func (f *FormModel) TableName() string {
	return "forms"
}

type FormSection struct {
	ID           string      `json:"id"`
	SectionTitle string      `json:"section_title"`
	Description  string      `json:"description"`
	Fields       []FormField `json:"fields"`
}

type FormField struct {
	ID           string            `json:"id"`
	Label        string            `json:"label"`
	Type         FormFieldType     `json:"type"`
	Options      []FormFieldOption `json:"options"`
	Required     bool              `json:"required"`
	IsMulti      bool              `json:"is_multi"`
	Placeholder  string            `json:"placeholder"`
	DefaultValue string            `json:"default_value"`
	HelpText     string            `json:"help_text"`
	Disabled     bool              `json:"disabled"`
}

type FormFieldOption struct {
	Label string `json:"label"`
	Value string `json:"value"`
}

type FormTemplate struct {
	shared.BaseModel
	Title             string        `json:"title,omitempty"`
	Sections          []FormSection `json:"sections,omitempty" gorm:"-"`
	Data              *string       `json:"data" gorm:"type:JSON"`
	CompanyID         *string       `gorm:"type:char(36);index" json:"company_id,omitempty"`
	Company           *CompanyModel `json:"company,omitempty" gorm:"foreignKey:CompanyID;constraint:OnDelete:CASCADE;"`
	CreatedByID       *string       `gorm:"type:char(36);index" json:"created_by_id,omitempty"`
	CreatedBy         *UserModel    `json:"created_by,omitempty" gorm:"foreignKey:CreatedByID;constraint:OnDelete:SET NULL;"`
	CreatedByMemberID *string       `gorm:"type:char(36);index" json:"created_by_member_id,omitempty"`
	CreatedByMember   *MemberModel  `json:"created_by_member,omitempty" gorm:"foreignKey:CreatedByMemberID;constraint:OnDelete:SET NULL;"`
}

func (f *FormTemplate) BeforeCreate(tx *gorm.DB) error {
	if f.ID == "" {
		tx.Statement.SetColumn("id", uuid.New().String())
	}
	if f.Data == nil {
		tx.Statement.SetColumn("data", "[]")
	}
	return nil
}

func (f *FormTemplate) AfterFind(tx *gorm.DB) error {
	if f.Data != nil {
		if err := json.Unmarshal([]byte(*f.Data), &f.Sections); err != nil {
			return err
		}
	}
	f.Data = nil

	return nil
}
