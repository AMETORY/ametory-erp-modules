package models

import (
	"encoding/json"
	"time"

	"github.com/AMETORY/ametory-erp-modules/shared"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type FieldType string

var (
	TEXT       FieldType = "TEXT"
	DATE       FieldType = "DATE"
	JSON       FieldType = "JSON"
	NUMBER     FieldType = "NUMBER"
	EMAIL      FieldType = "EMAIL"
	CHECKBOX   FieldType = "CHECKBOX"
	BOOLEAN    FieldType = "BOOLEAN"
	TEXTAREA   FieldType = "TEXTAREA"
	SELECT     FieldType = "SELECT"
	DATEPICKER FieldType = "DATEPICKER"
	FILE       FieldType = "FILE"
	CHECKBOXES FieldType = "CHECKBOXES"
	ANY        FieldType = "ANY"
	ARRAY      FieldType = "ARRAY"
)

type PermitType struct {
	shared.BaseModel
	Name               string                  `json:"name"`
	Slug               string                  `gorm:"type:varchar(255);uniqueIndex:slug_district" json:"slug"`
	Description        string                  `json:"description"`
	FieldDefinitions   []PermitFieldDefinition `gorm:"foreignKey:PermitTypeID" json:"field_definitions"`
	ApprovalFlow       []PermitApprovalFlow    `gorm:"foreignKey:PermitTypeID" json:"approval_flow"`
	PermitRequirements []PermitRequirement     `gorm:"many2many:permit_type_requirements;constraint:OnDelete:CASCADE;" json:"permit_requirements"`
	SubDistrictID      *string                 `gorm:"size:36;uniqueIndex:slug_district" json:"subdistrict_id"`
	SubDistrict        *SubDistrict            `gorm:"foreignKey:SubDistrictID" json:"subdistrict"`
	PermitTemplateID   *string                 `gorm:"type:varchar(36);index" json:"permit_template_id"`
	PermitTemplate     *PermitTemplate         `gorm:"foreignKey:PermitTemplateID" json:"permit_template"`
	TemplateConfig     *json.RawMessage        `json:"template_config"`
	BodyTemplate       string                  `json:"body_template"`
}

type PermitTypeRequirement struct {
	PermitTypeID        string            `gorm:"type:varchar(36);not null" json:"permit_type_id"`
	PermitType          PermitType        `gorm:"foreignKey:PermitTypeID" json:"permit_type"`
	PermitRequirementID string            `gorm:"type:varchar(36);not null" json:"permit_requirement_id"`
	PermitRequirement   PermitRequirement `gorm:"foreignKey:PermitRequirementID" json:"permit_requirement"`
	IsMandatory         bool              `json:"is_mandatory" gorm:"default:false"`
}

func (p *PermitType) BeforeCreate(tx *gorm.DB) (err error) {
	if p.ID == "" {
		p.ID = uuid.New().String()
	}
	return
}

type PermitRequirement struct {
	shared.BaseModel
	Name          string       `gorm:"type:varchar(255)" json:"name"`
	Description   string       `json:"description"`
	Code          string       `gorm:"type:varchar(255);index" json:"code"`
	SubDistrictID *string      `gorm:"size:36" json:"subdistrict_id"`
	SubDistrict   *SubDistrict `gorm:"foreignKey:SubDistrictID" json:"subdistrict"`
	IsMandatory   bool         `json:"is_mandatory" gorm:"-"`
}

func (p *PermitRequirement) BeforeCreate(tx *gorm.DB) (err error) {
	if p.ID == "" {
		p.ID = uuid.New().String()
	}
	return
}

type PermitFieldDefinition struct {
	shared.BaseModel
	PermitTypeID *string          `gorm:"index;constraint:OnDelete:CASCADE" json:"permit_type_id,omitempty"`
	PermitType   *PermitType      `gorm:"foreignKey:PermitTypeID" json:"permit_type,omitempty"`
	FieldKey     string           `gorm:"index" json:"field_key"`
	FieldLabel   string           `json:"field_label"`
	FieldType    FieldType        `json:"field_type"`
	IsRequired   bool             `json:"is_required"`
	Order        int              `json:"order"`
	Options      *json.RawMessage `gorm:"type:json" json:"options"`
}

type PermitTemplateConfig struct {
	TemplateName  string `json:"template_name"`
	IncludeLogo   bool   `json:"include_logo"`
	LogoPosition  string `json:"logo_position"`
	Logo          string `json:"logo"`
	HeaderText    string `json:"header_text"`
	HeaderAddress string `json:"header_address"`
	SignatureText string `json:"signature_text"`
	FooterText    string `json:"footer_text"`
}

func (p *PermitFieldDefinition) BeforeCreate(tx *gorm.DB) (err error) {
	if p.ID == "" {
		p.ID = uuid.New().String()
	}
	return
}

type PermitApprovalFlow struct {
	shared.BaseModel
	PermitTypeID string      `gorm:"index" json:"permit_type_id,omitempty"`
	PermitType   *PermitType `gorm:"foreignKey:PermitTypeID" json:"permit_type,omitempty"`
	StepOrder    int         `json:"step_order,omitempty"`
	Roles        []RoleModel `gorm:"many2many:permit_approval_flow_roles;constraint:OnDelete:CASCADE;" json:"roles,omitempty"`
	Description  string      `json:"description,omitempty"`
	ApprovalMode string      `gorm:"type:varchar(50);default:SINGLE" json:"approval_mode,omitempty"` // SINGLE, ALL
}

func (p *PermitApprovalFlow) BeforeCreate(tx *gorm.DB) (err error) {
	if p.ID == "" {
		p.ID = uuid.New().String()
	}
	return
}

type PermitApprovalDecision struct {
	shared.BaseModel
	PermitRequestID string        `gorm:"index" json:"permit_request_id,omitempty"`
	PermitRequest   PermitRequest `gorm:"foreignKey:PermitRequestID" json:"permit_request,omitempty"`
	StepOrder       int           `json:"step_order,omitempty"`
	ApprovedBy      *string       `json:"approved_by,omitempty"`
	ApprovedByUser  *UserModel    `gorm:"foreignKey:ApprovedBy;constraint:OnDelete:CASCADE;" json:"approved_by_user,omitempty"`
	ApprovedAt      time.Time     `json:"approved_at,omitempty"`
	Note            string        `json:"note,omitempty"`
	Decision        string        `gorm:"type:varchar(20);default:APPROVED" json:"decision,omitempty"` // APPROVED, REJECTED
}

func (p *PermitApprovalDecision) BeforeCreate(tx *gorm.DB) (err error) {
	if p.ID == "" {
		p.ID = uuid.New().String()
	}
	return
}
