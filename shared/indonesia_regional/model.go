package indonesia_regional

import "gorm.io/gorm"

type RegProvince struct {
	ID        string       `gorm:"primary_key" json:"id"`
	Name      string       `gorm:"type:varchar(100);not null" json:"name"`
	Regencies []RegRegency `gorm:"foreignKey:ProvinceID;references:ID" json:"regencies"`
}

func (RegProvince) TableName() string {
	return "reg_provinces"
}

type RegRegency struct {
	ID         string        `gorm:"primary_key" json:"id"`
	ProvinceID string        `gorm:"type:char(2);not null;index" json:"province_id"`
	Name       string        `gorm:"type:varchar(100);not null" json:"name"`
	Districts  []RegDistrict `gorm:"foreignKey:RegencyID;references:ID" json:"districts"`
}

func (RegRegency) TableName() string {
	return "reg_regencies"
}

type RegDistrict struct {
	ID        string       `gorm:"primary_key" json:"id"`
	RegencyID string       `gorm:"type:varchar(4);not null;index" json:"regency_id"`
	Name      string       `gorm:"type:varchar(100);not null" json:"name"`
	Villages  []RegVillage `gorm:"foreignKey:DistrictID;references:ID" json:"villages"`
}

func (RegDistrict) TableName() string {
	return "reg_districts"
}

type RegVillage struct {
	ID         string `gorm:"primary_key" json:"id"`
	DistrictID string `gorm:"type:varchar(6);not null;index" json:"district_id"`
	Name       string `gorm:"type:varchar(100);not null" json:"name"`
}

func (RegVillage) TableName() string {
	return "reg_villages"
}

func GetProvince(db *gorm.DB, id string) RegProvince {
	var province RegProvince
	db.Model(&RegProvince{}).Where("id = ?", id).First(&province)
	return province
}
func GetRegency(db *gorm.DB, id string) RegRegency {
	var regency RegRegency
	db.Model(&RegRegency{}).Where("id = ?", id).First(&regency)
	return regency
}

func GetDistrict(db *gorm.DB, id string) RegDistrict {
	var district RegDistrict
	db.Model(&RegDistrict{}).Where("id = ?", id).First(&district)
	return district
}

func GetVillage(db *gorm.DB, id string) RegVillage {
	var village RegVillage
	db.Model(&RegVillage{}).Where("id = ?", id).First(&village)
	return village
}
