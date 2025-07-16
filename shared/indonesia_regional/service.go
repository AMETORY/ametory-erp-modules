package indonesia_regional

import (
	"github.com/AMETORY/ametory-erp-modules/context"
)

type IndonesiaRegService struct {
	ctx *context.ERPContext
}

// NewIndonesiaRegService creates a new instance of IndonesiaRegService
// with the given ERPContext.
func NewIndonesiaRegService(ctx *context.ERPContext) *IndonesiaRegService {
	return &IndonesiaRegService{
		ctx: ctx,
	}
}

// GetProvinces retrieves a list of provinces.
// It accepts an optional search string to filter provinces by name.
func (ir *IndonesiaRegService) GetProvinces(search string) ([]RegProvince, error) {
	var provinces []RegProvince
	db := ir.ctx.DB
	if search != "" {
		db = db.Where("name ILIKE ?", "%"+search+"%")
	}
	err := db.Find(&provinces).Error
	return provinces, err
}

// GetRegencies retrieves a list of regencies.
// It accepts an optional provinceID to filter regencies by province
// and an optional search string to filter regencies by name.
func (ir *IndonesiaRegService) GetRegencies(provinceID *string, search string) ([]RegRegency, error) {
	var regencies []RegRegency
	db := ir.ctx.DB
	if provinceID != nil {
		db = db.Where("province_id = ?", *provinceID)
	}
	if search != "" {
		db = db.Where("name ILIKE ?", "%"+search+"%")
	}
	err := db.Find(&regencies).Error
	return regencies, err
}

// GetDistricts retrieves a list of districts.
// It accepts an optional regencyID to filter districts by regency
// and an optional search string to filter districts by name.
func (ir *IndonesiaRegService) GetDistricts(regencyID *string, search string) ([]RegDistrict, error) {
	var districts []RegDistrict
	db := ir.ctx.DB
	if regencyID != nil {
		db = db.Where("regency_id = ?", *regencyID)
	}
	if search != "" {
		db = db.Where("name ILIKE ?", "%"+search+"%")
	}
	err := db.Find(&districts).Error
	return districts, err
}

// GetVillages retrieves a list of villages.
// It accepts an optional districtID to filter villages by district
// and an optional search string to filter villages by name.
func (ir *IndonesiaRegService) GetVillages(districtID *string, search string) ([]RegVillage, error) {
	var villages []RegVillage
	db := ir.ctx.DB
	if districtID != nil {
		db = db.Where("district_id = ?", *districtID)
	}
	if search != "" {
		db = db.Where("name ILIKE ?", "%"+search+"%")
	}
	err := db.Find(&villages).Error
	return villages, err
}
