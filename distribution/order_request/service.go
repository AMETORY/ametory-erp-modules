package order_request

import (
	"fmt"
	"net/http"
	"time"

	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/inventory/product"
	"github.com/AMETORY/ametory-erp-modules/order/merchant"
	"github.com/AMETORY/ametory-erp-modules/shared/audit_trail"
	"github.com/AMETORY/ametory-erp-modules/shared/models"
	"github.com/AMETORY/ametory-erp-modules/utils"
	"github.com/morkid/paginate"
	"gorm.io/gorm"
)

type OrderRequestService struct {
	db                *gorm.DB
	ctx               *context.ERPContext
	merchantService   *merchant.MerchantService
	productService    *product.ProductService
	auditTrailService *audit_trail.AuditTrailService
}

func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(&models.OrderRequestModel{}, &models.OrderRequestItemModel{})
}

func NewOrderRequestService(db *gorm.DB, ctx *context.ERPContext, merchantService *merchant.MerchantService, productService *product.ProductService, auditTrailSrv *audit_trail.AuditTrailService) *OrderRequestService {
	return &OrderRequestService{db: db, ctx: ctx, merchantService: merchantService, productService: productService, auditTrailService: auditTrailSrv}
}

func (s *OrderRequestService) GetOrderRequestByID(orderRequestID string) (*models.OrderRequestModel, error) {
	orderRequest := models.OrderRequestModel{}
	err := s.db.Preload("Items").Where("id = ?", orderRequestID).First(&orderRequest).Error
	if err != nil {
		return nil, err
	}
	return &orderRequest, nil
}

func (s *OrderRequestService) GetOrderRequestByUserIDWithStatus(request http.Request, search string, userID string, status string) (paginate.Page, error) {
	pg := paginate.New()
	// pendingRequest := []models.OrderRequestModel{}
	stmt := s.db.Preload("Offers").Preload("Items", func(db *gorm.DB) *gorm.DB {
		return db.Preload("Product", func(db *gorm.DB) *gorm.DB {
			return db.Select("display_name", "id", "category_id", "brand_id").Preload("Category").Preload("Brand")
		}).Preload("Variant", func(db *gorm.DB) *gorm.DB {
			return db.Select("id", "display_name")
		})
	}).Where("user_id = ? AND status = ?", userID, status)
	stmt = stmt.Model(&models.OrderRequestModel{})
	utils.FixRequest(&request)
	page := pg.With(stmt).Request(request).Response(&[]models.OrderRequestModel{})
	page.Page = page.Page + 1

	return page, nil
}
func (s *OrderRequestService) GetOrderByStatus(userID string, status []string) (*models.OrderRequestModel, error) {
	pendingRequest := models.OrderRequestModel{}
	err := s.db.Preload("Offers").Preload("Items", func(db *gorm.DB) *gorm.DB {
		return db.Preload("Product", func(db *gorm.DB) *gorm.DB {
			return db.Select("display_name", "id", "category_id", "brand_id").Preload("Category").Preload("Brand")
		}).Preload("Variant", func(db *gorm.DB) *gorm.DB {
			return db.Select("id", "display_name")
		})
	}).Where("user_id = ? AND status in (?)", userID, status).First(&pendingRequest).Error
	return &pendingRequest, err
}

func (s *OrderRequestService) CreateOrderRequest(userID string, userLat, userLng float64, expiresAt time.Time) (*models.OrderRequestModel, error) {
	pendingRequest := models.OrderRequestModel{}
	err := s.db.Where("user_id = ? AND status in (?)", userID, []string{"PENDING", "OFFERING"}).First(&pendingRequest).Error
	if err == nil {
		if pendingRequest.ExpiresAt.Before(time.Now()) {
			pendingRequest.Status = "EXPIRED"
			if err := s.db.Save(&pendingRequest).Error; err != nil {
				return nil, err
			}
		}
		if pendingRequest.Status == "OFFERING" {
			s.db.Where("id = ?", pendingRequest.ID).Unscoped().Delete(&models.OfferModel{})
		}
	}

	if s.auditTrailService == nil {
		return nil, fmt.Errorf("audit trail service is not initialized")
	}
	orderRequest := models.OrderRequestModel{
		UserID:       userID,
		UserLat:      userLat,
		UserLng:      userLng,
		Status:       "PENDING",
		ExpiresAt:    expiresAt,
		ShippingData: "{}",
	}
	if err := s.db.Create(&orderRequest).Error; err != nil {
		return nil, err
	}
	s.auditTrailService.LogAction(userID, "CREATE", "ORDER_REQUEST", orderRequest.ID, "{}")
	return &orderRequest, nil
}

