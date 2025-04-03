package payment_term

import (
	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/shared/models"
	"github.com/AMETORY/ametory-erp-modules/utils"
	"gorm.io/gorm"
)

type PaymentTermService struct {
	ctx *context.ERPContext
	db  *gorm.DB
}

func NewPaymentTermService(db *gorm.DB, ctx *context.ERPContext) *PaymentTermService {
	return &PaymentTermService{ctx: ctx, db: db}
}

func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&models.PaymentTermModel{},
	)
}

func (s *PaymentTermService) GetPaymentTerms() []models.PaymentTermModel {
	var paymentTerms []models.PaymentTermModel
	s.db.Find(&paymentTerms)
	return paymentTerms
}

func (s *PaymentTermService) GroupPaymentTermsByCategory() map[string][]models.PaymentTermModel {
	var paymentTerms []models.PaymentTermModel
	s.db.Find(&paymentTerms)

	grouped := make(map[string][]models.PaymentTermModel)

	for _, term := range paymentTerms {
		grouped[term.Category] = append(grouped[term.Category], term)
	}
	return grouped
}

func (s *PaymentTermService) InitPaymentTerms() error {
	terms := []models.PaymentTermModel{
		{
			Name:        "Cash on Delivery",
			Code:        "COD",
			Description: "Pembayaran dilakukan saat barang diterima.",
			Category:    "Immediate Payment",
		},
		{
			Name:        "Net 7",
			Code:        "NET7",
			Description: "Pembayaran harus dilakukan dalam 7 hari setelah invoice diterima.",
			Category:    "Standard Payment",
			DueDays:     utils.PtrInt(7),
		},
		{
			Name:        "Net 14",
			Code:        "NET14",
			Description: "Pembayaran harus dilakukan dalam 14 hari setelah invoice diterima.",
			Category:    "Standard Payment",
			DueDays:     utils.PtrInt(14),
		},
		{
			Name:        "Net 30",
			Code:        "NET30",
			Description: "Pembayaran harus dilakukan dalam 30 hari setelah invoice diterima.",
			Category:    "Standard Payment",
			DueDays:     utils.PtrInt(30),
		},
		{
			Name:        "Net 60",
			Code:        "NET60",
			Description: "Pembayaran harus dilakukan dalam 60 hari setelah invoice diterima.",
			Category:    "Standard Payment",
			DueDays:     utils.PtrInt(60),
		},
		{
			Name:        "Net 90",
			Code:        "NET90",
			Description: "Pembayaran harus dilakukan dalam 90 hari setelah invoice diterima.",
			Category:    "Standard Payment",
			DueDays:     utils.PtrInt(90),
		},
		{
			Name:            "2/10 Net 30",
			Code:            "2_10_NET30",
			Description:     "Diskon 2% jika dibayar dalam 10 hari, jika tidak harus dibayar penuh dalam 30 hari.",
			Category:        "Early Payment Discount",
			DueDays:         utils.PtrInt(30),
			DiscountAmount:  utils.PtrFloat64(2),
			DiscountDueDays: utils.PtrInt(10),
		},
		{
			Name:            "3/15 Net 45",
			Code:            "3_15_NET45",
			Description:     "Diskon 3% jika dibayar dalam 15 hari, jika tidak harus dibayar penuh dalam 45 hari.",
			Category:        "Early Payment Discount",
			DueDays:         utils.PtrInt(45),
			DiscountAmount:  utils.PtrFloat64(3),
			DiscountDueDays: utils.PtrInt(15),
		},
		{
			Name:        "50/50",
			Code:        "50_50",
			Description: "50% di muka, 50% setelah pengiriman barang atau penyelesaian pekerjaan.",
			Category:    "Installment Payment",
		},
		{
			Name:        "30/40/30",
			Code:        "30_40_30",
			Description: "30% di muka, 40% saat pekerjaan 50% selesai, 30% setelah selesai.",
			Category:    "Installment Payment",
		},
		{
			Name:        "Milestone-Based",
			Code:        "MILESTONE",
			Description: "Pembayaran dilakukan berdasarkan progress yang telah disepakati.",
			Category:    "Installment Payment",
		},
		{
			Name:        "Monthly Billing",
			Code:        "MONTHLY",
			Description: "Pembayaran dilakukan setiap bulan sesuai kontrak.",
			Category:    "Subscription Payment",
			DueDays:     utils.PtrInt(30),
		},
		{
			Name:        "Quarterly Billing",
			Code:        "QUARTERLY",
			Description: "Pembayaran dilakukan setiap 3 bulan sekali.",
			Category:    "Subscription Payment",
			DueDays:     utils.PtrInt(90),
		},
		{
			Name:            "Annual Billing",
			Code:            "ANNUAL",
			Description:     "Pembayaran dilakukan setiap tahun dengan kemungkinan diskon.",
			Category:        "Subscription Payment",
			DueDays:         utils.PtrInt(365),
			DiscountAmount:  utils.PtrFloat64(5),
			DiscountDueDays: utils.PtrInt(30),
		},
		{
			Name:        "L/C at Sight",
			Code:        "LC_SIGHT",
			Description: "Pembayaran melalui Letter of Credit yang dapat dicairkan segera setelah dokumen diserahkan.",
			Category:    "Bank Payment",
		},
		{
			Name:        "L/C 30 Days",
			Code:        "LC_30",
			Description: "Letter of Credit dengan pembayaran tertunda 30 hari setelah dokumen disetujui.",
			Category:    "Bank Payment",
			DueDays:     utils.PtrInt(30),
		},
		{
			Name:        "Bank Guarantee",
			Code:        "BANK_GUARANTEE",
			Description: "Pembayaran dijamin oleh bank jika persyaratan dipenuhi.",
			Category:    "Bank Payment",
		},
	}

	return s.db.CreateInBatches(&terms, 10).Error
}
