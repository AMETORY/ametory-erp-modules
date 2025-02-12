package models

import (
	"sort"
	"strings"
	"time"

	"github.com/AMETORY/ametory-erp-modules/shared"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type VariantModel struct {
	shared.BaseModel
	ProductID        string                         `gorm:"index" json:"product_id,omitempty"`
	Product          *ProductModel                  `gorm:"foreignKey:ProductID" json:"product,omitempty"`
	SKU              string                         `gorm:"type:varchar(255);not null" json:"sku,omitempty"`
	Barcode          *string                        `gorm:"type:varchar(255)" json:"barcode,omitempty"`
	Price            float64                        `gorm:"not null;default:0" json:"price,omitempty"`
	Attributes       []VariantProductAttributeModel `gorm:"foreignKey:VariantID;constraint:OnDelete:CASCADE" json:"attributes,omitempty"`
	DisplayName      string                         `gorm:"type:varchar(255)" json:"display_name,omitempty"`
	TotalStock       float64                        `gorm:"-" json:"total_stock,omitempty"`
	SalesCount       float64                        `gorm:"-" json:"sales_count,omitempty"`
	Tags             []*TagModel                    `gorm:"many2many:variant_tags;constraint:OnDelete:CASCADE;" json:"tags,omitempty"`
	PriceList        []float64                      `gorm:"-" json:"price_list,omitempty"`
	OriginalPrice    float64                        `gorm:"-" json:"original_price,omitempty"`
	LastUpdatedStock *time.Time                     `gorm:"-" json:"last_updated_stock,omitempty"`
	LastStock        float64                        `gorm:"-" json:"last_stock,omitempty"`
	Height           float64                        `gorm:"default:10" json:"height,omitempty"`
	Length           float64                        `gorm:"default:10" json:"length,omitempty"`
	Weight           float64                        `gorm:"default:200" json:"weight,omitempty"`
	Width            float64                        `gorm:"default:10" json:"width,omitempty"`
	DiscountAmount   float64                        `gorm:"-" json:"discount_amount,omitempty"`
	DiscountType     string                         `gorm:"-" json:"discount_type,omitempty"`
	DiscountRate     float64                        `gorm:"-" json:"discount_rate,omitempty"`
	ActiveDiscount   *DiscountModel                 `gorm:"-" json:"active_discount,omitempty"`
	MerchantID       *string                        `json:"-" gorm:"-"`
}

func (VariantModel) TableName() string {
	return "product_variants"
}

func (v *VariantModel) BeforeCreate(tx *gorm.DB) error {
	if v.ID == "" {
		tx.Statement.SetColumn("id", uuid.New().String())
	}
	v.GenerateDisplayName(tx)
	return nil
}

func (v *VariantModel) GenerateDisplayName(tx *gorm.DB) {
	var productName string
	attributeValues := make([]string, len(v.Attributes))

	var product ProductModel
	if err := tx.Preload("Brand", func(db *gorm.DB) *gorm.DB {
		return db.Select("name", "id")
	}).Where("id = ?", v.ProductID).Select("id", "name", "brand_id").First(&product).Error; err != nil {
		return
	}
	productName = product.Name

	// Sort attributes by priority
	sort.SliceStable(v.Attributes, func(i, j int) bool {
		return v.Attributes[i].Attribute.Priority < v.Attributes[j].Attribute.Priority
	})

	// Collect attribute values
	for i, attr := range v.Attributes {
		attributeValues[i] = attr.Value
	}

	var displayNames []string
	if product.Brand != nil {
		displayNames = append(displayNames, product.Brand.Name)
	}
	displayNames = append(displayNames, productName)
	displayNames = append(displayNames, strings.Join(attributeValues, " "))

	// Generate display name
	v.DisplayName = strings.Join(displayNames, " ")
}

func (p *VariantModel) AfterFind(tx *gorm.DB) (err error) {
	var pp VariantModel
	tx.Select("price").Model(&p).First(&pp, "id = ?", p.ID)
	p.OriginalPrice = pp.Price

	var pm []VarianMerchant
	tx.Select("price").Where("variant_id = ?", p.ID).Find(&pm)
	p.PriceList = []float64{pp.Price}
	for _, v := range pm {
		if v.Price != p.Price {
			p.PriceList = append(p.PriceList, v.Price)
		}
	}
	if p.MerchantID != nil {
		var variantMerchant VarianMerchant
		tx.Select("price").Where("variant_id = ? AND merchant_id = ?", p.ID, *p.MerchantID).Find(&variantMerchant) // TODO: check if variant_merchant exists
		p.Price = variantMerchant.Price
		p.OriginalPrice = variantMerchant.Price
		// fmt.Println("KESINI", variantMerchant)
	}
	var discount DiscountModel
	tx.Where("product_id = ? AND is_active = ? AND start_date <= ?", p.ProductID, true, time.Now()).
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
		p.ActiveDiscount = &discount
	}

	// sort.Float64s(p.PriceList)
	return
}

type VarianMerchant struct {
	shared.BaseModel
	MerchantID       string        `gorm:"type:char(36);uniqueIndex:variant_merchant_merchant_id_variant_id_key"`
	Variant          VariantModel  `gorm:"foreignKey:VariantID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE"`
	VariantID        string        `gorm:"type:char(36);uniqueIndex:variant_merchant_merchant_id_variant_id_key"`
	Merchant         MerchantModel `gorm:"foreignKey:MerchantID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE"`
	LastUpdatedStock *time.Time    `gorm:"column:last_updated_stock" json:"last_updated_stock"`
	LastStock        float64       `gorm:"column:last_stock" json:"last_stock"`
	Price            float64
}

func (vm *VarianMerchant) BeforeCreate(tx *gorm.DB) error {
	if vm.ID == "" {
		tx.Statement.SetColumn("id", uuid.New().String())
	}
	return nil
}
