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

// NewPromotionService creates a new instance of PromotionService.
//
// db is the database service.
//
// ctx is the application context.
//
// It returns a pointer to a PromotionService.
func NewPromotionService(db *gorm.DB, ctx *context.ERPContext) *PromotionService {
	var inventoryService *inventory.InventoryService
	if ctx.InventoryService != nil {
		inventoryService = ctx.InventoryService.(*inventory.InventoryService)
	}
	return &PromotionService{db: db, ctx: ctx, inventoryService: inventoryService}
}

// Migrate applies the necessary database migrations for the promotion models.
//
// It ensures that the underlying database schema is up to date with the
// current version of the PromotionModel, PromotionRuleModel, and
// PromotionActionModel.
func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(&models.PromotionModel{}, &models.PromotionRuleModel{}, &models.PromotionActionModel{})
}

// CheckPromotionEligibilityByPosSales checks if a given POS sale is eligible for a promotion rule.
//
// ruleID is the ID of the promotion rule to check.
//
// ruleType is the type of the promotion rule to check.
//
// cartID is the ID of the cart to check.
//
// user is the user who made the POS sale. If user is nil, the function will not check the customer level.
//
// It returns true if the POS sale is eligible for the promotion rule, or false otherwise.
//
// If an error occurs while checking the eligibility, it returns false and the error.
func (s *PromotionService) CheckPromotionEligibilityByPosSales(ruleID string, ruleType string, cartID string, user *models.UserModel) (bool, error) {
	var cart models.CartModel
	if err := s.db.Model(&models.CartModel{}).Preload("Items").Where("id = ?", cartID).First(&cart).Error; err != nil {
		return false, err
	}
	var rule models.PromotionRuleModel
	err := s.ctx.DB.Where("id = ? and rule_type = ?", ruleID, ruleType).Find(&rule).Error
	if err != nil {
		return false, err
	}

	switch rule.RuleType {
	case "MIN_PURCHASE":
		minPurchase, err := strconv.ParseFloat(rule.RuleValue, 64)
		if err != nil {
			return false, err
		}
		orderTotal := cart.Total

		if orderTotal < minPurchase {
			return false, nil
		}
	case "MAX_PURCHASE":
		maxPurchase, err := strconv.ParseFloat(rule.RuleValue, 64)
		if err != nil {
			return false, err
		}
		orderTotal := cart.Total
		if orderTotal > maxPurchase {
			return false, nil
		}
	case "CATEGORY":
		for _, v := range cart.Items {
			if v.CategoryID == nil {
				continue
			}

			if &rule.RuleValue == v.CategoryID {
				return true, nil
			}
		}
		return false, nil
	case "CATEGORIES":
		for _, v := range cart.Items {
			if v.CategoryID == nil {
				continue
			}

			if utils.ContainsString(strings.Split(rule.RuleValue, ","), *v.CategoryID) {
				return true, nil
			}
		}
		return false, nil
	case "PRODUCTS":
		for _, v := range cart.Items {
			if utils.ContainsString(strings.Split(rule.RuleValue, ","), v.ProductID) {
				return true, nil
			}
		}
		return false, nil
	case "CUSTOMER_LEVEL":
		if user != nil {
			if rule.RuleValue == *user.CustomerLevel {
				return false, nil
			}
		}
		return false, nil
	default:
		return false, errors.New("unknown rule type")
	}

	return true, nil
}

// CheckPromotionEligibility checks if a given condition is eligible for a promotion rule.
//
// promotionID is the ID of the promotion that the rule belongs to.
//
// ruleType is the type of the promotion rule to check.
//
// condition is the condition of the promotion rule to check.
//
// It returns true if the condition is eligible for the promotion rule, or false otherwise.
//
// If an error occurs while checking the eligibility, it returns false and the error.
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

// ApplyPromotion applies a promotion to an order.
//
// The function takes a promotion ID, order total, and cart items as parameters.
// It returns a PromotionResult, which contains the discount, final total, free shipping flag, free items, discount shipping rate, and discount shipping amount.
//
// The function first retrieves all actions associated with the promotion.
// It then applies each action to the order, calculating the discount, free shipping, and free items.
// The function finally returns the PromotionResult.
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

// CreatePromotion creates a new promotion in the database.
//
// It takes a pointer to a PromotionModel as a parameter and attempts to
// insert it into the database. The function returns an error if the
// creation of the promotion fails, otherwise it returns nil.

func (s *PromotionService) CreatePromotion(data *models.PromotionModel) error {
	return s.db.Create(data).Error
}

// UpdatePromotion updates a promotion in the database.
//
// It takes an ID of the promotion to be updated and a pointer to a PromotionModel
// as parameters. The function then attempts to update the promotion in the
// database. The function returns an error if the update of the promotion fails,
// otherwise it returns nil.
func (s *PromotionService) UpdatePromotion(id string, data *models.PromotionModel) error {
	return s.db.Where("id = ?", id).Updates(data).Error
}

