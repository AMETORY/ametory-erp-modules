package indonesia_regional

type RegProvince struct {
	ID        string
	Name      string       `gorm:"type:varchar(100);not null"`
	Regencies []RegRegency `gorm:"foreignKey:ProvinceID;references:ID"`
}

func (RegProvince) TableName() string {
	return "reg_provinces"
}

type RegRegency struct {
	ID         string
	ProvinceID string        `gorm:"type:varchar(36);not null;index"`
	Name       string        `gorm:"type:varchar(100);not null"`
	Districts  []RegDistrict `gorm:"foreignKey:RegencyID;references:ID"`
}

func (RegRegency) TableName() string {
	return "reg_regencies"
}

type RegDistrict struct {
	ID        string       `gorm:"type:varchar(36);not null;index"`
	RegencyID string       `gorm:"type:varchar(36);not null;index"`
	Name      string       `gorm:"type:varchar(100);not null"`
	Villages  []RegVillage `gorm:"foreignKey:DistrictID;references:ID"`
}

func (RegDistrict) TableName() string {
	return "reg_districts"
}

type RegVillage struct {
	ID         string `gorm:"type:varchar(36);not null;index"`
	DistrictID string `gorm:"type:varchar(36);not null;index"`
	Name       string `gorm:"type:varchar(100);not null"`
}

func (RegVillage) TableName() string {
	return "reg_villages"
}
