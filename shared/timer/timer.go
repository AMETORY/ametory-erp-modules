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

func NewTimerService(orderRequestSvc *order_request.OrderRequestService) *TimerService {
	return &TimerService{orderRequestSvc: orderRequestSvc}
}

func (s *TimerService) StartOrderRequestTimer(orderRequestID string, timeout time.Duration) {
	time.AfterFunc(timeout, func() {
		orderRequest := models.OrderRequestModel{}
		err := s.db.Where("id = ?", orderRequestID).First(&orderRequest).Error
		if err != nil {
			return
		}

		if orderRequest.MerchantID == nil && orderRequest.Status == "Pending" {
			s.orderRequestSvc.CancelOrderRequest(orderRequestID, "Timeout")
		}
	})
}
