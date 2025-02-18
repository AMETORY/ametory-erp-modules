package promotion

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/inventory"
	"github.com/AMETORY/ametory-erp-modules/shared/models"
	"github.com/AMETORY/ametory-erp-modules/utils"
	"github.com/morkid/paginate"
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
	case "MIN_PURCHASE":
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
	case "MAX_PURCHASE":
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
	case "CATEGORY":
		if rule.RuleValue != condition {
			return false, nil
		}
	case "CATEGORIES":
		categories := strings.Split(rule.RuleValue, ",")
		if !utils.ContainsString(categories, condition) {
			return false, nil
		}
	case "PRODUCTS":
		products := strings.Split(rule.RuleValue, ",")
		if !utils.ContainsString(products, condition) {
			return false, nil
		}
	case "CUSTOMER_LEVEL":
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
		case "DISCOUNT":
			discountValue, err := strconv.ParseFloat(action.ActionValue, 64)
			if err != nil {
				return nil, err
			}
			discount += discountValue // Bisa berupa nilai tetap atau persentase
		case "DISCOUNT_PERCENT":
			discountPercent, err := strconv.ParseFloat(action.ActionValue, 64)
			if err != nil {
				return nil, err
			}
			discount += (orderTotal * discountPercent / 100)
		case "FREE_SHIPPING":
			freeShipping = true
		case "DISCOUNT_SHIPPING":
			discountShipping, err := strconv.ParseFloat(action.ActionValue, 64)
			if err != nil {
				return nil, err
			}
			discountShippingRate += discountShipping
		case "DISCOUNT_AMOUNT_SHIPPING":
			discountShipping, err := strconv.ParseFloat(action.ActionValue, 64)
			if err != nil {
				return nil, err
			}
			discountShippingAmount += discountShipping

		case "FREE_ITEM":
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

func (s *PromotionService) CreatePromotion(data *models.PromotionModel) error {
	return s.db.Create(data).Error
}

func (s *PromotionService) UpdatePromotion(id string, data *models.PromotionModel) error {
	return s.db.Where("id = ?", id).Updates(data).Error
}

func (s *PromotionService) DeletePromotion(id string) error {
	return s.db.Where("id = ?", id).Delete(&models.PromotionModel{}).Error
}

func (s *PromotionService) GetPromotionByID(id string) (*models.PromotionModel, error) {
	var invoice models.PromotionModel
	err := s.db.Where("id = ?", id).First(&invoice).Error
	return &invoice, err
}

func (s *PromotionService) GetPromotionByCode(code string) (*models.PromotionModel, error) {
	var invoice models.PromotionModel
	err := s.db.Where("code = ?", code).First(&invoice).Error
	return &invoice, err
}

func (s *PromotionService) AddRule(promotionID string, data *models.PromotionRuleModel) error {
	data.PromotionID = promotionID
	return s.db.Create(data).Error
}

func (s *PromotionService) AddAction(promotionID string, data *models.PromotionActionModel) error {
	data.PromotionID = promotionID
	return s.db.Create(data).Error
}

func (s *PromotionService) DeleteRule(ruleID string) error {
	return s.db.Where("id = ?", ruleID).Delete(&models.PromotionRuleModel{}).Error
}

func (s *PromotionService) DeleteAction(actionID string) error {
	return s.db.Where("id = ?", actionID).Delete(&models.PromotionActionModel{}).Error
}
func (s *PromotionService) GetPromotions(request http.Request, search string) (paginate.Page, error) {
	pg := paginate.New()
	stmt := s.db
	if search != "" {
		stmt = stmt.Where("banners.description ILIKE ? OR banners.title ILIKE ?",
			"%"+search+"%",
			"%"+search+"%",
		)
	}
	if request.Header.Get("ID-Company") != "" {
		stmt = stmt.Where("company_id = ?", request.Header.Get("ID-Company"))
	}

	request.URL.Query().Get("page")
	stmt = stmt.Model(&models.PromotionModel{})
	utils.FixRequest(&request)
	page := pg.With(stmt).Request(request).Response(&[]models.PromotionModel{})
	page.Page = page.Page + 1
	return page, nil
}
func (s *PromotionService) GetUserPromotions(request http.Request, search string) (paginate.Page, error) {
	pg := paginate.New()
	stmt := s.db
	if search != "" {
		stmt = stmt.Where("banners.description ILIKE ? OR banners.title ILIKE ?",
			"%"+search+"%",
			"%"+search+"%",
		)
	}
	if request.Header.Get("ID-Company") != "" {
		stmt = stmt.Where("company_id = ?", request.Header.Get("ID-Company"))
	}

	stmt = stmt.Where("start_date <= ? AND end_date >= ?", time.Now(), time.Now())

	stmt = stmt.Where("is_active = ?", true)

	request.URL.Query().Get("page")
	stmt = stmt.Model(&models.PromotionModel{})
	utils.FixRequest(&request)
	page := pg.With(stmt).Request(request).Response(&[]models.PromotionModel{})
	page.Page = page.Page + 1
	return page, nil
}

func (s *PromotionService) GetPromotionByName(title string) (*models.PromotionModel, error) {
	var banner models.PromotionModel
	err := s.db.Where("title = ?", title).First(&banner).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			banner = models.PromotionModel{
				Name: title,
			}
			err := s.db.Create(&banner).Error
			if err != nil {
				return nil, err
			}
			return &banner, nil
		}
		return nil, err
	}
	return &banner, nil
}
