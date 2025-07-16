package pharmacy

import (
	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/shared/models"
	"gorm.io/gorm"
)

// PharmacyService is the service for Pharmacy model
type PharmacyService struct {
	db  *gorm.DB
	ctx *context.ERPContext
}

// NewPharmacyService will create new PharmacyService
func NewPharmacyService(db *gorm.DB, ctx *context.ERPContext) *PharmacyService {
	return &PharmacyService{db: db, ctx: ctx}
}

// CreatePharmacy will create new Pharmacy
//
// This method will create new Pharmacy with the given payload.
//
// Example:
//
//	payload := &models.PharmacyModel{
//	  Name: "Pharmacy 1",
//	  Address: "Jl. example 1",
//	  City: "Jakarta",
//	  Province: "DKI Jakarta",
//	  Country: "Indonesia",
//	}
//
//	err := pharmacyService.CreatePharmacy(payload)
//	if err != nil {
//	  panic(err)
//	}
func (s *PharmacyService) CreatePharmacy(payload *models.PharmacyModel) error {
	return s.db.Create(payload).Error
}

// GetPharmacyByID will get Pharmacy with the given ID
//
// This method will get Pharmacy with the given ID.
//
// Example:
//
//	id := 1
//	pharmacy, err := pharmacyService.GetPharmacyByID(id)
//	if err != nil {
//	  panic(err)
//	}
func (s *PharmacyService) GetPharmacyByID(id uint) (*models.PharmacyModel, error) {
	var pharmacy models.PharmacyModel
	err := s.db.Preload("CreatedUser").Preload("UpdatedUser").First(&pharmacy, id).Error
	return &pharmacy, err
}

// GetPharmacyList will get list of Pharmacy with the given parameter
//
// This method will get list of Pharmacy with the given parameter.
//
// Example:
//
//	page := 1
//	pageSize := 10
//	name := "Pharmacy"
//	pharmacyList, err := pharmacyService.GetPharmacyList(page, pageSize, name)
//	if err != nil {
//	  panic(err)
//	}
func (s *PharmacyService) GetPharmacyList(page int, pageSize int, name string) ([]*models.PharmacyModel, error) {
	var pharmacyList []*models.PharmacyModel
	err := s.db.Preload("CreatedUser").Preload("UpdatedUser").
		Offset((page-1)*pageSize).
		Limit(pageSize).
		Where("name LIKE ?", "%"+name+"%").
		Find(&pharmacyList).Error
	return pharmacyList, err
}

// UpdatePharmacy will update Pharmacy with the given ID and payload
//
// This method will update Pharmacy with the given ID and payload.
//
// Example:
//
//	id := 1
//	payload := &models.PharmacyModel{
//	  Name: "Pharmacy 1",
//	  Address: "Jl. example 1",
//	  City: "Jakarta",
//	  Province: "DKI Jakarta",
//	  Country: "Indonesia",
//	}
//
//	err := pharmacyService.UpdatePharmacy(id, payload)
//	if err != nil {
//	  panic(err)
//	}
func (s *PharmacyService) UpdatePharmacy(id uint, payload *models.PharmacyModel) error {
	return s.db.Model(&models.PharmacyModel{}).Where("id = ?", id).
		Updates(payload).Error
}

// DeletePharmacy will delete Pharmacy with the given ID
//
// This method will delete Pharmacy with the given ID.
//
// Example:
//
//	id := 1
//	err := pharmacyService.DeletePharmacy(id)
//	if err != nil {
//	  panic(err)
//	}
func (s *PharmacyService) DeletePharmacy(id uint) error {
	return s.db.Delete(&models.PharmacyModel{}, "id = ?", id).Error
}
