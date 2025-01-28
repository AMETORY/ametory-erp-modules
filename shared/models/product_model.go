package models

import (
	"sort"
	"strings"
	"time"

	"github.com/AMETORY/ametory-erp-modules/shared"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ProductModel struct {
	shared.BaseModel
	Name             string                `gorm:"not null" json:"name,omitempty"`
	Description      *string               `json:"description,omitempty"`
	SKU              *string               `gorm:"type:varchar(255)" json:"sku,omitempty"`
	Barcode          *string               `gorm:"type:varchar(255)" json:"barcode,omitempty"`
	Price            float64               `gorm:"not null;default:0" json:"price,omitempty"`
	CompanyID        *string               `json:"company_id,omitempty"`
	Company          *CompanyModel         `gorm:"foreignKey:CompanyID;constraint:OnDelete:CASCADE" json:"company,omitempty"`
	DistributorID    *string               `gorm:"foreignKey:DistributorID;references:ID;constraint:OnDelete:CASCADE" json:"distributor_id,omitempty"`
	Distributor      *DistributorModel     `gorm:"foreignKey:DistributorID;constraint:OnDelete:CASCADE" json:"distributor,omitempty"`
	MasterProductID  *string               `json:"master_product_id,omitempty"`
	MasterProduct    *MasterProductModel   `gorm:"foreignKey:MasterProductID;constraint:OnDelete:CASCADE" json:"master_product,omitempty"`
	CategoryID       *string               `json:"category_id,omitempty"`
	Category         *ProductCategoryModel `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;foreignKey:CategoryID" json:"category,omitempty"`
	Prices           []PriceModel          `gorm:"-" json:"prices,omitempty"`
	Brand            *BrandModel           `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;foreignKey:BrandID" json:"brand,omitempty"`
	BrandID          *string               `json:"brand_id,omitempty"`
	ProductImages    []FileModel           `gorm:"-" json:"product_images,omitempty"`
	TotalStock       float64               `gorm:"-" json:"total_stock,omitempty"`
	LastUpdatedStock *time.Time            `gorm:"-" json:"last_updated_stock,omitempty"`
	LastStock        float64               `gorm:"-" json:"last_stock,omitempty"`
	Status           string                `gorm:"type:VARCHAR(20);default:'ACTIVE'" json:"status,omitempty"`
	Merchants        []*MerchantModel      `gorm:"many2many:product_merchants;constraint:OnDelete:CASCADE;" json:"merchants,omitempty"`
	DisplayName      string                `gorm:"type:varchar(255)" json:"display_name,omitempty"`
	Discounts        []DiscountModel       `gorm:"foreignKey:ProductID;constraint:OnDelete:CASCADE" json:"discounts,omitempty"`
	ActiveDiscount   *DiscountModel        `gorm:"-" json:"active_discount,omitempty"`
	PriceList        []float64             `gorm:"-" json:"price_list,omitempty"`
	OriginalPrice    float64               `gorm:"-" json:"original_price,omitempty"`
	Variants         []VariantModel        `gorm:"foreignKey:ProductID;constraint:OnDelete:CASCADE" json:"variants,omitempty"`
	Tags             []*TagModel           `gorm:"many2many:product_tags;constraint:OnDelete:CASCADE;" json:"tags"`
	Height           float64               `gorm:"default:10" json:"height,omitempty"`
	Length           float64               `gorm:"default:10" json:"length,omitempty"`
	Weight           float64               `gorm:"default:200" json:"weight,omitempty"`
	Width            float64               `gorm:"default:10" json:"width,omitempty"`
	DiscountAmount   float64               `gorm:"-" json:"discount_amount,omitempty"`
	DiscountType     string                `gorm:"-" json:"discount_type,omitempty"`
	DiscountRate     float64               `gorm:"-" json:"discount_rate,omitempty"`
}

func (ProductModel) TableName() string {
	return "products"
}

func (p *ProductModel) AfterFind(tx *gorm.DB) (err error) {
	var pp ProductModel
	tx.Select("price").Model(&p).First(&pp, "id = ?", p.ID)
	p.OriginalPrice = pp.Price

	var pm []ProductMerchant
	tx.Select("price").Where("product_model_id = ?", p.ID).Find(&pm)
	p.PriceList = []float64{pp.Price}
	for _, v := range pm {
		if v.Price != p.Price {
			p.PriceList = append(p.PriceList, v.Price)
		}
	}

	sort.Float64s(p.PriceList)
	var discount DiscountModel
	tx.Where("product_id = ? AND is_active = ? AND start_date <= ?", p.ID, true, time.Now()).
		Where("end_date IS NULL OR end_date >= ?", time.Now()).Order("created_at DESC").
		Find(&discount)
	if discount.ID != "" {
		discountAmount := float64(0)
		discountedPrice := p.Price
		switch discount.Type {
		case DiscountPercentage:
			discountAmount = p.Price * (discount.Value / 100)
			discountedPrice -= p.Price * (discount.Value / 100)
		case DiscountAmount:
			discountAmount = discount.Value
			discountedPrice -= discount.Value
		}

		// Pastikan harga tidak negatif
		if discountedPrice < 0 {
			discountedPrice = 0
		}
		p.Price = discountedPrice
		p.DiscountAmount = discountAmount
		p.DiscountType = string(discount.Type)
		p.DiscountRate = discount.Value
	}
	return
}

func (p *ProductModel) BeforeCreate(tx *gorm.DB) (err error) {
	if p.ID == "" {
		tx.Statement.SetColumn("id", uuid.New().String())
	}
	return
}

func (p *ProductModel) BeforeSave(tx *gorm.DB) (err error) {
	p.GenerateDisplayName(tx)
	return
}

type ProductMerchant struct {
	ProductModelID   string     `gorm:"primaryKey;uniqueIndex:product_merchants_product_id_merchant_id_key" json:"product_id"`
	MerchantModelID  string     `gorm:"primaryKey;uniqueIndex:product_merchants_product_id_merchant_id_key" json:"merchant_id"`
	LastUpdatedStock *time.Time `gorm:"column:last_updated_stock" json:"last_updated_stock"`
	LastStock        float64    `gorm:"column:last_stock" json:"last_stock"`
	Price            float64    `gorm:"column:price" json:"price"`
}

func (v *ProductModel) GenerateDisplayName(tx *gorm.DB) {

	var displayNames []string
	var brand BrandModel
	if v.BrandID != nil {
		tx.Find(&brand, "id = ?", v.BrandID)
		displayNames = append(displayNames, brand.Name)
	}

	displayNames = append(displayNames, v.Name)

	// Generate display name
	v.DisplayName = strings.Join(displayNames, " ")
}
