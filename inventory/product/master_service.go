package product

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/shared/models"
	"github.com/AMETORY/ametory-erp-modules/utils"
	"github.com/morkid/paginate"
	"gorm.io/gorm"
)

type MasterProductService struct {
	db  *gorm.DB
	ctx *context.ERPContext
}

// NewMasterProductService creates a new instance of MasterProductService.
//
// Args:
//
//	db: the gorm database instance.
//	ctx: the erp context.
//
// Returns:
//
//	A new instance of MasterProductService.
func NewMasterProductService(db *gorm.DB, ctx *context.ERPContext) *MasterProductService {
	return &MasterProductService{db: db, ctx: ctx}
}

func (s *MasterProductService) CreateMasterProduct(data *models.MasterProductModel) error {
	return s.db.Create(data).Error
}

// UpdateMasterProduct updates a master product.
//
// Args:
//
//	id: the id of the master product to update.
//	data: the updated data of the master product.
//
// Returns:
//
//	an error if any error occurs.
func (s *MasterProductService) UpdateMasterProduct(id string, data *models.MasterProductModel) error {
	return s.db.Where("id = ?", id).Updates(data).Error
}

// DeleteMasterProduct deletes a master product in the database.
//
// Args:
//
//	id: the id of the master product to delete.
//
// Returns:
//
//	an error if any error occurs.
func (s *MasterProductService) DeleteMasterProduct(id string) error {
	return s.db.Where("id = ?", id).Delete(&models.MasterProductModel{}).Error
}

// GetMasterProductByID retrieves a master product by its id.
//
// Args:
//
//	id: the id of the master product to retrieve.
//
// Returns:
//
//	the master product if found, and an error if any error occurs.
func (s *MasterProductService) GetMasterProductByID(id string) (*models.MasterProductModel, error) {
	var product models.MasterProductModel
	err := s.db.Preload("Category", func(db *gorm.DB) *gorm.DB {
		return db.Select("id", "name")
	}).Preload("Brand", func(db *gorm.DB) *gorm.DB {
		return db.Select("id", "name")
	}).Where("id = ?", id).First(&product).Error
	product.Prices, _ = s.ListPricesOfProduct(product.ID)
	product.ProductImages, _ = s.ListImagesOfProduct(product.ID)
	return &product, err
}

// GetMasterProductByCode retrieves a master product by its code.
//
// Args:
//
//	code: the code of the master product to retrieve.
//
// Returns:
//
//	the master product if found, and an error if any error occurs.
func (s *MasterProductService) GetMasterProductByCode(code string) (*models.MasterProductModel, error) {
	var product models.MasterProductModel
	err := s.db.Where("code = ?", code).First(&product).Error
	return &product, err
}

// GetMasterProducts retrieves a paginated list of master products from the database.
//
// It takes an http.Request and a search query string as input. The method uses
// GORM to query the database for master products, applying the search query to
// the master product name, description, SKU, and barcode fields. If the request
// contains a company ID header, the method also filters the result by the
// company ID. The function utilizes pagination to manage the result set and
// applies any necessary request modifications using the utils.FixRequest utility.
//
// The function returns a paginated page of MasterProductModel and an error if
// the operation fails.
func (s *MasterProductService) GetMasterProducts(request http.Request, search string) (paginate.Page, error) {
	pg := paginate.New()
	stmt := s.db.Preload("Category", func(db *gorm.DB) *gorm.DB {
		return db.Select("id", "name")
	}).Preload("Brand", func(db *gorm.DB) *gorm.DB {
		return db.Select("id", "name")
	})
	if search != "" {
		stmt = stmt.Where("master_products.description ILIKE ? OR master_products.sku ILIKE ? OR master_products.name ILIKE ? OR master_products.barcode ILIKE ?",
			"%"+search+"%",
			"%"+search+"%",
			"%"+search+"%",
			"%"+search+"%",
		)
	}
	if request.Header.Get("ID-Company") != "" {
		stmt = stmt.Where("company_id = ?", request.Header.Get("ID-Company"))
	}
	stmt = stmt.Model(&models.MasterProductModel{})
	utils.FixRequest(&request)
	page := pg.With(stmt).Request(request).Response(&[]models.MasterProductModel{})
	page.Page = page.Page + 1
	items := page.Items.(*[]models.MasterProductModel)
	newItems := make([]models.MasterProductModel, 0)
	for _, v := range *items {
		img, err := s.ListImagesOfProduct(v.ID)
		if err == nil {
			v.ProductImages = img
		}
		prices, err := s.ListPricesOfProduct(v.ID)
		if err == nil {
			v.Prices = prices
		}
		newItems = append(newItems, v)
	}
	page.Items = &newItems
	return page, nil
}

// CreatePriceCategory creates a new price category in the database.
//
// Args:
//
//	data: the price category data to create.
//
// Returns:
//
//	an error if any error occurs.
func (s *MasterProductService) CreatePriceCategory(data *models.PriceCategoryModel) error {
	return s.db.Create(data).Error
}

