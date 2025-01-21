package cart

import (
	"errors"

	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/inventory"
	"github.com/AMETORY/ametory-erp-modules/shared/models"
	"gorm.io/gorm"
)

type CartService struct {
	db               *gorm.DB
	ctx              *context.ERPContext
	inventoryService *inventory.InventoryService
}

func NewCartService(db *gorm.DB, ctx *context.ERPContext, inventoryService *inventory.InventoryService) *CartService {
	return &CartService{db: db, ctx: ctx, inventoryService: inventoryService}
}

func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(&models.CartModel{}, &models.CartItemModel{})
}

func (s *CartService) GetCartByID(cartID string) (*models.CartModel, error) {
	var cart models.CartModel
	err := s.db.Preload("Items.Product").Where("id = ?", cartID).First(&cart).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("cart not found")
		}
		return nil, err
	}
	subTotal, _ := s.CountSubTotalByCartID(cartID)
	cart.SubTotal = subTotal
	for i, v := range cart.Items {
		img, _ := s.inventoryService.ProductService.ListImagesOfProduct(v.ProductID)
		v.Product.ProductImages = img
		cart.Items[i] = v
	}
	return &cart, nil
}

func (s *CartService) GetOrCreateActiveCart(userID string) (*models.CartModel, error) {
	var cart models.CartModel
	err := s.db.Preload("Items.Product").Where("user_id = ? AND status = ?", userID, "ACTIVE").First(&cart).Error
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
	for i, v := range cart.Items {
		img, _ := s.inventoryService.ProductService.ListImagesOfProduct(v.ProductID)
		v.Product.ProductImages = img
		cart.Items[i] = v
	}
	return &cart, nil
}

func (s *CartService) AddItemToCart(userID string, productID string, quantity float64) error {
	// Dapatkan cart active
	cart, err := s.GetOrCreateActiveCart(userID)
	if err != nil {
		return err
	}
	var product models.ProductModel
	if err := s.ctx.DB.Select("price").Model(&models.ProductModel{}).Where("id = ?", productID).First(&product).Error; err != nil {
		return err
	}

	// Cek apakah item sudah ada di cart
	var existingItem models.CartItemModel
	err = s.db.Where("cart_id = ? AND product_id = ?", cart.ID, productID).First(&existingItem).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Tambahkan item baru ke cart
			item := models.CartItemModel{
				CartID:    cart.ID,
				ProductID: productID,
				Quantity:  quantity,
				Price:     product.Price,
			}
			if err := s.db.Create(&item).Error; err != nil {
				return err
			}
		} else {
			return err
		}
	} else {
		// Update quantity jika item sudah ada
		existingItem.Quantity += quantity
		existingItem.Price = product.Price
		if err := s.db.Save(&existingItem).Error; err != nil {
			return err
		}
	}

	return nil
}

func (s *CartService) DeleteItemCart(userID string, productID string) error {
	// Dapatkan cart active
	cart, err := s.GetOrCreateActiveCart(userID)
	if err != nil {
		return err
	}

	// Hapus item dari cart
	if err := s.db.Where("cart_id = ? AND product_id = ?", cart.ID, productID).Unscoped().Delete(&models.CartItemModel{}).Error; err != nil {
		return err
	}

	return nil
}

func (s *CartService) UpdateItemCart(userID string, productID string, quantity float64) error {
	// Dapatkan cart active
	cart, err := s.GetOrCreateActiveCart(userID)
	if err != nil {
		return err
	}

	// Update quantity item di cart
	var existingItem models.CartItemModel
	if err := s.db.Where("cart_id = ? AND product_id = ?", cart.ID, productID).First(&existingItem).Error; err != nil {
		return err
	}
	existingItem.Quantity = quantity
	if err := s.db.Save(&existingItem).Error; err != nil {
		return err
	}

	return nil
}
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
func (s *CartService) CountSubTotal(userID string) (float64, error) {
	var subTotal float64
	err := s.db.Model(&models.CartItemModel{}).
		Where("cart_id IN (SELECT id FROM carts WHERE user_id = ? AND status = ?)", userID, "ACTIVE").
		Select("SUM(quantity * price) AS sub_total").
		Scan(&subTotal).Error
	if err != nil {
		return 0, err
	}
	return subTotal, nil
}
