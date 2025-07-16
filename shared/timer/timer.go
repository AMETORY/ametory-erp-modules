package timer

import (
	"time"

	"github.com/AMETORY/ametory-erp-modules/distribution/order_request"
	"github.com/AMETORY/ametory-erp-modules/shared/models"
	"gorm.io/gorm"
)

type TimerService struct {
	db              *gorm.DB
	orderRequestSvc *order_request.OrderRequestService
}

// NewTimerService creates a new instance of TimerService with the given
// order request service.
func NewTimerService(orderRequestSvc *order_request.OrderRequestService) *TimerService {
	return &TimerService{orderRequestSvc: orderRequestSvc}
}

// StartOrderRequestTimer sets a timer for an order request.
//
// This method initiates a countdown for the specified order request ID. If the
// order request remains in the "Pending" status and is not claimed by any
// merchant (MerchantID is nil) by the end of the timeout duration, the order
// request will be automatically cancelled with a "Timeout" reason.
//
// Params:
//   - orderRequestID (string): The ID of the order request to set the timer for.
//   - timeout (time.Duration): The duration after which the order request should
//     be cancelled if unclaimed.
func (s *TimerService) StartOrderRequestTimer(orderRequestID string, timeout time.Duration) {
	time.AfterFunc(timeout, func() {
		orderRequest := models.OrderRequestModel{}
		err := s.db.Where("id = ?", orderRequestID).First(&orderRequest).Error
		if err != nil {
			return
		}

		if orderRequest.MerchantID == nil && orderRequest.Status == "Pending" {
			s.orderRequestSvc.CancelOrderRequest(orderRequest.UserID, orderRequestID, "Timeout")
		}
	})
}
