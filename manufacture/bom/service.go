package bom

import (
	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/shared/models"
	"gorm.io/gorm"
)

// BuildOfMaterialService is a service that provides operations for managing Bill of Materials (BOMs)
type BuildOfMaterialService struct {
	db  *gorm.DB
	ctx *context.ERPContext
}

// NewBuildOfMaterialService creates a new instance of BuildOfMaterialService with the given database connection and context.
func NewBuildOfMaterialService(db *gorm.DB, ctx *context.ERPContext) *BuildOfMaterialService {
	return &BuildOfMaterialService{
		db:  db,
		ctx: ctx,
	}
}

// Migrate creates the database table for BillOfMaterial and its related tables.
func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(&models.BillOfMaterial{}, &models.BOMItem{}, &models.BOMOperation{})
}

// CreateBillOfMaterial creates a new BillOfMaterial with the given details.
func (s *BuildOfMaterialService) CreateBillOfMaterial(bom *models.BillOfMaterial) error {
	return s.db.Create(bom).Error
}

// UpdateBillOfMaterial updates the BillOfMaterial with the given ID with the given details.
func (s *BuildOfMaterialService) UpdateBillOfMaterial(id string, bom *models.BillOfMaterial) error {
	return s.db.Model(&models.BillOfMaterial{}).Where("id = ?", id).Updates(bom).Error
}

// DeleteBillOfMaterial deletes the BillOfMaterial with the given ID.
func (s *BuildOfMaterialService) DeleteBillOfMaterial(id string) error {
	return s.db.Where("id = ?", id).Delete(&models.BillOfMaterial{}).Error
}

// GetBillOfMaterial returns the BillOfMaterial with the given ID.
func (s *BuildOfMaterialService) GetBillOfMaterial(id string) (*models.BillOfMaterial, error) {
	var bom models.BillOfMaterial
	if err := s.db.Where("id = ?", id).First(&bom).Error; err != nil {
		return nil, err
	}
	return &bom, nil
}

// GetAllBillOfMaterials returns all BillOfMaterials.
func (s *BuildOfMaterialService) GetAllBillOfMaterials() ([]models.BillOfMaterial, error) {
	var boms []models.BillOfMaterial
	if err := s.db.Find(&boms).Error; err != nil {
		return nil, err
	}
	return boms, nil
}

// AddItem adds a new BOMItem to the BillOfMaterial with the given BOM ID.
func (s *BuildOfMaterialService) AddItem(bomID string, item *models.BOMItem) error {
	item.BOMID = bomID
	return s.db.Create(item).Error
}

// UpdateItem updates the BOMItem with the given item ID.
func (s *BuildOfMaterialService) UpdateItem(itemID string, item *models.BOMItem) error {
	return s.db.Model(&models.BOMItem{}).Where("id = ?", itemID).Updates(item).Error
}

// DeleteItem deletes the BOMItem with the given item ID.
func (s *BuildOfMaterialService) DeleteItem(itemID string) error {
	return s.db.Where("id = ?", itemID).Delete(&models.BOMItem{}).Error
}

// AddOperation adds a new BOMOperation to the BillOfMaterial with the given BOM ID.
func (s *BuildOfMaterialService) AddOperation(bomID string, operation *models.BOMOperation) error {
	operation.BOMID = bomID
	return s.db.Create(operation).Error
}

// UpdateOperation updates the BOMOperation with the given operation ID.
func (s *BuildOfMaterialService) UpdateOperation(operationID string, operation *models.BOMOperation) error {
	return s.db.Model(&models.BOMOperation{}).Where("id = ?", operationID).Updates(operation).Error
}

// DeleteOperation deletes the BOMOperation with the given operation ID.
func (s *BuildOfMaterialService) DeleteOperation(operationID string) error {
	return s.db.Where("id = ?", operationID).Delete(&models.BOMOperation{}).Error
}
