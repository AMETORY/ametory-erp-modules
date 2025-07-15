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

// Migrate migrates the order request database tables.
//
// The order request database tables are:
// - order_requests
// - order_request_items
func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(&models.OrderRequestModel{}, &models.OrderRequestItemModel{})
}

// NewOrderRequestService creates a new OrderRequestService instance.
//
// This service is responsible for managing order request-related operations
// within the application. It requires a database connection, ERP context,
// merchant service, product service, and audit trail service.
func NewOrderRequestService(db *gorm.DB, ctx *context.ERPContext, merchantService *merchant.MerchantService, productService *product.ProductService, auditTrailSrv *audit_trail.AuditTrailService) *OrderRequestService {
	return &OrderRequestService{db: db, ctx: ctx, merchantService: merchantService, productService: productService, auditTrailService: auditTrailSrv}
}

// GetOrderRequestByID retrieves an order request by its ID.
//
// This method queries the database for an order request
// with the specified ID and preloads its associated items.
// If successful, it returns the order request model;
// otherwise, it returns an error.
//
// Params:
// - orderRequestID (string): The ID of the order request to retrieve.
//
// Returns:
// - (*models.OrderRequestModel): The order request model if found.
// - (error): An error object if the retrieval fails.

func (s *OrderRequestService) GetOrderRequestByID(orderRequestID string) (*models.OrderRequestModel, error) {
	orderRequest := models.OrderRequestModel{}
	err := s.db.Preload("Items").Where("id = ?", orderRequestID).First(&orderRequest).Error
	if err != nil {
		return nil, err
	}
	return &orderRequest, nil
}

// GetOrderRequestByUserIDWithStatus retrieves paginated order requests by user ID and status.
//
// This method queries the database for order requests with the specified user ID
// and status, preloading its associated items. If successful, it returns a paginated
// page of order requests; otherwise, it returns an error.
//
// Params:
// - request (http.Request): The HTTP request.
// - search (string): The search query string.
// - userID (string): The user ID.
// - status (string): The order request status.
//
// Returns:
// - (paginate.Page): The paginated page of order requests if found.
// - (error): An error object if the retrieval fails.
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

// GetOrderByStatus retrieves an order request with the specified user ID and status.
//
// This method queries the database for an order request with the specified user ID
// and status, preloading its associated items and offers. If successful, it returns
// the order request model; otherwise, it returns an error.
//
// Params:
// - userID (string): The user ID.
// - status ([]string): The order request statuses.
//
// Returns:
// - (*models.OrderRequestModel): The order request model if found.
// - (error): An error object if the retrieval fails.
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

// CreateOrderRequest creates a new order request for a specified user.
//
// This method first checks for any existing "PENDING" or "OFFERING"
// order requests for the user. If found, it updates the status of expired
// requests to "EXPIRED" and removes any associated offers. It then creates
// a new order request with the provided user ID, latitude, longitude, and
// expiration time, setting its initial status to "PENDING". The method
// logs the creation action using the audit trail service.
//
// Params:
// - userID (string): The ID of the user creating the order request.
// - userLat (float64): The latitude of the user's location.
// - userLng (float64): The longitude of the user's location.
// - expiresAt (time.Time): The expiration time of the order request.
//
// Returns:
// - (*models.OrderRequestModel): The newly created order request model.
// - (error): An error object if the creation fails, or nil if successful.

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

// AddOrderRequestItem adds a new item to an existing order request.
//
// This method first checks that the order request exists and is in the "PENDING"
// or "OFFERING" state. It then appends the provided item to the existing items
// associated with the order request. If the append operation fails, the method
// returns an error.
//
// Params:
// - orderRequestID (string): The ID of the order request to add the item to.
// - item (models.OrderRequestItemModel): The item to add to the order request.
//
// Returns:
// - (error): An error object if the append operation fails, or nil if successful.
func (s *OrderRequestService) AddOrderRequestItem(orderRequestID string, item models.OrderRequestItemModel) error {
	orderRequest := models.OrderRequestModel{}
	err := s.db.Where("id = ?", orderRequestID).First(&orderRequest).Error
	if err != nil {
		return err
	}

	return s.db.Model(&orderRequest).Association("Items").Append(&item)

}

