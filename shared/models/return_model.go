package models

import (
	"time"

	"github.com/AMETORY/ametory-erp-modules/shared"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ReturnModel struct {
	shared.BaseModel
	ReturnNumber string              `gorm:"type:varchar(255);not null" json:"return_number"`
	Description  string              `gorm:"type:varchar(255);not null" json:"description"`
	Date         time.Time           `json:"date,omitempty"`
	ReturnType   string              `gorm:"type:varchar(255);not null" json:"return_type"`
	RefID        string              `gorm:"type:varchar(255);not null" json:"ref_id"`
	CompanyID    *string             `json:"company_id,omitempty"`
	Company      *CompanyModel       `gorm:"foreignKey:CompanyID;constraint:OnDelete:CASCADE" json:"company,omitempty"`
	UserID       *string             `json:"user_id,omitempty"`
	User         *UserModel          `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"user,omitempty"`
	Reason       string              `gorm:"type:varchar(255);not null" json:"reason"`
	Notes        string              `gorm:"type:text" json:"notes"`
	Status       string              `gorm:"type:varchar(255);default:'DRAFT'" json:"status"`
	ReleasedAt   *time.Time          `json:"released_at,omitempty"`
	ReleasedByID *string             `json:"released_by_id,omitempty"`
	ReleasedBy   *UserModel          `gorm:"foreignKey:ReleasedByID;constraint:OnDelete:CASCADE" json:"released_by,omitempty"`
	Items        []ReturnItemModel   `gorm:"foreignKey:ReturnID;constraint:OnDelete:CASCADE" json:"items"`
	PurchaseRef  *PurchaseOrderModel `gorm:"-" json:"purchase_ref,omitempty"`
	SalesRef     *SalesModel         `gorm:"-" json:"sales_ref,omitempty"`
}

func (r *ReturnModel) BeforeCreate(tx *gorm.DB) (err error) {
	if r.ID == "" {
		tx.Statement.SetColumn("id", uuid.New().String())
	}
	return
}

func (r *ReturnModel) TableName() string {
	return "returns"
}

type ReturnItemModel struct {
	shared.BaseModel
	Description        string          `json:"description,omitempty"`
	Notes              string          `gorm:"type:text" json:"notes"`
	ReturnID           string          `gorm:"type:char(36);index" json:"return_id"`
	Return             *ReturnModel    `gorm:"foreignKey:ReturnID;constraint:OnDelete:CASCADE" json:"return"`
	ProductID          *string         `json:"product_id"`
	Product            *ProductModel   `gorm:"foreignKey:ProductID;constraint:OnDelete:CASCADE" json:"product"`
	VariantID          *string         `json:"variant_id"`
	Variant            *VariantModel   `gorm:"foreignKey:VariantID;constraint:OnDelete:CASCADE" json:"variant"`
	Quantity           float64         `gorm:"type:decimal(13,2)" json:"quantity"`
	OriginalQuantity   float64         `gorm:"type:decimal(13,2)" json:"original_quantity"`
	UnitPrice          float64         `gorm:"type:decimal(13,2)" json:"unit_price"`
	DiscountPercent    float64         `json:"discount_percent,omitempty"`
	DiscountAmount     float64         `json:"discount_amount,omitempty"`
	UnitID             *string         `json:"unit_id,omitempty"` // Relasi ke unit
	Unit               *UnitModel      `gorm:"foreignKey:UnitID;constraint:OnDelete:CASCADE" json:"unit,omitempty"`
	TaxID              *string         `json:"tax_id,omitempty"`
	TotalTax           float64         `json:"total_tax,omitempty"`
	SubtotalBeforeDisc float64         `json:"subtotal_before_disc,omitempty"`
	Tax                *TaxModel       `gorm:"foreignKey:TaxID;constraint:Restrict:SET NULL" json:"tax,omitempty"`
	WarehouseID        *string         `json:"warehouse_id,omitempty"`
	Warehouse          *WarehouseModel `gorm:"foreignKey:WarehouseID;constraint:OnDelete:CASCADE" json:"warehouse,omitempty"`
	Value              float64         `gorm:"not null;default:1" json:"value"`
	Total              float64         `json:"total,omitempty"`
	SubTotal           float64         `json:"sub_total,omitempty"`
}

func (ri *ReturnItemModel) BeforeCreate(tx *gorm.DB) (err error) {
	if ri.ID == "" {
		tx.Statement.SetColumn("id", uuid.New().String())
	}
	return
}

func (ri *ReturnItemModel) TableName() string {
	return "return_items"
}
