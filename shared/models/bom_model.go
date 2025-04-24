package models

import (
	"github.com/AMETORY/ametory-erp-modules/shared"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type BillOfMaterial struct {
	shared.BaseModel
	Code       string         `json:"code"`       // e.g. BOM0001
	ProductID  string         `json:"product_id"` // reference to master product
	Version    float64        `json:"version"`
	Revision   float64        `json:"revision"`
	Status     string         `json:"status"` // e.g. Active, Inactive
	Items      []BOMItem      `json:"items" gorm:"foreignKey:BOMID"`
	Operations []BOMOperation `json:"operations" gorm:"foreignKey:BOMID"`
}

func (bom *BillOfMaterial) BeforeCreate(tx *gorm.DB) (err error) {
	if bom.ID == "" {
		tx.Statement.SetColumn("id", uuid.New().String())
	}
	return
}

type BOMItem struct {
	shared.BaseModel
	BOMID     string          `json:"bom_id"`
	BOM       *BillOfMaterial `json:"bom" gorm:"foreignKey:BOMID;constraint:OnDelete:CASCADE"`
	ProductID *string         `json:"product_id"` // reference to product master
	Product   *ProductModel   `json:"product" gorm:"foreignKey:ProductID;constraint:OnDelete:CASCADE"`
	ItemName  string          `json:"item_name"` // for denormalized UI
	Quantity  float64         `json:"quantity"`
	UnitID    string          `json:"unit_id"`
	Unit      *UnitModel      `json:"unit" gorm:"foreignKey:UnitID;constraint:OnDelete:CASCADE"`
}

func (bi *BOMItem) BeforeCreate(tx *gorm.DB) (err error) {
	if bi.ID == "" {
		tx.Statement.SetColumn("id", uuid.New().String())
	}
	return
}

type BOMOperation struct {
	shared.BaseModel
	BOMID        string          `json:"bom_id"`
	BOM          *BillOfMaterial `json:"bom" gorm:"foreignKey:BOMID;constraint:OnDelete:CASCADE"`
	Operation    string          `json:"operation"` // e.g. ASSEMBLY, INSPECTION
	WorkCenterID string          `json:"work_center_id"`
	WorkCenter   *WorkCenter     `json:"work_center" gorm:"foreignKey:WorkCenterID;constraint:OnDelete:CASCADE"`
}

func (bo *BOMOperation) BeforeCreate(tx *gorm.DB) (err error) {
	if bo.ID == "" {
		tx.Statement.SetColumn("id", uuid.New().String())
	}
	return
}
