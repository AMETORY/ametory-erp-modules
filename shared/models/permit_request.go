package models

import (
	"encoding/json"
	"time"

	"github.com/AMETORY/ametory-erp-modules/shared"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

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
	RefID                string                    `json:"ref_id,omitempty"`
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
	StepOrder       int           `json:"step_order,omitempty"`
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
