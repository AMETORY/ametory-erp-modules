package models

import (
	"time"

	"github.com/AMETORY/ametory-erp-modules/shared"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type WithdrawalStatus string

const (
	WithdrawalStatusPending  WithdrawalStatus = "PENDING"
	WithdrawalStatusSuccess  WithdrawalStatus = "SUCCESS"
	WithdrawalStatusRejected WithdrawalStatus = "REJECTED"
	WithdrawalStatusCanceled WithdrawalStatus = "CANCELED"
)

type WithdrawalModel struct {
	shared.BaseModel
	Code               string                `json:"code,omitempty" gorm:"not null"`
	Total              float64               `json:"total,omitempty" gorm:"not null"`
	BankAccount        string                `json:"bank_account,omitempty" gorm:"not null"`
	BankCode           string                `json:"bank_code,omitempty" gorm:"not null"`
	BeneficiaryName    string                `json:"beneficiary_name,omitempty" gorm:"not null"`
	Status             WithdrawalStatus      `json:"status,omitempty" gorm:"type:varchar(50);not null;default:PENDING"`
	Remarks            string                `json:"remarks,omitempty" gorm:"type:text"`
	RequestDate        time.Time             `json:"request_date,omitempty"`
	DisbursementDate   *time.Time            `json:"disbursement_date,omitempty"`
	DisbursementMethod *string               `json:"disbursement_method,omitempty"`
	ApprovalDate       *time.Time            `json:"approval_date,omitempty"`
	ApprovalBy         *string               `json:"approval_by,omitempty" gorm:"type:char(36)"`
	ApprovalByUser     *AdminModel           `gorm:"foreignKey:ApprovalBy;references:ID" json:"approval_by_user,omitempty"`
	RejectedBy         *string               `json:"rejected_by,omitempty" gorm:"type:char(36)"`
	RejectedByUser     *AdminModel           `gorm:"foreignKey:RejectedBy;references:ID" json:"rejected_by_user,omitempty"`
	RequestedBy        *string               `json:"requested_by,omitempty" gorm:"type:char(36)"`
	RequestedByUser    *UserModel            `gorm:"foreignKey:RequestedBy;references:ID" json:"requested_by_user,omitempty"`
	MerchantID         *string               `json:"merchant_id,omitempty" gorm:"type:char(36)"`
	Merchant           *MerchantModel        `gorm:"foreignKey:MerchantID;references:ID" json:"merchant,omitempty"`
	Files              []FileModel           `gorm:"-" json:"files,omitempty"`
	Items              []WithdrawalItemModel `gorm:"foreignKey:WithdrawalID;references:ID;constraint:OnDelete:CASCADE" json:"withdrawal_items,omitempty"`
}

func (WithdrawalModel) TableName() string {
	return "withdrawals"
}

func (w *WithdrawalModel) BeforeCreate(tx *gorm.DB) (err error) {
	if w.ID == "" {
		w.ID = uuid.New().String()
	}
	return
}

func (w *WithdrawalModel) AfterFind(tx *gorm.DB) (err error) {
	var files []FileModel
	err = tx.Where("ref_id = ? and ref_type = ?", w.ID, "withdrawal").Order("created_at desc").Find(&files).Error
	if err != nil {
		return err
	}
	w.Files = files
	return
}

type WithdrawalItemModel struct {
	shared.BaseModel
	WithdrawalID string      `json:"withdrawal_id,omitempty" gorm:"type:char(36);index"`
	Amount       float64     `json:"amount,omitempty" gorm:"not null"`
	PosID        *string     `json:"pos_id,omitempty" gorm:"type:char(36);index"`
	Pos          *POSModel   `gorm:"foreignKey:PosID;references:ID;constraint:OnDelete:CASCADE" json:"pos,omitempty"`
	SalesID      *string     `json:"sales_id,omitempty" gorm:"type:char(36);index"`
	Sales        *SalesModel `gorm:"foreignKey:SalesID;references:ID;constraint:OnDelete:CASCADE" json:"sales,omitempty"`
}

func (wi WithdrawalItemModel) TableName() string {
	return "withdrawal_items"
}

func (wi *WithdrawalItemModel) BeforeCreate(tx *gorm.DB) (err error) {
	if wi.ID == "" {
		wi.ID = uuid.New().String()
	}
	return
}
