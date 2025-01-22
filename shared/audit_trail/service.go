package audit_trail

import (
	"github.com/AMETORY/ametory-erp-modules/context"
	"gorm.io/gorm"
)

type AuditTrailService struct {
	erpContext *context.ERPContext
	db         *gorm.DB
}

func NewAuditTrailService(erpContext *context.ERPContext) *AuditTrailService {
	if !erpContext.SkipMigration {
		Migrate(erpContext.DB)
	}
	return &AuditTrailService{erpContext: erpContext, db: erpContext.DB}
}

func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(&AuditTrailModel{})
}

func (s *AuditTrailService) LogAction(userID string, action AuditAction, entity string, entityID, details string) error {
	auditTrail := AuditTrailModel{
		UserID:   userID,
		Action:   action,
		Entity:   entity,
		EntityID: entityID,
		Details:  details,
	}
	return s.db.Create(&auditTrail).Error
}

func (s *AuditTrailService) GetAuditTrails(entity string, entityID uint) ([]AuditTrailModel, error) {
	var auditTrails []AuditTrailModel
	err := s.db.Where("entity = ? AND entity_id = ?", entity, entityID).Find(&auditTrails).Error
	return auditTrails, err
}