// GetAvailableMerchant retrieves a list of nearby merchants for an order request.
//
// This method takes an order request ID and a maximum distance (in kilometers) as
// parameters, and returns a slice of MerchantModel representing the nearby
// merchants. If the order request does not exist, or if the retrieval fails, the
// method returns an error.
//
// The maximum distance parameter is in kilometers.
//
// Params:
// - orderRequestID (string): The ID of the order request to find nearby merchants for.
// - maxDistance (float64): The maximum distance in kilometers to search for merchants.
//
// Returns:
// - ([]models.MerchantModel): A slice of merchant models representing the nearby merchants.
// - (error): An error object if the retrieval fails, or nil if successful.
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

// FinishOrderRequest marks an order request as "Completed" and logs the action using the audit trail service.
//
// This method first checks that the order request exists and is in the "Accepted" state. If the order request
// is not in the "Accepted" state, the method returns an error. It then updates the status of the order request
// to "Completed" and logs the action using the audit trail service. If the update or logging fails, the method
// returns an error.
//
// Params:
// - orderRequestID (string): The ID of the order request to finish.
//
// Returns:
// - (error): An error object if the operation fails, or nil if successful.
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

// GetPendingOrderRequests retrieves a list of pending order requests associated with a merchant.
//
// This method takes a merchant ID as a parameter and queries the database to find all
// order requests with the status "Pending" and the specified merchant ID. It returns
// a slice of OrderRequestModel and an error if the operation fails.
//
// Params:
// - merchantID (string): The ID of the merchant to find pending order requests for.
//
// Returns:
// - ([]models.OrderRequestModel): A slice of order request models that are pending for the merchant.
// - (error): An error object if the retrieval fails, or nil if successful.
func (s *OrderRequestService) GetPendingOrderRequests(merchantID string) ([]models.OrderRequestModel, error) {
	var orderRequests []models.OrderRequestModel
	err := s.db.Where("status = ? AND merchant_id = ?", "Pending", merchantID).Find(&orderRequests).Error
	return orderRequests, err
}

// CancelOrderRequest cancels an order request by updating its status to "CANCELLED" and
// setting a cancellation reason. The method takes a user ID, order request ID, and
// cancellation reason as parameters.
//
// The method first checks that the order request exists and is in the "OFFERING"
// or "PENDING" state. If the order request is not in one of these states, the method
// returns an error. It then updates the status of the order request to "CANCELLED" and
// sets the cancellation reason. If the update fails, the method returns an error.
//
// Params:
// - userID (string): The ID of the user who is cancelling the order request.
// - orderRequestID (string): The ID of the order request to cancel.
// - reason (string): The reason for cancelling the order request.
//
// Returns:
// - (error): An error object if the update fails, or nil if successful.
func (s *OrderRequestService) CancelOrderRequest(userID, orderRequestID, reason string) error {
	return s.db.Model(&models.OrderRequestModel{}).Where("user_id = ? AND id = ? AND status IN (?)", userID, orderRequestID, []string{"OFFERING", "PENDING"}).
		Updates(map[string]interface{}{"status": "CANCELLED", "cancellation_reason": reason}).
		Error
}

// DeleteOrderRequest permanently deletes an order request by its ID.
//
// This method takes a user ID and order request ID as parameters and permanently
// deletes the order request from the database. If the deletion fails, the method
// returns an error.
//
// Params:
// - userID (string): The ID of the user who is deleting the order request.
// - orderRequestID (string): The ID of the order request to delete.
//
// Returns:
// - (error): An error object if the deletion fails, or nil if successful.
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
