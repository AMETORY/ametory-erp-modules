package company

import (
	"fmt"

	"github.com/AMETORY/ametory-erp-modules/utils"
	"gorm.io/gorm"
)

type CompanyModel struct {
	utils.BaseModel
	Name                  string `json:"name"`
	Logo                  string `json:"logo"`
	Cover                 string `json:"cover"`
	LegalEntity           string `json:"legal_entity"`
	Email                 string `json:"email"`
	Phone                 string `json:"phone"`
	Fax                   string `json:"fax"`
	Address               string `json:"address"`
	ContactPerson         string `json:"contact_person"`
	ContactPersonPosition string `json:"contact_person_position"`
	TaxPayerNumber        string `json:"tax_payer_number"`
	UserID                string `json:"user_id"`
	Status                string `json:"status" gorm:"type:VARCHAR(20) DEFAULT 'ACTIVE'"`
	EmployeeActiveCount   int64  `json:"employee_active_count"`
	EmployeeResignCount   int64  `json:"employee_resign_count"`
	EmployeePendingCount  int64  `json:"employee_pending_count"`
}

func (CompanyModel) TableName() string {
	return "companies"
}

type CompanyService struct {
	db            *gorm.DB
	SkipMigration bool
}

func NewCompanyService(db *gorm.DB, skipMigrate bool) *CompanyService {
	fmt.Println("INIT COMPANY SERVICE")
	var service = CompanyService{db: db, SkipMigration: skipMigrate}
	err := service.Migrate()
	if err != nil {
		fmt.Println(err)
		return nil
	}
	return &service
}

func (s *CompanyService) Migrate() error {
	if s.SkipMigration {
		return nil
	}
	return s.db.AutoMigrate(&CompanyModel{})
}

func (s *CompanyService) DB() *gorm.DB {
	return s.db
}
