package promotion

import (
	"errors"
	"strconv"
	"strings"

	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/inventory"
	"github.com/AMETORY/ametory-erp-modules/shared/models"
	"github.com/AMETORY/ametory-erp-modules/utils"
	"gorm.io/gorm"
)

type PromotionService struct {
	ctx              *context.ERPContext
	db               *gorm.DB
	inventoryService *inventory.InventoryService
}

func NewPromotionService(db *gorm.DB, ctx *context.ERPContext) *PromotionService {
	var inventoryService *inventory.InventoryService
	if ctx.InventoryService != nil {
		inventoryService = ctx.InventoryService.(*inventory.InventoryService)
	}
	return &PromotionService{db: db, ctx: ctx, inventoryService: inventoryService}
}

func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(&models.PromotionModel{}, &models.PromotionRuleModel{}, &models.PromotionActionModel{})
}

func (s *PromotionService) CheckPromotionEligibility(promotionID string, ruleType string, condition string) (bool, error) {
	var rule models.PromotionRuleModel
	err := s.ctx.DB.Where("promotion_id = ? and rule_type = ?", promotionID, ruleType).Find(&rule).Error
	if err != nil {
		return false, err
	}

	switch rule.RuleType {
	case "min_purchase":
		minPurchase, err := strconv.ParseFloat(rule.RuleValue, 64)
		if err != nil {
			return false, err
		}
		orderTotal, err := strconv.ParseFloat(condition, 64)
		if err != nil {
			return false, err
		}
		if orderTotal < minPurchase {
			return false, nil
		}
	case "max_purchase":
		maxPurchase, err := strconv.ParseFloat(rule.RuleValue, 64)
		if err != nil {
			return false, err
		}
		orderTotal, err := strconv.ParseFloat(condition, 64)
		if err != nil {
			return false, err
		}
		if orderTotal > maxPurchase {
			return false, nil
		}
	case "category":
		if rule.RuleValue != condition {
			return false, nil
		}
	case "categories":
		categories := strings.Split(rule.RuleValue, ",")
		if !utils.ContainsString(categories, condition) {
			return false, nil
		}
	case "products":
		products := strings.Split(rule.RuleValue, ",")
		if !utils.ContainsString(products, condition) {
			return false, nil
		}
	case "customer_level":
		if rule.RuleValue != condition {
			return false, nil
		}
	default:
		return false, errors.New("unknown rule type")
	}

	return true, nil
}

func (s *PromotionService) ApplyPromotion(promotionID string, orderTotal float64, cartItems map[string]int) (*PromotionResult, error) {
	var actions []models.PromotionActionModel
	err := s.ctx.DB.Where("promotion_id = ?", promotionID).Find(&actions).Error
	if err != nil {
		return nil, err
	}

	discount := 0.0
	discountShippingRate := 0.0
	discountShippingAmount := 0.0
	freeShipping := false
	freeItems := map[string]int{}

	for _, action := range actions {
		switch action.ActionType {
		case "discount":
			discountValue, err := strconv.ParseFloat(action.ActionValue, 64)
			if err != nil {
				return nil, err
			}
			discount += discountValue // Bisa berupa nilai tetap atau persentase
		case "discount_percent":
			discountPercent, err := strconv.ParseFloat(action.ActionValue, 64)
			if err != nil {
				return nil, err
			}
			discount += (orderTotal * discountPercent / 100)
		case "free_shipping":
			freeShipping = true
		case "discount_shipping":
			discountShipping, err := strconv.ParseFloat(action.ActionValue, 64)
			if err != nil {
				return nil, err
			}
			discountShippingRate += discountShipping
		case "discount_amount_shipping":
			discountShipping, err := strconv.ParseFloat(action.ActionValue, 64)
			if err != nil {
				return nil, err
			}
			discountShippingAmount += discountShipping

		case "free_item":
			// Free item biasanya berupa produk tertentu

			// Contoh menggunakan inventory service
			freeItems[action.ActionValue] = cartItems[action.ActionValue]

		}
	}

	// Hitung total setelah diskon
	finalTotal := orderTotal - discount
	if finalTotal < 0 {
		finalTotal = 0
	}

	// Return hasil penerapan promosi
	return &PromotionResult{
		Discount:               discount,
		FinalTotal:             finalTotal,
		FreeShipping:           freeShipping,
		FreeItems:              freeItems,
		DiscountShippingRate:   discountShippingRate,
		DiscountShippingAmount: discountShippingAmount,
	}, nil
}

type PromotionResult struct {
	FinalTotal             float64        `json:"final_total"`
	Discount               float64        `json:"discount"`
	FreeShipping           bool           `json:"free_shipping"`
	FreeItems              map[string]int `json:"free_items"`
	DiscountShippingRate   float64        `json:"discount_shipping_rate"`
	DiscountShippingAmount float64        `json:"discount_shipping_amount"`
}