func (s *MasterProductService) AddPriceToMasterProduct(productID string, data *models.MasterProductPriceModel) error {
	if data.PriceCategoryID == "" {
		return errors.New("price category id is required")
	}
	data.MasterProductID = productID
	return s.db.Create(data).Error
}

// ListPricesOfProduct retrieves all the prices of a master product.
//
// Args:
//
//	productID: the id of the master product whose prices to retrieve.
//
// Returns:
//
//	all the prices of the master product if found, and an error if any error occurs.
func (s *MasterProductService) ListPricesOfProduct(productID string) ([]models.MasterProductPriceModel, error) {
	var prices []models.MasterProductPriceModel
	err := s.db.Preload("PriceCategory").Where("master_product_id = ?", productID).Find(&prices).Error
	return prices, err
}

// ListImagesOfProduct retrieves all the images of a master product.
//
// Args:
//
//	productID: the id of the master product whose images to retrieve.
//
// Returns:
//
//	all the images of the master product if found, and an error if any error occurs.
func (s *MasterProductService) ListImagesOfProduct(productID string) ([]models.FileModel, error) {
	var images []models.FileModel
	err := s.db.Where("ref_id = ? and ref_type = ?", productID, "master-product").Find(&images).Error
	return images, err
}

// DeletePriceFromMasterProduct deletes a specific price from a master product in the database.
//
// Args:
//
//	productID: the ID of the master product whose price to delete.
//	priceID: the ID of the price to delete.
//
// Returns:
//
//	an error if the deletion fails, otherwise returns nil.
func (s *MasterProductService) DeletePriceFromMasterProduct(productID string, priceID string) error {
	return s.db.Where("master_product_id = ? and id = ?", productID, priceID).Delete(&models.MasterProductPriceModel{}).Error
}

// DeleteImageFromMasterProduct deletes a specific image from a master product in the database.
//
// Args:
//
//	productID: the ID of the master product whose image to delete.
//	imageID: the ID of the image to delete.
//
// Returns:
//
//	an error if the deletion fails, otherwise returns nil.
func (s *MasterProductService) DeleteImageFromMasterProduct(productID string, imageID string) error {
	return s.db.Where("ref_id = ? and ref_type = ? and id = ?", productID, "master-product", imageID).Delete(&models.FileModel{}).Error
}

// ConvertToProducts creates a new product in the database from a master product.
//
// Args:
//
//	ids: a list of ids of the master products to convert.
//	distributorID: the id of the distributor to associate with the new product.
//
// Returns:
//
//	a list of errors if any error occurs, otherwise an empty list.
//
// If a product with the same master product id already exists, it will not be converted.
func (s *MasterProductService) ConvertToProducts(ids []string, distributorID *string) []string {
	newErrors := make([]string, 0)
	if len(ids) == 0 {
		return []string{"no ids provided"}
	}
	masterProducts := make([]models.MasterProductModel, 0)
	err := s.db.Where("id in (?)", ids).Find(&masterProducts).Error
	if err != nil {
		return []string{err.Error()}
	}

	for _, masterProduct := range masterProducts {
		existingProduct := models.ProductModel{}
		err = s.db.Where("master_product_id = ?", masterProduct.ID).First(&existingProduct).Error
		if err == nil {
			newErrors = append(newErrors, fmt.Sprintf("product with master id %s already exists", masterProduct.ID))
			continue
		}

		product := models.ProductModel{
			Name:            masterProduct.Name,
			Description:     masterProduct.Description,
			SKU:             masterProduct.SKU,
			Barcode:         masterProduct.Barcode,
			Price:           masterProduct.Price,
			CompanyID:       masterProduct.CompanyID,
			DistributorID:   distributorID,
			MasterProductID: &masterProduct.ID,
			CategoryID:      masterProduct.CategoryID,
			BrandID:         masterProduct.BrandID,
		}
		err = s.db.Create(&product).Error
		if err != nil {
			newErrors = append(newErrors, err.Error())
		}
		images, err := s.ListImagesOfProduct(masterProduct.ID)
		if err != nil {
			newErrors = append(newErrors, err.Error())
		}

		for _, v := range images {
			err := s.db.Create(&models.FileModel{
				FileName: v.FileName,
				MimeType: v.MimeType,
				Path:     v.Path,
				Provider: v.Provider,
				URL:      v.URL,
				RefID:    product.ID,
				RefType:  "product",
			}).Error
			if err != nil {
				newErrors = append(newErrors, err.Error())
			}
		}

		listPrices, err := s.ListPricesOfProduct(masterProduct.ID)
		if err != nil {
			newErrors = append(newErrors, err.Error())
		}
		for _, v := range listPrices {
			err := s.db.Create(&models.PriceModel{
				Amount:          v.Amount,
				Currency:        v.Currency,
				PriceCategoryID: v.PriceCategoryID,
				EffectiveDate:   v.EffectiveDate,
				MinQuantity:     v.MinQuantity,
				ProductID:       product.ID,
			}).Error
			if err != nil {
				newErrors = append(newErrors, err.Error())
			}
		}

	}

	return newErrors
}
