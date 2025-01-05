package product

import (
	"github.com/AMETORY/ametory-erp-modules/utils"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PriceCategoryModel struct {
	utils.BaseModel
	Name        string `gorm:"unique;not null"`
	Description string
}

func (PriceCategoryModel) TableName() string {
	return "price_categories"
}

func (p *PriceCategoryModel) BeforeCreate(tx *gorm.DB) (err error) {
	if p.ID == "" {
		tx.Statement.SetColumn("id", uuid.New().String())
	}
	return
}

var samplePriceCategories = []PriceCategoryModel{
	{
		Name:        "Retail",
		Description: "Retail Price",
	},
	{
		Name:        "Wholesale",
		Description: "Wholesale Price",
	},
	{
		Name:        "Distributor",
		Description: "Distributor Price",
	},
	{
		Name:        "Drop Shipper",
		Description: "Drop Shipper Price",
	},
	{
		Name:        "Liburan",
		Description: "Harga Liburan (Indonesia)",
	},
	{
		Name:        "Hari Raya Nyepi",
		Description: "Harga Hari Raya Nyepi (Indonesia)",
	},
	{
		Name:        "Hari Raya Idul Fitri",
		Description: "Harga Hari Raya Idul Fitri (Indonesia)",
	},
	{
		Name:        "Hari Raya Idul Adha",
		Description: "Harga Hari Raya Idul Adha (Indonesia)",
	},
	{
		Name:        "Hari Natal",
		Description: "Harga Hari Natal (Indonesia)",
	},
	{
		Name:        "Waisak",
		Description: "Harga Waisak (Indonesia)",
	},
	{
		Name:        "Chinese New Year",
		Description: "Harga Tahun Baru Imlek (Indonesia)",
	},
	{
		Name:        "Promo HARBOLNAS",
		Description: "Harga Promo Hari Belanja Online Nasional (Indonesia)",
	},
	{
		Name:        "Promo Puasa",
		Description: "Harga Promo Selama Bulan Puasa (Indonesia)",
	},
	{
		Name:        "Hari Kemerdekaan",
		Description: "Harga Hari Kemerdekaan (Indonesia)",
	},
}

func SamplePriceCategories() []PriceCategoryModel {
	return samplePriceCategories
}
