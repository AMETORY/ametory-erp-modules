package offering

import (
	"encoding/json"
	"fmt"

	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/shared/audit_trail"
	"github.com/AMETORY/ametory-erp-modules/shared/models"
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
	return &OfferingService{db: db, ctx: ctx, auditTrailService: auditTrailSrv}
}

func Migrate(tx *gorm.DB) error {
	return tx.AutoMigrate(&models.OfferModel{})
}

func (s *OfferingService) GetOffersForByIDs(userID string, offerIds []string) ([]models.OfferModel, error) {
	var offers []models.OfferModel
	err := s.db.Model(&models.OfferModel{}).Where("user_id = ? AND id IN (?)", userID, offerIds).Find(&offers).Error
	return offers, err
}

func (s *OfferingService) GetOffersForUser(userID, status string, orderRequestID *string) ([]models.OfferModel, error) {
	var offers []models.OfferModel
	db := s.db.Model(&models.OfferModel{})
	if orderRequestID != nil {
		db = db.Where("order_request_id = ?", *orderRequestID)
	}
	err := db.Find(&offers, "user_id = ? AND status = ?", userID, status).Error
	return offers, err
}

func (s *OfferingService) CreateOffer(merchant models.MerchantAvailableProduct, userID string) (*models.OfferModel, error) {
	if s.auditTrailService == nil {
		return nil, fmt.Errorf("audit trail service is not initialized")
	}
	b, _ := json.Marshal(merchant)

	offer := models.OfferModel{
		UserID:                       userID,
		OrderRequestID:               merchant.OrderRequestID,
		MerchantID:                   merchant.MerchantID,
		SubTotal:                     merchant.SubTotal,
		TotalPrice:                   merchant.TotalPrice,
		ShippingFee:                  merchant.ShippingFee,
		Distance:                     merchant.Distance,
		Status:                       "PENDING",
		MerchantAvailableProductData: string(b),
		MerchantAvailableProduct:     merchant,
	}
	err := s.db.Create(&offer).Error
	if err != nil {
		return nil, err
	}

	s.auditTrailService.LogAction(userID, "CREATE", "OFFER", offer.ID, fmt.Sprintf("{\"order_request_id\": \"%s\", \"merchant_id\": \"%s\"}", merchant.OrderRequestID, merchant.MerchantID))
	return &offer, nil
}

func (s *OfferingService) CreateOffers(orderRequest models.OrderRequestModel, availableMerchants []models.MerchantModel) error {
	if s.auditTrailService == nil {
		return fmt.Errorf("audit trail service is not initialized")
	}
	return s.db.Transaction(func(tx *gorm.DB) error {

		if orderRequest.Status == "Accepted" {
			return fmt.Errorf("order request is already accepted")
		}
		for _, merchant := range availableMerchants {
			offer := models.OfferModel{
				UserID:         orderRequest.UserID,
				OrderRequestID: orderRequest.ID,
				MerchantID:     merchant.ID,
				SubTotal:       orderRequest.SubTotal,
				TotalPrice:     orderRequest.TotalPrice,
				ShippingFee:    orderRequest.ShippingFee,
				Status:         "PENDING",
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
		offer := models.OfferModel{}
		err := tx.Where("id = ?", offerID).First(&offer).Error
		if err != nil {
			return err
		}
		if offer.Status == "Taken" {
			return fmt.Errorf("offer is already taken")
		}

		orderRequest := models.OrderRequestModel{}
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
// 		orderRequest := models.OrderRequestModel{}
// 		err := tx.Where("id = ?", orderRequestID).First(&orderRequest).Error
// 		if err != nil {
// 			return err
// 		}
// 		if orderRequest.Status == "Accepted" {
// 			return fmt.Errorf("order request is already accepted")
// 		}

// 		offer := models.OfferModel{
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
// func (s *OfferingService) GetOffers(productIDs []string, userLat, userLng float64, maxDistance float64) ([]models.OfferModel, error) {
// 	// Cari merchant terdekat
// 	merchants, err := s.merchantService.GetNearbyMerchants(userLat, userLng, maxDistance) // Radius 10 km
// 	if err != nil {
// 		return nil, err
// 	}

// 	var offers []models.OfferModel
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
// 				// offer := models.OfferModel{
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
