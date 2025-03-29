package models

import (
	"time"

	"github.com/AMETORY/ametory-erp-modules/shared"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CompanyModel struct {
	shared.BaseModel
	Name                  string           `json:"name"`
	Logo                  string           `json:"logo"`
	Cover                 string           `json:"cover"`
	LegalEntity           string           `json:"legal_entity"`
	Email                 string           `json:"email"`
	Phone                 string           `json:"phone"`
	Fax                   string           `json:"fax"`
	Address               string           `json:"address"`
	ContactPerson         string           `json:"contact_person"`
	ContactPersonPosition string           `json:"contact_person_position"`
	TaxPayerNumber        string           `json:"tax_payer_number,omitempty"`
	UserID                *string          `json:"user_id,omitempty" gorm:"constraint:OnDelete:CASCADE;"`
	User                  *UserModel       `json:"user,omitempty" gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE;"`
	Status                string           `json:"status" gorm:"type:VARCHAR(20);DEFAULT:'ACTIVE'"`
	EmployeeActiveCount   int64            `json:"employee_active_count,omitempty"`
	EmployeeResignCount   int64            `json:"employee_resign_count,omitempty"`
	EmployeePendingCount  int64            `json:"employee_pending_count,omitempty"`
	Merchants             []MerchantModel  `json:"merchants,omitempty" gorm:"-"`
	Warehouses            []WarehouseModel `json:"warehouses,omitempty" gorm:"-"`
	EmailVerifiedAt       *time.Time       `gorm:"index" json:"email_verified_at,omitempty"`
	Users                 []UserModel      `gorm:"many2many:user_companies;" json:"users,omitempty"`
	BankAccount           *string          `json:"bank_account" `
	BankCode              *string          `json:"bank_code" `
	BeneficiaryName       *string          `json:"beneficiary_name" `
	IsCooperation         bool             `json:"is_cooperation" `
	CompanyCategoryID     *string          `json:"company_category_id,omitempty"`
	CompanyCategory       *CompanyCategory `json:"company_category,omitempty" gorm:"foreignKey:CompanyCategoryID;constraint:OnDelete:SET NULL;"`
	CustomCategory        *string          `json:"custom_category_id,omitempty"`
	ZipCode               *string          `json:"zip_code,omitempty"`
	ProvinceID            *string          `json:"province_id,omitempty" gorm:"type:char(2);index;constraint:OnDelete:SET NULL;"`
	RegencyID             *string          `json:"regency_id,omitempty" gorm:"type:char(4);index;constraint:OnDelete:SET NULL;"`
	DistrictID            *string          `json:"district_id,omitempty" gorm:"type:char(6);index;constraint:OnDelete:SET NULL;"`
	VillageID             *string          `json:"village_id,omitempty" gorm:"type:char(10);index;constraint:OnDelete:SET NULL;"`
}

func (CompanyModel) TableName() string {
	return "companies"
}

func (c *CompanyModel) BeforeCreate(tx *gorm.DB) error {
	if c.ID == "" {
		tx.Statement.SetColumn("id", uuid.New().String())
	}
	return nil
}

type CompanyCategory struct {
	shared.BaseModel
	Name          string         `json:"name" gorm:"uniqueIndex"`
	IsCooperative bool           `json:"is_cooperative" gorm:"default:false"`
	SectorID      *string        `json:"sector_id,omitempty" gorm:"type:char(36);index;constraint:OnDelete:SET NULL;"`
	CompanySector *CompanySector `json:"company_sector,omitempty" gorm:"foreignKey:SectorID;constraint:OnDelete:SET NULL;"`
}

type CompanySector struct {
	shared.BaseModel
	Name          string            `json:"name" gorm:"uniqueIndex"`
	IsCooperative bool              `json:"is_cooperative" gorm:"default:false"`
	Categories    []CompanyCategory `json:"categories,omitempty" gorm:"foreignKey:SectorID;constraint:OnDelete:SET NULL;"`
}
