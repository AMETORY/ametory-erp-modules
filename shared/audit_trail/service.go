package audit_trail

import (
	"gorm.io/gorm"
)

type AuditTrailService struct {
	db *gorm.DB
}

func NewAuditTrailService(db *gorm.DB) *AuditTrailService {
	return &AuditTrailService{db: db}
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
