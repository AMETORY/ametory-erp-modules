package audit_trail

import (
	"github.com/AMETORY/ametory-erp-modules/context"
	"gorm.io/gorm"
)

type AuditTrailService struct {
	erpContext *context.ERPContext
	db         *gorm.DB
}

// NewAuditTrailService creates a new instance of AuditTrailService with the given ERP context.
// It will call Migrate() if SkipMigration is not set to true.
func NewAuditTrailService(erpContext *context.ERPContext) *AuditTrailService {
	if !erpContext.SkipMigration {
		Migrate(erpContext.DB)
	}
	return &AuditTrailService{erpContext: erpContext, db: erpContext.DB}
}

// Migrate applies the necessary database migrations for the audit trail model.
//
// It ensures that the underlying database schema is up to date with the
// current version of the AuditTrailModel.
//
// If the migration fails, the error is returned to the caller.
func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(&AuditTrailModel{})
}

// LogAction creates a new audit trail record with the provided details.
//
// It constructs an AuditTrailModel using the given userID, action, entity,
// entityID, and details, and inserts it into the database. If the insertion
// fails, the error is returned to the caller.
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

// GetAuditTrails retrieves the audit trail for a given entity and entity ID.
//
// It performs a query against the audit trail table with the given entity and
// entity ID, and returns the results. If the query fails, the error is returned to
// the caller.
//
// The returned slice is empty if no audit trail records match the query.
func (s *AuditTrailService) GetAuditTrails(entity string, entityID uint) ([]AuditTrailModel, error) {
	var auditTrails []AuditTrailModel
	err := s.db.Where("entity = ? AND entity_id = ?", entity, entityID).Find(&auditTrails).Error
	return auditTrails, err
}
