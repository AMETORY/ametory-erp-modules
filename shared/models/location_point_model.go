package models

import (
	"github.com/AMETORY/ametory-erp-modules/shared"
	"github.com/AMETORY/ametory-erp-modules/shared/indonesia_regional"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// LocationPointModel adalah model database untuk warehouse location
type LocationPointModel struct {
	shared.BaseModel
	Name        string                          `gorm:"not null;type:varchar(255)" json:"name,omitempty"`
	Description string                          `gorm:"type:text" json:"description,omitempty"`
	WarehouseID *string                         `gorm:"size:36" json:"warehouse_id,omitempty"`
	Warehouse   *WarehouseModel                 `gorm:"foreignKey:WarehouseID;constraint:OnDelete:CASCADE" json:"warehouse,omitempty"`
	Type        string                          `gorm:"type:varchar(20)" json:"type,omitempty"` // Warehouse, RegionalHub, Posko, etc
	Address     string                          `gorm:"type:varchar(255)" json:"address,omitempty"`
	Latitude    float64                         `json:"latitude" gorm:"type:decimal(10,8);"`
	Longitude   float64                         `json:"longitude" gorm:"type:decimal(11,8);"`
	ZipCode     *string                         `json:"zip_code,omitempty"`
	ProvinceID  *string                         `json:"province_id,omitempty" gorm:"type:char(2);index;constraint:OnDelete:SET NULL;"`
	Province    *indonesia_regional.RegProvince `gorm:"-" json:"province,omitempty"`
	RegencyID   *string                         `json:"regency_id,omitempty" gorm:"type:char(4);index;constraint:OnDelete:SET NULL;"`
	Regency     *indonesia_regional.RegRegency  `gorm:"-" json:"regency,omitempty"`
	DistrictID  *string                         `json:"district_id,omitempty" gorm:"type:char(6);index;constraint:OnDelete:SET NULL;"`
	District    *indonesia_regional.RegDistrict `gorm:"-" json:"district,omitempty"`
	VillageID   *string                         `json:"village_id,omitempty" gorm:"type:char(10);index;constraint:OnDelete:SET NULL;"`
	Village     *indonesia_regional.RegVillage  `gorm:"-" json:"village,omitempty"`
}

func (m *LocationPointModel) BeforeCreate(tx *gorm.DB) (err error) {
	if m.ID == "" {
		tx.Statement.SetColumn("id", uuid.New().String())
	}
	return
}

func (m *LocationPointModel) TableName() string {
	return "location_points"
}

func (m *LocationPointModel) AfterFind(tx *gorm.DB) (err error) {
	if m.ProvinceID != nil {
		prov := indonesia_regional.GetProvince(tx, *m.ProvinceID)
		m.Province = &prov
	}
	if m.RegencyID != nil {
		reg := indonesia_regional.GetRegency(tx, *m.RegencyID)
		m.Regency = &reg
	}
	if m.DistrictID != nil {
		dis := indonesia_regional.GetDistrict(tx, *m.DistrictID)
		m.District = &dis
	}
	if m.VillageID != nil {
		vil := indonesia_regional.GetVillage(tx, *m.VillageID)
		m.Village = &vil
	}
	return
}
