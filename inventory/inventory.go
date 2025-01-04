package inventory

import (
	"fmt"

	"github.com/AMETORY/ametory-erp-modules/inventory/product"
	"gorm.io/gorm"
)

type InventoryService struct {
	db                     *gorm.DB
	ProductService         *product.ProductService
	ProductCategoryService *product.ProductCategoryService
	// TransactionService *transaction.TransactionService
	SkipMigration bool
}

func NewInventoryService(db *gorm.DB, skipMigrate bool) *InventoryService {
	fmt.Println("INIT INVENTORY SERVICE")
	var service = InventoryService{
		db:                     db,
		SkipMigration:          skipMigrate,
		ProductService:         product.NewProductService(db),
		ProductCategoryService: product.NewProductCategoryService(db),
	}
	err := service.Migrate()
	if err != nil {
		fmt.Println(err)
		return nil
	}
	return &service
}

func (s *InventoryService) Migrate() error {
	if s.SkipMigration {
		return nil
	}
	if err := product.Migrate(s.db); err != nil {
		fmt.Println("ERROR ACCOUNT", err)
		return err
	}
	// if err := transaction.Migrate(s.db); err != nil {
	// 	return err
	// }

	return nil
}
func (s *InventoryService) DB() *gorm.DB {
	return s.db
}
