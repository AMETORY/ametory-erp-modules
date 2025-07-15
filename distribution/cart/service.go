package cart

import (
	"errors"
	"fmt"

	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/inventory"
	"github.com/AMETORY/ametory-erp-modules/shared/models"
	"gorm.io/gorm"
)

type CartService struct {
	db               *gorm.DB
	ctx              *context.ERPContext
	inventoryService *inventory.InventoryService
	merchantID       *string
}

// NewCartService creates a new CartService instance.
//
// It initializes the CartService with the given database connection, ERP context,
// and inventory service. This service is responsible for managing cart-related
// operations within the application.

func NewCartService(db *gorm.DB, ctx *context.ERPContext, inventoryService *inventory.InventoryService) *CartService {
	return &CartService{db: db, ctx: ctx, inventoryService: inventoryService}
}

// Migrate performs the automatic migration for cart-related database models.
//
// It ensures that the database schema for CartModel and CartItemModel
// is up-to-date, creating or modifying the tables as necessary.

func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(&models.CartModel{}, &models.CartItemModel{})
}

// SetMerchantID sets the merchant ID associated with the cart service.
//
// This is a mandatory configuration step before using the cart service.
// It sets the merchant ID to be used for looking up products and variants
// when adding items to the cart.
func (s *CartService) SetMerchantID(merchantID string) {
	s.merchantID = &merchantID
}

// GetCartByID returns the cart with the given ID, preloaded with all its items.
//
// The function will return an error if the cart is not found or if there is a
// database error. The returned cart will have the subtotal, discount amount,
// and subtotal before discount calculated.
func (s *CartService) GetCartByID(cartID string) (*models.CartModel, error) {
	var cart models.CartModel
	err := s.db.Preload("Items").Where("status = ?", "ACTIVE").Where("id = ?", cartID).First(&cart).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("cart not found")
		}
		return nil, err
	}
	subTotal, _ := s.CountSubTotalByCartID(cartID)
	cart.SubTotal = subTotal
	discountAmount := float64(0)
	for i, v := range cart.Items {
		product := models.ProductModel{}
		if err := s.db.Model(&product).Where("id = ?", v.ProductID).First(&product).Error; err != nil {
			return nil, err
		}
		product.GetPriceAndDiscount(s.db)
		var adjustmentPrice, originalPrice float64 = product.AdjustmentPrice, product.OriginalPrice
		if v.VariantID != nil {
			var variant models.VariantModel
			variant.MerchantID = s.merchantID
			if err := s.db.Preload("Tags").Where("id = ?", v.VariantID).First(&variant).Error; err != nil {
				return nil, err
			}
			v.Variant = &variant
			variant.GetPriceAndDiscount(s.db)
			adjustmentPrice = variant.AdjustmentPrice
			originalPrice = variant.OriginalPrice
		}
		s.parseItem(&v, originalPrice, adjustmentPrice)
		discountAmount += v.Quantity * v.DiscountAmount
		cart.Items[i] = v
		cart.SubTotalBeforeDiscount += v.Quantity * v.OriginalPrice
	}
	cart.CustomerData = "{}"
	cart.DiscountAmount = discountAmount
	return &cart, nil
}

// ClearActiveCart deletes all items in the active cart of the given user.
//
// If there is no active cart, one will be created first.
//
// The function returns an error if there is a database error.
func (s *CartService) ClearActiveCart(userID string) error {
	var cart models.CartModel
	err := s.db.Preload("Items").Where("user_id = ? AND status = ?", userID, "ACTIVE").First(&cart).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			cart = models.CartModel{
				UserID: userID,
				Status: "ACTIVE",
			}
			if err := s.db.Create(&cart).Error; err != nil {
				return err
			}
		} else {
			return err
		}
	}

	for _, v := range cart.Items {
		s.DeleteItemCart(userID, v.ID)
	}
	return nil
}

// GetOrCreateActiveCart returns the active cart of the given user ID, or creates
// one if it does not exist.
//
// The function returns the cart with all its items preloaded, and with the
// subtotal, discount amount, and subtotal before discount calculated.
//
// The function returns an error if there is a database error.
func (s *CartService) GetOrCreateActiveCart(userID string) (*models.CartModel, error) {
	var cart models.CartModel
	err := s.db.Preload("Items").Where("user_id = ? AND status = ?", userID, "ACTIVE").First(&cart).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			cart = models.CartModel{
				UserID: userID,
				Status: "ACTIVE",
			}
			if err := s.db.Create(&cart).Error; err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}
	subTotal, _ := s.CountSubTotal(userID)
	cart.SubTotal = subTotal
	discountAmount := float64(0)
	for i, v := range cart.Items {
		product := models.ProductModel{}
		if err := s.db.Preload("Category").Preload("Brand").Model(&product).Where("id = ?", v.ProductID).First(&product).Error; err != nil {
			return nil, err
		}
		product.GetPriceAndDiscount(s.db)
		var adjustmentPrice, originalPrice float64 = product.AdjustmentPrice, product.OriginalPrice
		if v.VariantID != nil {
			var variant models.VariantModel
			variant.MerchantID = s.merchantID
			if err := s.db.Preload("Tags").Where("id = ?", v.VariantID).First(&variant).Error; err != nil {
				return nil, err
			}
			v.Variant = &variant
			variant.GetPriceAndDiscount(s.db)
			adjustmentPrice = variant.AdjustmentPrice
			originalPrice = variant.OriginalPrice
		}
		s.parseItem(&v, originalPrice, adjustmentPrice)
		discountAmount += v.Quantity * v.DiscountAmount
		v.Category = product.Category
		v.Brand = product.Brand
		cart.Items[i] = v
		cart.SubTotalBeforeDiscount += v.Quantity * v.OriginalPrice
	}
	cart.CustomerData = "{}"
	cart.DiscountAmount = discountAmount
	return &cart, nil
}

