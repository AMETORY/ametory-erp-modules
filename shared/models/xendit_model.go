package models

import "github.com/AMETORY/ametory-erp-modules/shared"

type XenditModel struct {
	shared.BaseModel
	MerchantID      *string        `json:"merchant_id" gorm:"index;constraint:OnDelete:CASCADE;"`
	Merchant        *MerchantModel `gorm:"foreignKey:MerchantID;constraint:OnDelete:CASCADE;" json:"merchant,omitempty"`
	EnableQRIS      bool           `json:"enable_qris" gorm:"default:false"`
	EnableDANA      bool           `json:"enable_dana" gorm:"default:false"`
	EnableLinkAja   bool           `json:"enable_link_aja" gorm:"default:false"`
	EnableShopeePay bool           `json:"enable_shopee_pay" gorm:"default:false"`
	EnableOVO       bool           `json:"enable_ovo" gorm:"default:false"`
	EnableBCA       bool           `json:"enable_bca" gorm:"default:false"`
	EnableMANDIRI   bool           `json:"enable_mandiri" gorm:"default:false"`
	EnableBNI       bool           `json:"enable_bni" gorm:"default:false"`
	EnableBRI       bool           `json:"enable_bri" gorm:"default:false"`
	QRISFee         float64        `json:"qris_fee" gorm:"default:0.7"`
	DANAFee         float64        `json:"dana_fee" gorm:"default:3"`
	OVOFee          float64        `json:"ovo_fee" gorm:"default:3.18"`
	LinkAjaFee      float64        `json:"link_aja_fee" gorm:"default:3.15"`
	ShopeePayFee    float64        `json:"shopee_pay_fee" gorm:"default:4"`
	VAFee           float64        `json:"va_fee" gorm:"default:7000"`
}

func (XenditModel) TableName() string {
	return "xendit"
}
