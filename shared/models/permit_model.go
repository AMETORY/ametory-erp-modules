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

type PermitUploadedDocument struct {
	shared.BaseModel
	PermitRequestID       *string        `gorm:"index" json:"permit_request_id,omitempty"`
	PermitRequest         *PermitRequest `gorm:"foreignKey:PermitRequestID" json:"permit_request,omitempty"`
	FileName              string         `json:"file_name,omitempty"`
	FileURL               string         `json:"file_url,omitempty"`
	UploadedByID          *string        `gorm:"type:char(36);index" json:"uploaded_by_id,omitempty"`
	UploadedBy            *UserModel     `gorm:"foreignKey:UploadedByID;constraint:OnDelete:CASCADE;" json:"uploaded_by,omitempty"`
	PermitRequirementCode *string        `gorm:"type:varchar(255);index" json:"permit_requirement_code,omitempty"`
}

func (p *PermitUploadedDocument) BeforeCreate(tx *gorm.DB) (err error) {
	if p.ID == "" {
		p.ID = uuid.New().String()
	}
	return
}

type PermitFieldDefinition struct {
	shared.BaseModel
	PermitTypeID string           `gorm:"index" json:"permit_type_id"`
	PermitType   PermitType       `gorm:"foreignKey:PermitTypeID" json:"permit_type"`
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

type PermitDynamicRequestData struct {
	shared.BaseModel
	PermitRequestID string           `gorm:"index" json:"permit_request_id,omitempty"`
	Data            *json.RawMessage `gorm:"type:json" json:"data,omitempty"`
}

func (p *PermitDynamicRequestData) BeforeCreate(tx *gorm.DB) (err error) {
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

type PermitRequest struct {
	shared.BaseModel
	Code                 string                    `gorm:"uniqueIndex" json:"code,omitempty"`
	PermitTypeID         string                    `json:"permit_type_id,omitempty"`
	PermitType           PermitType                `gorm:"foreignKey:PermitTypeID;constraint:OnDelete:CASCADE;" json:"permit_type,omitempty"`
	CitizenID            string                    `json:"citizen_id,omitempty"`
	Citizen              Citizen                   `gorm:"foreignKey:CitizenID;constraint:OnDelete:CASCADE;" json:"citizen,omitempty"`
	Status               string                    `json:"status,omitempty"`
	SubmittedAt          time.Time                 `json:"submitted_at,omitempty"`
	ApprovedAt           *time.Time                `json:"approved_at,omitempty"`
	CurrentStep          int                       `json:"current_step"`
	RegisterNumber       string                    `gorm:"type:varchar(255)" json:"register_number,omitempty"`
	CurrentStepRoles     []RoleModel               `gorm:"many2many:permit_request_current_step_roles;constraint:OnDelete:CASCADE;" json:"current_step_roles,omitempty"`
	ApprovalLogs         []PermitApprovalLog       `gorm:"foreignKey:PermitRequestID;constraint:OnDelete:CASCADE;" json:"approval_logs,omitempty"`
	Documents            []PermitUploadedDocument  `gorm:"foreignKey:PermitRequestID;constraint:OnDelete:CASCADE;" json:"documents,omitempty"`
	SubDistrictID        *string                   `json:"sub_district_id,omitempty"`
	SubDistrict          *SubDistrict              `gorm:"foreignKey:SubDistrictID;constraint:OnDelete:CASCADE;" json:"sub_district,omitempty"`
	DynamicRequestData   *PermitDynamicRequestData `gorm:"foreignKey:PermitRequestID;constraint:OnDelete:CASCADE;" json:"dynamic_request_data,omitempty"`
	FinalPermitDocuments []FinalPermitDocument     `gorm:"foreignKey:PermitRequestID;constraint:OnDelete:CASCADE;" json:"final_permit_documents,omitempty"`
}

func (p *PermitRequest) BeforeCreate(tx *gorm.DB) (err error) {
	if p.ID == "" {
		p.ID = uuid.New().String()
	}
	return
}

type PermitApprovalLog struct {
	shared.BaseModel
	PermitRequestID string        `gorm:"index" json:"permit_request_id,omitempty"`
	PermitRequest   PermitRequest `gorm:"foreignKey:PermitRequestID" json:"permit_request,omitempty"`
	Step            string        `json:"step,omitempty"`
	StepRoleID      *string       `json:"step_role_id,omitempty"`
	StepRole        *RoleModel    `gorm:"foreignKey:StepRoleID" json:"step_role,omitempty"`
	Status          string        `json:"status,omitempty"`
	ApprovedBy      *string       `json:"approved_by,omitempty"`
	ApprovedByUser  *UserModel    `gorm:"foreignKey:ApprovedBy" json:"approved_by_user,omitempty"`
	ApprovedAt      time.Time     `json:"approved_at,omitempty"`
	Note            string        `json:"note,omitempty"`
}

func (p *PermitApprovalLog) BeforeCreate(tx *gorm.DB) (err error) {
	if p.ID == "" {
		p.ID = uuid.New().String()
	}
	return
}

type FinalPermitDocument struct {
	shared.BaseModel
	PermitRequestID string        `gorm:"index" json:"permit_request_id,omitempty"`
	PermitRequest   PermitRequest `gorm:"foreignKey:PermitRequestID" json:"permit_request,omitempty"`
	FileName        string        `json:"file_name,omitempty"`
	FileURL         string        `json:"file_url,omitempty"`
	GeneratedAt     time.Time     `json:"generated_at,omitempty"`
	GeneratedBy     *string       `json:"generated_by,omitempty"`
	GeneratedByUser *UserModel    `gorm:"foreignKey:GeneratedBy" json:"generated_by_user,omitempty"`
}

func (p *FinalPermitDocument) BeforeCreate(tx *gorm.DB) (err error) {
	if p.ID == "" {
		p.ID = uuid.New().String()
	}
	return
}
