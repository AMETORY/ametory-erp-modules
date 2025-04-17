package models

import (
	"time"

	"github.com/AMETORY/ametory-erp-modules/shared"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AssetModel struct {
	shared.BaseModel
	CompanyID                        *string                 `json:"company_id,omitempty"`
	Company                          *CompanyModel           `gorm:"foreignKey:CompanyID;constraint:OnDelete:CASCADE;" json:"company,omitempty"`
	UserID                           *string                 `json:"user_id,omitempty"`
	User                             *UserModel              `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE;" json:"user,omitempty"`
	Name                             string                  `json:"name"`
	AssetNumber                      string                  `json:"asset_number"`
	Date                             time.Time               `json:"date"`
	DepreciationMethod               string                  `json:"depreciation_method"`
	LifeTime                         float64                 `json:"life_time"`
	IsDepreciationAsset              bool                    `json:"is_depreciation_asset"`
	Description                      string                  `json:"description,omitempty"`
	AcquisitionCost                  float64                 `json:"acquisition_cost"`
	AccountFixedAssetID              *string                 `gorm:"size:36" json:"account_fixed_asset_id,omitempty" `
	AccountFixedAsset                *AccountModel           `gorm:"foreignKey:AccountFixedAssetID;constraint:OnDelete:CASCADE;" json:"account_fixed_asset,omitempty"`
	AccountCurrentAssetID            *string                 `gorm:"size:36" json:"account_current_asset_id,omitempty" `
	AccountCurrentAsset              *AccountModel           `gorm:"foreignKey:AccountCurrentAssetID;constraint:OnDelete:CASCADE;" json:"account_current_asset,omitempty"`
	AccountDepreciationID            *string                 `gorm:"size:36" json:"account_depreciation_id,omitempty" `
	AccountDepreciation              *AccountModel           `gorm:"foreignKey:AccountDepreciationID;constraint:OnDelete:CASCADE;" json:"account_depreciation,omitempty"`
	AccountAccumulatedDepreciationID *string                 `gorm:"size:36" json:"account_accumulated_depreciation_id,omitempty" `
	AccountAccumulatedDepreciation   *AccountModel           `gorm:"foreignKey:AccountAccumulatedDepreciationID;constraint:OnDelete:CASCADE;" json:"account_accumulated_depreciation,omitempty"`
	SalvageValue                     float64                 `json:"salvage_value"`
	BookValue                        float64                 `json:"book_value"`
	Status                           string                  `json:"status" gorm:"default:'DRAFT'"` // PENDING', 'ACTIVE', 'DISPOSED
	IsMonthly                        bool                    `json:"is_monthly"`
	Depreciations                    []DepreciationCostModel `json:"depreciations,omitempty" gorm:"-"`
	DepreciationMethodLabel          string                  `json:"depreciation_method_label,omitempty" gorm:"-"`
}

func (a *AssetModel) BeforeCreate(tx *gorm.DB) (err error) {
	if a.ID == "" {
		tx.Statement.SetColumn("id", uuid.New().String())
	}
	return
}

func (AssetModel) TableName() string {
	return "assets"
}

type DepreciationCostModel struct {
	shared.BaseModel
	CompanyID     *string       `json:"company_id,omitempty"`
	Company       *CompanyModel `gorm:"foreignKey:CompanyID;constraint:OnDelete:CASCADE;" json:"company,omitempty"`
	UserID        *string       `json:"user_id,omitempty"`
	User          *UserModel    `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE;" json:"user,omitempty"`
	AssetID       *string       `gorm:"size:36" json:"asset_id,omitempty"`
	Asset         *AssetModel   `gorm:"foreignKey:AssetID;constraint:OnDelete:CASCADE;" json:"asset,omitempty"`
	Amount        float64       `json:"amount,omitempty"`
	Period        int           `json:"period,omitempty"`
	Month         int           `json:"month,omitempty"`
	ExecutedAt    *time.Time    `json:"executed_at,omitempty"`
	TransactionID *string       `json:"transaction_id,omitempty"`
	Status        string        `json:"status,omitempty" gorm:"default:'PENDING'"` // 'PENDING', 'ACTIVE', 'DONE'
}

func (d *DepreciationCostModel) BeforeCreate(tx *gorm.DB) (err error) {
	if d.ID == "" {
		tx.Statement.SetColumn("id", uuid.New().String())
	}
	return
}

func (DepreciationCostModel) TableName() string {
	return "depreciation_costs"
}
