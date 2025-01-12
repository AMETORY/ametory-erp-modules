package order_request

import (
	"fmt"
	"time"

	"github.com/AMETORY/ametory-erp-modules/inventory/product"
	"github.com/AMETORY/ametory-erp-modules/order/merchant"
	"github.com/AMETORY/ametory-erp-modules/shared/audit_trail"
	"gorm.io/gorm"
)

type OrderRequestService struct {
	db                *gorm.DB
	merchantService   *merchant.MerchantService
	productService    *product.ProductService
	auditTrailService *audit_trail.AuditTrailService
}

func NewOrderRequestService(db *gorm.DB, merchantService *merchant.MerchantService, productService *product.ProductService, auditTrailSrv *audit_trail.AuditTrailService) *OrderRequestService {
	return &OrderRequestService{db: db, merchantService: merchantService, productService: productService, auditTrailService: auditTrailSrv}
}

func (s *OrderRequestService) CreateOrderRequest(userID string, userLat, userLng float64, expiresAt time.Time) (*OrderRequestModel, error) {
	if s.auditTrailService == nil {
		return nil, fmt.Errorf("audit trail service is not initialized")
	}
	orderRequest := OrderRequestModel{
		UserID:    userID,
		UserLat:   userLat,
		UserLng:   userLng,
		Status:    "Pending",
		ExpiresAt: expiresAt,
	}
	if err := s.db.Create(&orderRequest).Error; err != nil {
		return nil, err
	}
	s.auditTrailService.LogAction(userID, "CREATE", "ORDER_REQUEST", orderRequest.ID, "{}")
	return &orderRequest, nil
}

func (s *OrderRequestService) GetAvailableMerchant(orderRequestID string, maxDistance float64) ([]merchant.MerchantModel, error) {

	orderRequest := OrderRequestModel{}
	err := s.db.Where("id = ?", orderRequestID).First(&orderRequest).Error
	if err != nil {
		return nil, err
	}

	merchants, err := s.merchantService.GetNearbyMerchants(orderRequest.UserLat, orderRequest.UserLng, maxDistance)
	if err != nil {
		return nil, err
	}
	productIDs := []string{}
	for _, v := range orderRequest.Items {
		productIDs = append(productIDs, *v.ProductID)

	}
	availableMerchants := []merchant.MerchantModel{}
	for _, merchant := range merchants {
		// Dapatkan produk dari merchant
		products, err := s.productService.GetProductsByMerchant(merchant.ID, productIDs)
		if err != nil {
			return nil, err
		}
		if len(products) == len(productIDs) {
			availableMerchants = append(availableMerchants, merchant)
		}
	}

	return availableMerchants, nil
}

// FinishOrderRequest digunakan untuk mengupdate status order request menjadi "Completed"
// Jika order request tidak dalam status "Accepted", maka akan mengembalikan error
func (s *OrderRequestService) FinishOrderRequest(orderRequestID string) error {
	if s.auditTrailService == nil {
		return fmt.Errorf("audit trail service is not initialized")
	}
	orderRequest := OrderRequestModel{}
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
// Fungsi ini akan mengembalikan slice of OrderRequestModel
func (s *OrderRequestService) GetPendingOrderRequests(merchantID string) ([]OrderRequestModel, error) {
	var orderRequests []OrderRequestModel
	err := s.db.Where("status = ? AND merchant_id = ?", "Pending", merchantID).Find(&orderRequests).Error
	return orderRequests, err
}

func (s *OrderRequestService) CancelOrderRequest(orderRequestID string, reason string) error {
	return s.db.Model(&OrderRequestModel{}).Where("id = ?", orderRequestID).
		Updates(map[string]interface{}{"status": "Cancelled", "cancellation_reason": reason}).
		Error
}

// func (s *OrderRequestService) AcceptOrderRequest(orderRequestID, merchantID, offerID string, totalPrice, shippingFee float64) error {
// 	tx := s.db.Begin()
// 	defer func() {
// 		if r := recover(); r != nil {
// 			tx.Rollback()
// 		}
// 	}()

// 	if err := tx.Model(&OrderRequestModel{}).Where("id = ?", orderRequestID).Updates(map[string]interface{}{
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
