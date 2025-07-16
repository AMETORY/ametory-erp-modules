package treatment_queue

import (
	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/shared/models"
	"gorm.io/gorm"
)

// TreatmentQueueService is the service for treatment_queue model
type TreatmentQueueService struct {
	db  *gorm.DB
	ctx *context.ERPContext
}

// NewTreatmentQueueService creates new TreatmentQueueService
func NewTreatmentQueueService(db *gorm.DB, ctx *context.ERPContext) *TreatmentQueueService {
	return &TreatmentQueueService{db: db, ctx: ctx}
}

// CreateTreatmentQueue creates a new treatment queue in the database.
//
// It takes a pointer to a TreatmentQueueModel as an argument and returns an error
// if the creation fails.
func (s *TreatmentQueueService) CreateTreatmentQueue(queue *models.TreatmentQueueModel) error {
	if err := s.db.Create(queue).Error; err != nil {
		return err
	}
	return nil
}

// GetTreatmentQueueByID retrieves a treatment queue by its ID from the database.
//
// It takes a string argument representing the treatment queue ID and returns a pointer
// to a TreatmentQueueModel and an error. If the retrieval fails, it returns an error.
func (s *TreatmentQueueService) GetTreatmentQueueByID(ID string) (*models.TreatmentQueueModel, error) {
	var queue models.TreatmentQueueModel
	if err := s.db.First(&queue, "id = ?", ID).Error; err != nil {
		return nil, err
	}
	return &queue, nil
}

// UpdateTreatmentQueue updates a treatment queue in the database.
//
// It takes a string argument representing the ID of the treatment queue and a pointer
// to a TreatmentQueueModel as an argument and returns an error if the update fails.
func (s *TreatmentQueueService) UpdateTreatmentQueue(ID string, queue *models.TreatmentQueueModel) error {
	if err := s.db.Model(&models.TreatmentQueueModel{}).Where("id = ?", ID).Updates(queue).Error; err != nil {
		return err
	}
	return nil
}

// DeleteTreatmentQueue deletes a treatment queue from the database.
//
// It takes a string argument representing the treatment queue ID and returns an error
// if the deletion fails.
func (s *TreatmentQueueService) DeleteTreatmentQueue(ID string) error {
	if err := s.db.Delete(&models.TreatmentQueueModel{}, "id = ?", ID).Error; err != nil {
		return err
	}
	return nil
}