// parseItem is a helper function that parses a cart item model.
// It loads the product images, sets the display name, original price, discount amount, discount rate, and discount type.
// It also calculates the subtotal and subtotal before discount, and sets the original price and adjustment price.
// If the item has a variant, it sets the display name and original price from the variant.
func (s *CartService) parseItem(v *models.CartItemModel, originalPrice, adjustmentPrice float64) {
	img, _ := s.inventoryService.ProductService.ListImagesOfProduct(v.ProductID)
	v.OriginalPrice = v.Product.Price
	v.ProductImages = img
	v.DisplayName = v.Product.DisplayName
	if v.Product.ActiveDiscount != nil {
		v.ActiveDiscount = v.Product.ActiveDiscount
	}

	v.DiscountAmount = v.Product.DiscountAmount
	v.DiscountRate = v.Product.DiscountRate
	v.DiscountType = v.Product.DiscountType
	if v.VariantID != nil {
		v.DisplayName = v.Variant.DisplayName
		// fmt.Println("ORIGINAL PRICE", v.Variant.OriginalPrice)
		v.OriginalPrice = v.Variant.OriginalPrice
		v.DiscountAmount = v.Variant.DiscountAmount
		v.DiscountRate = v.Variant.DiscountRate
		v.DiscountType = v.Variant.DiscountType
		if v.Variant.ActiveDiscount != nil {
			v.ActiveDiscount = v.Variant.ActiveDiscount
		}
	}

	v.SubTotal = v.Quantity * v.Price
	v.SubTotalBeforeDiscount = v.Quantity * v.OriginalPrice
	v.OriginalPrice = originalPrice
	v.AdjustmentPrice = adjustmentPrice

}

// func (s *CartService) countDiscount(discount *models.DiscountModel, price float64) (float64, float64) {
// 	if discount == nil {
// 		return 0, price
// 	}
// 	discountAmount := float64(0)
// 	discountedPrice := price
// 	switch discount.Type {
// 	case models.DiscountPercentage:
// 		discountAmount = price * (discount.Value / 100)
// 		discountedPrice -= price * (discount.Value / 100)
// 	case models.DiscountAmount:
// 		discountAmount = discount.Value
// 		discountedPrice -= discount.Value
// 	}

// 	// Pastikan harga tidak negatif
// 	if discountedPrice < 0 {
// 		discountedPrice = 0
// 	}

// 	return discountAmount, discountedPrice
// }

