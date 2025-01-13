package offering

import (
	"fmt"

	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/distribution/order_request"
	"github.com/AMETORY/ametory-erp-modules/order/merchant"
	"github.com/AMETORY/ametory-erp-modules/shared/audit_trail"
	"gorm.io/gorm"
)

type OfferingService struct {
	db                *gorm.DB
	ctx               *context.ERPContext
	auditTrailService *audit_trail.AuditTrailService
	// merchantService   *merchant.MerchantService
	// productService    *product.ProductService
}

func NewOfferingService(db *gorm.DB, ctx *context.ERPContext, auditTrailSrv *audit_trail.AuditTrailService) *OfferingService {
	return &OfferingService{db: db, ctx: ctx}
}

func (s *OfferingService) GetOffersForUser(userID, status string, orderRequest *string) ([]OfferModel, error) {
	var offers []OfferModel
	db := s.db.Model(&OfferModel{})
	if orderRequest != nil {
		db = db.Where("order_request_id = ?", *orderRequest)
	}
	err := db.Find(&offers, "user_id = ? AND status = ?", userID, status).Error
	return offers, err
}

func (s *OfferingService) CreateOffer(orderRequestID string, merchant merchant.MerchantModel, shippingFee float64, userID string) error {
	if s.auditTrailService == nil {
		return fmt.Errorf("audit trail service is not initialized")
	}
	orderRequest := order_request.OrderRequestModel{}
	err := s.db.Where("id = ?", orderRequestID).First(&orderRequest).Error
	if err != nil {
		return err
	}
	offer := OfferModel{
		UserID:         userID,
		OrderRequestID: orderRequestID,
		MerchantID:     merchant.ID,
		TotalPrice:     orderRequest.TotalPrice,
		ShippingFee:    shippingFee,
		Status:         "Pending",
	}
	err = s.db.Create(&offer).Error
	if err != nil {
		return err
	}
	s.auditTrailService.LogAction(userID, "CREATE", "OFFER", offer.ID, fmt.Sprintf("{\"order_request_id\": \"%s\", \"merchant_id\": \"%s\"}", orderRequestID, merchant.ID))
	return nil
}

func (s *OfferingService) CreateOffers(orderRequestID string, availableMerchants []merchant.MerchantModel) error {
	if s.auditTrailService == nil {
		return fmt.Errorf("audit trail service is not initialized")
	}
	return s.db.Transaction(func(tx *gorm.DB) error {
		orderRequest := order_request.OrderRequestModel{}
		err := tx.Where("id = ?", orderRequestID).First(&orderRequest).Error
		if err != nil {
			return err
		}
		if orderRequest.Status == "Accepted" {
			return fmt.Errorf("order request is already accepted")
		}
		for _, merchant := range availableMerchants {
			offer := OfferModel{
				UserID:         orderRequestID,
				OrderRequestID: orderRequestID,
				MerchantID:     merchant.ID,
				TotalPrice:     orderRequest.TotalPrice,
				ShippingFee:    orderRequest.ShippingFee,
				Status:         "Pending",
			}
			err := tx.Create(&offer).Error
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func (s *OfferingService) TakeOffer(offerID string, totalPrice, shippingFee, distance float64) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		offer := OfferModel{}
		err := tx.Where("id = ?", offerID).First(&offer).Error
		if err != nil {
			return err
		}
		if offer.Status == "Taken" {
			return fmt.Errorf("offer is already taken")
		}

		orderRequest := order_request.OrderRequestModel{}
		err = tx.Where("id = ?", offer.OrderRequestID).First(&orderRequest).Error
		if err != nil {
			return err
		}
		switch orderRequest.Status {
		case "Accepted":
			return fmt.Errorf("order request is already accepted")
		case "Rejected":
			return fmt.Errorf("order request is already rejected")
		case "Cancelled":
			return fmt.Errorf("order request is already cancelled")
		case "Completed":
			return fmt.Errorf("order request is already completed")
		}

		offer.TotalPrice = totalPrice
		offer.ShippingFee = shippingFee
		offer.Distance = distance

		offer.Status = "Taken"
		err = tx.Save(&offer).Error
		if err != nil {
			return err
		}

		orderRequest.Status = "Accepted"
		orderRequest.MerchantID = &offer.MerchantID
		err = tx.Save(&orderRequest).Error
		if err != nil {
			return err
		}

		return nil
	})
}

// func (s *OfferingService) TakeOrder(orderRequestID string, merchantID string) error {
// 	return s.db.Transaction(func(tx *gorm.DB) error {
// 		orderRequest := order_request.OrderRequestModel{}
// 		err := tx.Where("id = ?", orderRequestID).First(&orderRequest).Error
// 		if err != nil {
// 			return err
// 		}
// 		if orderRequest.Status == "Accepted" {
// 			return fmt.Errorf("order request is already accepted")
// 		}

// 		offer := OfferModel{
// 			UserID:         orderRequest.UserID,
// 			OrderRequestID: orderRequest.ID,
// 			MerchantID:     merchantID,
// 			TotalPrice:     orderRequest.TotalPrice,
// 			ShippingFee:    orderRequest.ShippingFee,
// 			Status:         "Accepted",
// 		}
// 		err = tx.Create(&offer).Error
// 		if err != nil {
// 			return err
// 		}
// 		return nil
// 	})
// }

// max distance of 10 km
// func (s *OfferingService) GetOffers(productIDs []string, userLat, userLng float64, maxDistance float64) ([]OfferModel, error) {
// 	// Cari merchant terdekat
// 	merchants, err := s.merchantService.GetNearbyMerchants(userLat, userLng, maxDistance) // Radius 10 km
// 	if err != nil {
// 		return nil, err
// 	}

// 	var offers []OfferModel
// 	for _, merchant := range merchants {
// 		// Dapatkan produk dari merchant
// 		products, err := s.productService.GetProductsByMerchant(merchant.ID, []string{})
// 		if err != nil {
// 			return nil, err
// 		}

// 		// Hitung jarak
// 		distance := location.Haversine(userLat, userLng, merchant.Latitude, merchant.Longitude)
// 		fmt.Println("Jarak", distance)
// 		// Hitung harga penawaran
// 		for _, product := range products {
// 			if contains(productIDs, product.ID) {
// 				// offer := OfferModel{
// 				// 	MerchantID: merchant.ID,
// 				// 	ProductID:  product.ID,
// 				// 	Price:      s.pricingEngine.CalculateOffer(product.Price, distance),
// 				// 	Distance:   distance,
// 				// }
// 				// offers = append(offers, offer)
// 			}
// 		}
// 	}

// 	return offers, nil
// }

func contains(arr []string, item string) bool {
	for _, a := range arr {
		if a == item {
			return true
		}
	}
	return false
}
