package audit_trail

import (
	"github.com/AMETORY/ametory-erp-modules/shared"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AuditAction string

const (
	ActionCreate   AuditAction = "CREATE"
	ActionUpdate   AuditAction = "UPDATE"
	ActionDelete   AuditAction = "DELETE"
	ActionAccept   AuditAction = "ACCEPT"
	ActionReject   AuditAction = "REJECT"
	ActionComplete AuditAction = "COMPLETE"
)

type AuditTrailModel struct {
	shared.BaseModel
	UserID   string      `gorm:"not null"` // ID user yang melakukan aksi
	Action   AuditAction `gorm:"not null"` // Jenis aksi (CREATE, UPDATE, DELETE, dll.)
	Entity   string      `gorm:"not null"` // Entitas yang terlibat (misalnya, "OrderRequest")
	EntityID string      `gorm:"not null"` // ID entitas yang terlibat
	Details  string      `gorm:"type:json"`
}

func (a *AuditTrailModel) TableName() string {
	return "audit_trails"
}

func (a *AuditTrailModel) BeforeCreate(tx *gorm.DB) (err error) {
	if a.ID == "" {
		tx.Statement.SetColumn("id", uuid.New().String())
	}
	return
}
