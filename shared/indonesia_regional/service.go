package indonesia_regional

import (
	"github.com/AMETORY/ametory-erp-modules/context"
)

type IndonesiaRegService struct {
	ctx *context.ERPContext
}

func NewIndonesiaRegService(ctx *context.ERPContext) *IndonesiaRegService {
	return &IndonesiaRegService{
		ctx: ctx,
	}
}

func (ir *IndonesiaRegService) GetProvinces(search string) ([]RegProvince, error) {
	var provinces []RegProvince
	db := ir.ctx.DB
	if search != "" {
		db = db.Where("name ilike ?", "%"+search+"%")
	}
	err := db.Find(&provinces).Error
	return provinces, err
}

func (ir *IndonesiaRegService) GetRegencies(provinceID *string, search string) ([]RegRegency, error) {
	var regencies []RegRegency
	db := ir.ctx.DB
	if provinceID != nil {
		db = db.Where("province_id = ?", *provinceID)
	}
	if search != "" {
		db = db.Where("name ilike ?", "%"+search+"%")
	}
	err := db.Find(&regencies).Error
	return regencies, err
}

func (ir *IndonesiaRegService) GetDistricts(regencyID *string, search string) ([]RegDistrict, error) {
	var districts []RegDistrict
	db := ir.ctx.DB
	if regencyID != nil {
		db = db.Where("regency_id = ?", regencyID)
	}
	if search != "" {
		db = db.Where("name ilike ?", "%"+search+"%")
	}
	err := db.Find(&districts).Error
	return districts, err
}

func (ir *IndonesiaRegService) GetVillages(districtID *string, search string) ([]RegVillage, error) {
	var villages []RegVillage
	db := ir.ctx.DB
	if districtID != nil {
		db = db.Where("district_id = ?", districtID)
	}
	if search != "" {
		db = db.Where("name ilike ?", "%"+search+"%")
	}
	err := db.Find(&villages).Error
	return villages, err
}
