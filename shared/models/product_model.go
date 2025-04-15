package models

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/AMETORY/ametory-erp-modules/shared"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ProductModel struct {
	shared.BaseModel
	Name             string                 `gorm:"not null" json:"name,omitempty"`
	Description      *string                `json:"description,omitempty"`
	SKU              *string                `gorm:"type:varchar(255)" json:"sku,omitempty"`
	Barcode          *string                `gorm:"type:varchar(255)" json:"barcode,omitempty"`
	Price            float64                `gorm:"not null;default:0" json:"price,omitempty"`
	CompanyID        *string                `json:"company_id,omitempty"`
	Company          *CompanyModel          `gorm:"foreignKey:CompanyID;constraint:OnDelete:CASCADE" json:"company,omitempty"`
	DistributorID    *string                `gorm:"foreignKey:DistributorID;references:ID;constraint:OnDelete:CASCADE" json:"distributor_id,omitempty"`
	Distributor      *DistributorModel      `gorm:"foreignKey:DistributorID;constraint:OnDelete:CASCADE" json:"distributor,omitempty"`
	MasterProductID  *string                `json:"master_product_id,omitempty"`
	MasterProduct    *MasterProductModel    `gorm:"foreignKey:MasterProductID;constraint:OnDelete:CASCADE" json:"master_product,omitempty"`
	CategoryID       *string                `json:"category_id,omitempty"`
	Category         *ProductCategoryModel  `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;foreignKey:CategoryID" json:"category,omitempty"`
	Prices           []PriceModel           `gorm:"-" json:"prices,omitempty"`
	Brand            *BrandModel            `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;foreignKey:BrandID" json:"brand,omitempty"`
	BrandID          *string                `json:"brand_id,omitempty"`
	ProductImages    []FileModel            `gorm:"-" json:"product_images,omitempty"`
	TotalStock       float64                `gorm:"-" json:"total_stock,omitempty"`
	SalesCount       float64                `gorm:"-" json:"sales_count,omitempty"`
	LastUpdatedStock *time.Time             `gorm:"-" json:"last_updated_stock,omitempty"`
	LastStock        float64                `gorm:"-" json:"last_stock,omitempty"`
	Status           string                 `gorm:"type:VARCHAR(20);default:'ACTIVE'" json:"status,omitempty"`
	Merchants        []*MerchantModel       `gorm:"many2many:product_merchants;constraint:OnDelete:CASCADE;" json:"merchants,omitempty"`
	DisplayName      string                 `gorm:"type:varchar(255)" json:"display_name,omitempty"`
	Discounts        []DiscountModel        `gorm:"foreignKey:ProductID;constraint:OnDelete:CASCADE" json:"discounts,omitempty"`
	ActiveDiscount   *DiscountModel         `gorm:"-" json:"active_discount,omitempty"`
	PriceList        []float64              `gorm:"-" json:"price_list,omitempty"`
	OriginalPrice    float64                `gorm:"-" json:"original_price,omitempty"`
	AdjustmentPrice  float64                `gorm:"-" json:"adjustment_price,omitempty"`
	Variants         []VariantModel         `gorm:"foreignKey:ProductID;constraint:OnDelete:CASCADE" json:"variants,omitempty"`
	Tags             []*TagModel            `gorm:"many2many:product_tags;constraint:OnDelete:CASCADE;" json:"tags"`
	Height           float64                `gorm:"default:10" json:"height,omitempty"`
	Length           float64                `gorm:"default:10" json:"length,omitempty"`
	Weight           float64                `gorm:"default:200" json:"weight,omitempty"`
	Width            float64                `gorm:"default:10" json:"width,omitempty"`
	DiscountAmount   float64                `gorm:"-" json:"discount_amount,omitempty"`
	DiscountType     string                 `gorm:"-" json:"discount_type,omitempty"`
	DiscountRate     float64                `gorm:"-" json:"discount_rate,omitempty"`
	MerchantID       *string                `json:"-" gorm:"-"`
	ProductImageIDs  []string               `gorm:"-" json:"product_image_ids,omitempty"`
	Feedbacks        []ProductFeedbackModel `gorm:"foreignKey:ProductID;constraint:OnDelete:CASCADE" json:"feedbacks,omitempty"`
	Units            []*UnitModel           `gorm:"many2many:product_units;constraint:OnDelete:CASCADE;" json:"units,omitempty"`
	DefaultUnit      *UnitModel             `gorm:"-" json:"default_unit,omitempty"`
}

func (ProductModel) TableName() string {
	return "products"
}

func (p *ProductModel) AfterFind(tx *gorm.DB) (err error) {
	productUnits := []ProductUnitData{}
	tx.Where("product_model_id = ?", p.ID).Find(&productUnits)
	var units []*UnitModel
	for _, v := range productUnits {
		var unit UnitModel
		tx.Where("id = ?", v.UnitModelID).Find(&unit)
		unit.Value = v.Value
		unit.IsDefault = v.IsDefault
		units = append(units, &unit)
		if v.IsDefault {
			p.DefaultUnit = &unit
		}
	}
	p.Units = units
	return nil
}

func (p *ProductModel) GetPrices(tx *gorm.DB) (err error) {
	err = tx.Model(&p.Prices).Preload("PriceCategory").Where("product_id = ?", p.ID).Find(&p.Prices).Error
	return
}
func (p *ProductModel) GetPriceAndDiscount(tx *gorm.DB) (err error) {
	var pp ProductModel
	tx.Select("price").Model(&p).First(&pp, "id = ?", p.ID)
	p.OriginalPrice = pp.Price

	var pm []ProductMerchant
	tx.Select("price").Where("product_model_id = ? and price > 0", p.ID).Find(&pm)
	p.PriceList = []float64{pp.Price}
	for _, v := range pm {
		if v.Price != p.Price && v.Price > 0 {
			p.PriceList = append(p.PriceList, v.Price)
		}
	}

	if p.MerchantID != nil {
		// fmt.Println("MERCHANT ID @ PRODUCT", *p.MerchantID)
		var productMerchant ProductMerchant
		err := tx.Select("price", "adjustment_price").Where("product_model_id = ? AND merchant_model_id = ?", p.ID, *p.MerchantID).First(&productMerchant) // TODO: check if variant_merchant exists

		if err == nil {
			p.AdjustmentPrice = productMerchant.AdjustmentPrice
			p.Price += productMerchant.AdjustmentPrice
			// fmt.Println("MERCHANT PRICE 2", p.Price)
			// p.OriginalPrice = productMerchant.Price
		}
	} else {
		fmt.Println("MERCHANT ID @ PRODUCT", "NOT FOUND")
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
		case DiscountAmount:
			discountAmount = discount.Value
		}

		discountedPrice -= discountAmount
		// Pastikan harga tidak negatif
		if discountedPrice < 0 {
			discountedPrice = 0
		}
		p.Price = discountedPrice
		p.DiscountAmount = discountAmount
		p.DiscountType = string(discount.Type)
		p.DiscountRate = discount.Value
		p.ActiveDiscount = &discount
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
	AdjustmentPrice  float64    `gorm:"column:adjustment_price;default:0" json:"adjustment_price"`
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

type PopularProduct struct {
	ID          string  `json:"id"`
	DisplayName string  `json:"display_name"`
	TotalStock  float64 `json:"total_stock"`
	TotalSale   float64 `json:"total_sale"`
	TotalView   float64 `json:"total_view"`
}
