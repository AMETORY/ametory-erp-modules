package models

import (
	"time"

	"gorm.io/gorm"

	"github.com/AMETORY/ametory-erp-modules/shared"
	"github.com/google/uuid"
)

type MedicineModel struct {
	shared.BaseModel
	Name        string         `gorm:"type:varchar(255);not null" json:"name"`
	Barcode     string         `gorm:"type:varchar(255);not null" json:"barcode"`
	PharmacyID  string         `gorm:"type:char(36);index" json:"pharmacy_id"`
	Pharmacy    PharmacyModel  `gorm:"foreignKey:PharmacyID" json:"pharmacy"`
	Stock       float64        `gorm:"type:float;default:0" json:"stock"`
	Unit        string         `gorm:"type:varchar(255);default:'pcs'" json:"unit"`
	IsGeneric   bool           `gorm:"default:false" json:"is_generic"`
	IsRecipe    bool           `gorm:"default:false" json:"is_recipe"`
	Price       float64        `gorm:"type:float;default:0" json:"price"`
	PriceBuy    float64        `gorm:"type:float;default:0" json:"price_buy"`
	CreatedAt   time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at"`
	IsAvailable bool           `gorm:"type:boolean;default:true" json:"is_available"`
	Note        string         `gorm:"type:text;default:null" json:"note"`
}

func (MedicineModel) TableName() string {
	return "medicines"
}

func (m *MedicineModel) BeforeCreate(tx *gorm.DB) error {
	// Add logic before creating a medicine record if needed
	if m.ID == "" {
		tx.Statement.SetColumn("id", uuid.New().String())
	}
	return nil
}

type PharmacyModel struct {
	shared.BaseModel
	Name        string `gorm:"type:varchar(255);not null" json:"name"`
	Barcode     string `gorm:"type:varchar(255);not null" json:"barcode"`
	IsAvailable bool   `gorm:"type:boolean;default:true" json:"is_available"`
	Note        string `gorm:"type:text;default:null" json:"note"`
}

func (PharmacyModel) TableName() string {
	return "pharmacies"
}

func (p *PharmacyModel) BeforeCreate(tx *gorm.DB) error {
	// Add logic before creating a pharmacy record if needed
	if p.ID == "" {
		tx.Statement.SetColumn("id", uuid.New().String())
	}
	return nil
}
