package models

import (
	"time"

	"github.com/AMETORY/ametory-erp-modules/shared"
)

type CooperativeMemberModel struct {
	shared.BaseModel
	CompanyID              *string            `gorm:"size:30" json:"-" bson:"company_id,omitempty"`
	Name                   string             `json:"name"`
	MemberIDNumber         string             `json:"member_id_number" gorm:"member_id_number" sql:"member_id_number"`
	JoinDate               time.Time          `json:"join_date"`
	Active                 bool               `json:"active"`
	Email                  string             `json:"email"`
	Picture                string             `json:"picture"`
	PhoneNumber            string             `json:"phone_number"`
	Address                string             `json:"address"`
	City                   string             `json:"city"`
	ZipCode                string             `json:"zip_code"`
	Country                string             `json:"country"`
	ConnectedTo            *string            `json:"connected_to" `
	TotalSavings           float64            `json:"total_savings" gorm:"-"`
	TotalLoans             float64            `json:"total_loans" gorm:"-"`
	TotalRemainLoans       float64            `json:"total_remain_loans" gorm:"-"`
	TotalTransactions      float64            `json:"total_transactions" gorm:"-"`
	TotalDisbursement      float64            `json:"total_disbursement" gorm:"-"`
	NetSurplusTransactions []TransactionModel `gorm:"-"`
}

func (CooperativeMemberModel) TableName() string {
	return "cooperative_members"
}
