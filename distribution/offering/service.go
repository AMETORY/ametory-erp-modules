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

// NewOfferingService creates a new instance of OfferingService.
//
// The service provides methods for managing offerings, requiring a GORM
// database instance for CRUD operations, an ERP context for authentication
// and request handling, and an AuditTrailService for tracking changes.

func NewOfferingService(db *gorm.DB, ctx *context.ERPContext, auditTrailSrv *audit_trail.AuditTrailService) *OfferingService {
	return &OfferingService{db: db, ctx: ctx, auditTrailService: auditTrailSrv}
}

// Migrate performs the auto-migration for the OfferModel schema.
//
// This function uses the provided GORM database transaction to
// automatically migrate the database schema for the OfferModel.
// It ensures that the database table structure is up-to-date
// with the current model definition.
//
// Params:
// - tx (*gorm.DB): The database transaction to be used for the migration.
//
// Returns:
// - error: An error object if the migration fails, otherwise nil.

func Migrate(tx *gorm.DB) error {
	return tx.AutoMigrate(&models.OfferModel{})
}

// GetOffersForByIDs retrieves a list of offers for a specified user and offer IDs.
//
// This method queries the database to find all offers associated with the
// provided user ID and a list of offer IDs. It returns a slice of OfferModel
// and an error if the operation fails.
//
// Params:
// - userID (string): The ID of the user whose offers are to be retrieved.
// - offerIds ([]string): A slice of offer IDs to filter the offers.
//
// Returns:
// - ([]models.OfferModel): A slice of offer models that match the criteria.
// - (error): An error object if the database query fails, otherwise nil.

func (s *OfferingService) GetOffersForByIDs(userID string, offerIds []string) ([]models.OfferModel, error) {
	var offers []models.OfferModel
	err := s.db.Model(&models.OfferModel{}).Where("user_id = ? AND id IN (?)", userID, offerIds).Find(&offers).Error
	return offers, err
}

// GetOffersForUser retrieves a list of offers for a specified user ID, status, and order request ID.
//
// This method queries the database to find all offers associated with the provided user ID,
// status, and order request ID. It returns a slice of OfferModel and an error if the operation fails.
//
// Params:
// - userID (string): The ID of the user whose offers are to be retrieved.
// - status (string): The status of the offers to be retrieved.
// - orderRequestID (*string): An optional order request ID to filter the offers.
//
// Returns:
// - ([]models.OfferModel): A slice of offer models that match the criteria.
// - (error): An error object if the database query fails, otherwise nil.
func (s *OfferingService) GetOffersForUser(userID, status string, orderRequestID *string) ([]models.OfferModel, error) {
	var offers []models.OfferModel
	db := s.db.Model(&models.OfferModel{})
	if orderRequestID != nil {
		db = db.Where("order_request_id = ?", *orderRequestID)
	}
	err := db.Find(&offers, "user_id = ? AND status = ?", userID, status).Error
	return offers, err
}

// CreateOffer creates a new offer in the system.
//
// This method takes a MerchantAvailableProduct and a user ID as parameters,
// constructs an OfferModel with the provided information, and persists it to
// the database. The offer's status is initially set to "PENDING". The method
// also logs the action using the audit trail service.
//
// Params:
// - merchant (models.MerchantAvailableProduct): The merchant information
//   and product details for the offer.
// - userID (string): The ID of the user for whom the offer is being created.
//
// Returns:
// - (*models.OfferModel): The created offer model, or nil if an error occurs.
// - (error): An error object if the creation fails, or nil if the operation is successful.

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
		SubTotalBeforeDiscount:       merchant.SubTotalBeforeDiscount,
		TotalDiscountAmount:          merchant.TotalDiscountAmount,
		TotalPrice:                   merchant.TotalPrice,
		ServiceFee:                   merchant.ServiceFee,
		ShippingFee:                  merchant.ShippingFee,
		ShippingType:                 merchant.ShippingType,
		CourierName:                  merchant.CourierName,
		Distance:                     merchant.Distance,
		Tax:                          merchant.Tax,
		TaxType:                      merchant.TaxType,
		TaxAmount:                    merchant.TaxAmount,
		TotalTaxAmount:               merchant.TotalTaxAmount,
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

// CreateOffers creates offers for an order request from a list of available merchants.
//
// This method takes an OrderRequestModel and a slice of MerchantAvailableProduct
// as parameters, constructs an OfferModel for each merchant, and persists them to
// the database. The offer's status is initially set to "PENDING". The method
// also logs the action using the audit trail service.
//
// If the order request is already accepted, the method returns an error.
//
// Params:
//   - orderRequest (models.OrderRequestModel): The order request for which offers will be created.
//   - availableMerchants ([]models.MerchantAvailableProduct): A slice of merchant information
//     and product details for the offers.
//
// Returns:
// - (error): An error object if the creation fails, or nil if the operation is successful.
func (s *OfferingService) CreateOffers(orderRequest models.OrderRequestModel, availableMerchants []models.MerchantAvailableProduct) error {
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
				MerchantID:     merchant.MerchantID,
				SubTotal:       orderRequest.SubTotal,
				TotalPrice:     orderRequest.TotalPrice,
				ShippingFee:    orderRequest.ShippingFee,
				ShippingType:   merchant.ShippingType,
				CourierName:    merchant.CourierName,
				Tax:            merchant.Tax,
				TaxType:        merchant.TaxType,
				TaxAmount:      merchant.TaxAmount,
				TotalTaxAmount: merchant.TotalTaxAmount,
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

// TakeOffer takes an offer and updates the order request status to Accepted.
//
// This method takes an offer ID, total price, shipping fee, and distance as parameters.
// It first checks if the offer is already taken, and returns an error if it is.
// Then it verifies the order request status and returns an error if the order request
// is not in the pending status.
//
// If the offer is taken successfully, the method also updates the order request status
// to Accepted and sets the merchant ID.
//
// Params:
// - offerID (string): The ID of the offer to be taken.
// - totalPrice (float64): The total price of the offer.
// - shippingFee (float64): The shipping fee of the offer.
// - distance (float64): The distance of the offer.
//
// Returns:
// - (error): An error object if the taking fails, or nil if the operation is successful.
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