// AddItemToCart adds a new item to the active cart of the given user ID.
// If the item is already in the cart, it updates the quantity.
// The function returns an error if there is a database error.
// It also returns an error if the product is not active.
func (s *CartService) AddItemToCart(userID string, productID string, variantID *string, quantity float64) error {
	// Dapatkan cart active
	cart, err := s.GetOrCreateActiveCart(userID)
	if err != nil {
		return err
	}
	price := float64(0)

	var product models.ProductModel
	product.MerchantID = s.merchantID
	if err := s.ctx.DB.Model(&models.ProductModel{}).Where("id = ?", productID).First(&product).Error; err != nil {
		return err
	}
	if product.Status != "ACTIVE" {
		return errors.New("product not active")
	}
	product.GetPriceAndDiscount(s.db)
	var width, height, weight, length float64 = product.Width, product.Height, product.Weight, product.Length
	originalPrice := product.OriginalPrice
	price = product.Price
	fmt.Println("PRICE #1", price)
	var discountAmount float64 = product.DiscountAmount
	var discountType string = product.DiscountType
	var discountRate float64 = product.DiscountRate
	var adjustmentPrice float64 = product.AdjustmentPrice

	// Cek apakah item sudah ada di cart
	var existingItem models.CartItemModel
	if variantID == nil {
		err = s.db.Where("cart_id = ? AND product_id = ?", cart.ID, productID).First(&existingItem).Error
	} else {
		err = s.db.Where("cart_id = ? AND product_id = ? AND variant_id = ?", cart.ID, productID, variantID).First(&existingItem).Error
		variant := models.VariantModel{}
		variant.MerchantID = s.merchantID
		s.ctx.DB.Where("id = ?", *variantID).First(&variant)
		variant.GetPriceAndDiscount(s.db)

		if variant.Width != 0 {
			width = variant.Width
		}
		if variant.Height != 0 {
			height = variant.Height
		}
		if variant.Weight != 0 {
			weight = variant.Weight
		}
		if variant.Length != 0 {
			length = variant.Length
		}
		price = variant.Price
		discountAmount = variant.DiscountAmount
		discountType = variant.DiscountType
		discountRate = variant.DiscountRate
		originalPrice = variant.OriginalPrice
		adjustmentPrice = variant.AdjustmentPrice
		fmt.Println("PRICE #2", price, originalPrice, adjustmentPrice)
	}
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Tambahkan item baru ke cart

			fmt.Println("PRICE #3", price, originalPrice, adjustmentPrice)
			item := models.CartItemModel{
				CartID:          cart.ID,
				ProductID:       productID,
				VariantID:       variantID,
				Quantity:        quantity,
				Price:           price,
				DiscountAmount:  discountAmount,
				DiscountType:    discountType,
				DiscountRate:    discountRate,
				OriginalPrice:   originalPrice,
				AdjustmentPrice: adjustmentPrice,
				Width:           width,
				Height:          height,
				Weight:          weight,
				Length:          length,
				CategoryID:      product.CategoryID,
				BrandID:         product.BrandID,
			}
			if err := s.db.Create(&item).Error; err != nil {
				return err
			}
		} else {
			return err
		}
	} else {
		fmt.Println("PRICE #4", price)
		// Update quantity jika item sudah ada
		existingItem.Quantity += quantity
		existingItem.Price = price
		if err := s.db.Save(&existingItem).Error; err != nil {
			return err
		}
	}

	return nil
}

// DeleteItemCart deletes an item from the active cart of the given user ID.
//
// The function will return an error if there is a database error.
func (s *CartService) DeleteItemCart(userID string, itemID string) error {
	// Dapatkan cart active
	cart, err := s.GetOrCreateActiveCart(userID)
	if err != nil {
		return err
	}

	// Hapus item dari cart
	if err := s.db.Where("cart_id = ? AND id = ?", cart.ID, itemID).Unscoped().Delete(&models.CartItemModel{}).Error; err != nil {
		return err
	}

	return nil
}

// UpdateItemCart updates the quantity of an item in the active cart of the given user ID.
//
// It takes the user ID, item ID, and quantity as arguments.
//
// It returns an error if there is a database error or if the item is not found in the active cart.
func (s *CartService) UpdateItemCart(userID string, itemID string, quantity float64) error {
	// Dapatkan cart active
	cart, err := s.GetOrCreateActiveCart(userID)
	if err != nil {
		return err
	}

	// Update quantity item di cart
	var existingItem models.CartItemModel
	if err := s.db.Where("cart_id = ? AND id = ?", cart.ID, itemID).First(&existingItem).Error; err != nil {
		return err
	}
	existingItem.Quantity = quantity
	if err := s.db.Save(&existingItem).Error; err != nil {
		return err
	}

	return nil
}

// FinishCart changes the status of the active cart for the given user to "FINISHED".
//
// This function retrieves the active cart for the specified user and updates its
// status to indicate that the cart is complete. It returns an error if there is a
// database error or if the active cart cannot be retrieved.

func (s *CartService) FinishCart(userID string) error {
	// Dapatkan cart active
	cart, err := s.GetOrCreateActiveCart(userID)
	if err != nil {
		return err
	}

	// Ubah status cart menjadi finished
	cart.Status = "FINISHED"
	if err := s.db.Save(&cart).Error; err != nil {
		return err
	}

	return nil
}

// CountSubTotalByCartID returns the total price of all items in the cart with the given cart ID.
//
// The function takes a cart ID as an argument and returns the total price of all items in
// the cart with the given ID. It returns an error if there is a database error.
func (s *CartService) CountSubTotalByCartID(cartID string) (float64, error) {
	var subTotal float64
	err := s.db.Model(&models.CartItemModel{}).
		Where("cart_id = ?", cartID).
		Select("SUM(quantity * price) AS sub_total").
		Scan(&subTotal).Error
	if err != nil {
		return 0, err
	}
	return subTotal, nil
}

// CountSubTotal calculates the total price of all items in the active cart for the given user.
//
// It takes a user ID as an argument and sums up the product of quantity and price of
// all items in the user's active cart. The function returns the subtotal price and an error
// if there is a database error during the retrieval process.

func (s *CartService) CountSubTotal(userID string) (float64, error) {
	var subTotal float64
	err := s.db.Model(&models.CartItemModel{}).
		Where("cart_id IN (SELECT id FROM carts WHERE user_id = ? AND status = ?)", userID, "ACTIVE").
		Select("COALESCE(SUM(quantity * price), 0) AS sub_total").
		Scan(&subTotal).Error
	if err != nil {
		return 0, err
	}
	return subTotal, nil
}