// DeletePromotion deletes a promotion from the database.
//
// It takes the ID of the promotion to be deleted as a parameter. The function
// then attempts to delete the promotion from the database. The function returns
// an error if the deletion of the promotion fails, otherwise it returns nil.
func (s *PromotionService) DeletePromotion(id string) error {
	return s.db.Where("id = ?", id).Delete(&models.PromotionModel{}).Error
}

// GetPromotionByID retrieves a promotion from the database by its ID.
//
// It takes the ID of the promotion to be retrieved as a parameter. The function
// then attempts to retrieve the promotion from the database. The function
// returns the retrieved promotion and an error if the retrieval of the promotion
// fails.
func (s *PromotionService) GetPromotionByID(id string) (*models.PromotionModel, error) {
	var invoice models.PromotionModel
	err := s.db.Where("id = ?", id).
		Preload("Rules").
		Preload("Actions").
		First(&invoice).Error
	return &invoice, err
}

// GetPromotionByCode retrieves a promotion from the database by its code.
//
// It takes the code of the promotion to be retrieved as a parameter. The
// function then attempts to retrieve the promotion from the database. The
// function returns the retrieved promotion and an error if the retrieval of the
// promotion fails.
func (s *PromotionService) GetPromotionByCode(code string) (*models.PromotionModel, error) {
	var invoice models.PromotionModel
	err := s.db.Where("code = ?", code).First(&invoice).Error
	return &invoice, err
}

// AddRule adds a rule to a promotion in the database.
//
// It takes the ID of the promotion to add the rule to and a pointer to a
// PromotionRuleModel as parameters. The function then attempts to add the rule
// to the promotion in the database. The function returns an error if the
// addition of the rule fails, otherwise it returns nil.
func (s *PromotionService) AddRule(promotionID string, data *models.PromotionRuleModel) error {
	data.PromotionID = promotionID
	return s.db.Create(data).Error
}

// AddAction adds an action to a promotion in the database.
//
// It takes the ID of the promotion to add the action to and a pointer to a
// PromotionActionModel as parameters. The function then attempts to add the
// action to the promotion in the database. The function returns an error if the
// addition of the action fails, otherwise it returns nil.
func (s *PromotionService) AddAction(promotionID string, data *models.PromotionActionModel) error {
	data.PromotionID = promotionID
	return s.db.Create(data).Error
}

// DeleteRule deletes a rule from a promotion in the database.
//
// It takes the ID of the rule to be deleted as a parameter. The function then
// attempts to delete the rule from the promotion in the database. The function
// returns an error if the deletion of the rule fails, otherwise it returns nil.
func (s *PromotionService) DeleteRule(ruleID string) error {
	return s.db.Where("id = ?", ruleID).Delete(&models.PromotionRuleModel{}).Error
}

// DeleteAction deletes an action from a promotion in the database.
//
// It takes the ID of the action to be deleted as a parameter. The function then
// attempts to delete the action from the promotion in the database. The function
// returns an error if the deletion of the action fails, otherwise it returns nil.
func (s *PromotionService) DeleteAction(actionID string) error {
	return s.db.Where("id = ?", actionID).Delete(&models.PromotionActionModel{}).Error
}

// GetPromotions retrieves a paginated list of promotions from the database.
//
// It takes an http.Request and a search query string as input. The method uses
// GORM to query the database for promotions, applying the search query to
// the description and title fields. If the request contains a company ID
// header, the method also filters the result by the company ID. The function
// utilizes pagination to manage the result set and applies any necessary
// request modifications using the utils.FixRequest utility.
//
// The function returns a paginated page of PromotionModel and an error if
// the operation fails.

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

// GetUserPromotions retrieves a paginated list of active promotions from the database
// that are visible to the current user.
//
// It takes an http.Request and a search query string as input. The method uses
// GORM to query the database for promotions, applying the search query to
// the description and title fields. If the request contains a company ID
// header, the method also filters the result by the company ID. The function
// utilizes pagination to manage the result set and applies any necessary
// request modifications using the utils.FixRequest utility.
//
// The function returns a paginated page of PromotionModel and an error if
// the operation fails.
func (s *PromotionService) GetUserPromotions(request http.Request, search string) (paginate.Page, error) {
	pg := paginate.New()
	stmt := s.db.Preload("Rules").Preload("Actions")
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

// GetPromotionByName retrieves a promotion by its name from the database.
//
// If the promotion does not exist, a new promotion is created with the given name.
// It takes a string title as a parameter and returns a pointer to a PromotionModel
// and an error. If the operation to retrieve or create the promotion fails, an error
// is returned.
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