func (s *OrderRequestService) AddOrderRequestItem(orderRequestID string, item models.OrderRequestItemModel) error {
	orderRequest := models.OrderRequestModel{}
	err := s.db.Where("id = ?", orderRequestID).First(&orderRequest).Error
	if err != nil {
		return err
	}

	return s.db.Model(&orderRequest).Association("Items").Append(&item)

}

func (s *OrderRequestService) GetAvailableMerchant(orderRequestID string, maxDistance float64) ([]models.MerchantModel, error) {

	orderRequest := models.OrderRequestModel{}
	err := s.db.Where("id = ?", orderRequestID).First(&orderRequest).Error
	if err != nil {
		return nil, err
	}

	merchants, err := s.merchantService.GetNearbyMerchants(orderRequest.UserLat, orderRequest.UserLng, maxDistance) // in km
	if err != nil {
		return nil, err
	}

	return merchants, nil
}

// FinishOrderRequest digunakan untuk mengupdate status order request menjadi "Completed"
// Jika order request tidak dalam status "Accepted", maka akan mengembalikan error
func (s *OrderRequestService) FinishOrderRequest(orderRequestID string) error {
	if s.auditTrailService == nil {
		return fmt.Errorf("audit trail service is not initialized")
	}
	orderRequest := models.OrderRequestModel{}
	err := s.db.Where("id = ?", orderRequestID).First(&orderRequest).Error
	if err != nil {
		return err
	}

	// Order request hanya dapat di finish jika statusnya adalah "Accepted"
	if orderRequest.Status != "Accepted" {
		return fmt.Errorf("order request is not in an accepted state")
	}

	// Update status order request menjadi "Completed"
	orderRequest.Status = "Completed"
	s.auditTrailService.LogAction(orderRequest.UserID, "FINISH", "ORDER_REQUEST", orderRequest.ID, "{}")
	return s.db.Save(&orderRequest).Error
}

// GetPendingOrderRequests digunakan untuk mendapatkan order request yang statusnya "Pending" dan merchant_id sesuai dengan parameter
// Fungsi ini akan mengembalikan slice of models.OrderRequestModel
func (s *OrderRequestService) GetPendingOrderRequests(merchantID string) ([]models.OrderRequestModel, error) {
	var orderRequests []models.OrderRequestModel
	err := s.db.Where("status = ? AND merchant_id = ?", "Pending", merchantID).Find(&orderRequests).Error
	return orderRequests, err
}

func (s *OrderRequestService) CancelOrderRequest(userID, orderRequestID, reason string) error {
	return s.db.Model(&models.OrderRequestModel{}).Where("user_id = ? AND id = ? AND status IN (?)", userID, orderRequestID, []string{"OFFERING", "PENDING"}).
		Updates(map[string]interface{}{"status": "CANCELLED", "cancellation_reason": reason}).
		Error
}
func (s *OrderRequestService) DeleteOrderRequest(userID, orderRequestID string) error {
	return s.db.Model(&models.OrderRequestModel{}).Where("user_id = ? AND id = ?", userID, orderRequestID).Unscoped().Delete(&models.OrderRequestModel{}).Error
}

// func (s *OrderRequestService) AcceptOrderRequest(orderRequestID, merchantID, offerID string, totalPrice, shippingFee float64) error {
// 	tx := s.db.Begin()
// 	defer func() {
// 		if r := recover(); r != nil {
// 			tx.Rollback()
// 		}
// 	}()

// 	if err := tx.Model(&models.OrderRequestModel{}).Where("id = ?", orderRequestID).Updates(map[string]interface{}{
// 		"status":       "Accepted",
// 		"merchant_id":  merchantID,
// 		"total_price":  totalPrice,
// 		"shipping_fee": shippingFee,
// 		"offer_id":     offerID,
// 	}).Error; err != nil {
// 		tx.Rollback()
// 		return err
// 	}

// 	if err := tx.Table("offers").Where("id = ?", offerID).Update("status", "Taken").Error; err != nil {
// 		tx.Rollback()
// 		return err
// 	}

// 	return tx.Commit().Error
// }
