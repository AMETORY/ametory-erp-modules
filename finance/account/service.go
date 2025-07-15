package account

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/shared/constants"
	"github.com/AMETORY/ametory-erp-modules/shared/models"
	"github.com/AMETORY/ametory-erp-modules/utils"
	"github.com/morkid/paginate"
	"gorm.io/gorm"
)

type AccountService struct {
	db  *gorm.DB
	ctx *context.ERPContext
}

var defaultAccountGroups = []map[string]any{
	{
		"label": constants.OPERATING_VALUE,
		"value": constants.OPERATING,
		"subgroups": []map[string]any{
			{
				"label": constants.ACCEPTANCE_FROM_CUSTOMERS_VALUE,
				"value": constants.ACCEPTANCE_FROM_CUSTOMERS,
			},
			{
				"label": constants.OTHER_CURRENT_ASSETS_VALUE,
				"value": constants.OTHER_CURRENT_ASSETS,
			},
			{
				"label": constants.PAYMENT_TO_VENDORS_VALUE,
				"value": constants.PAYMENT_TO_VENDORS,
			},
			{
				"label": constants.CREDIT_CARDS_AND_OTHER_SHORT_TERM_LIABILITIES_VALUE,
				"value": constants.CREDIT_CARDS_AND_OTHER_SHORT_TERM_LIABILITIES,
			},
			{
				"label": constants.OTHER_INCOME_VALUE,
				"value": constants.OTHER_INCOME,
			},
			{
				"label": constants.OPERATIONAL_EXPENSES_VALUE,
				"value": constants.OPERATIONAL_EXPENSES,
			},
			{
				"label": constants.RETURNS_PAYMENT_OF_TAXES_VALUE,
				"value": constants.RETURNS_PAYMENT_OF_TAXES,
			},
			{
				"label": constants.COOPERATIVE_ACCEPTANCE_FROM_MEMBER_LABEL,
				"value": constants.COOPERATIVE_ACCEPTANCE_FROM_MEMBER,
			},
			{
				"label": constants.COOPERATIVE_ACCEPTANCE_FROM_NON_MEMBER_LABEL,
				"value": constants.COOPERATIVE_ACCEPTANCE_FROM_NON_MEMBER,
			},
		},
	},
	{
		"label": constants.INVESTING_VALUE,
		"value": constants.INVESTING,
		"subgroups": []map[string]any{
			{
				"label": constants.ACQUISITION_SALE_OF_ASSETS_VALUE,
				"value": constants.ACQUISITION_SALE_OF_ASSETS,
			},
			{
				"label": constants.OTHER_INVESTMENT_ACTIVITIES_VALUE,
				"value": constants.OTHER_INVESTMENT_ACTIVITIES,
			},
			{
				"label": constants.INVESTMENT_PARTNERSHIP_VALUE,
				"value": constants.INVESTMENT_PARTNERSHIP,
			},
		},
	},
	{
		"label": constants.FINANCING_VALUE,
		"value": constants.FINANCING,
		"subgroups": []map[string]any{
			{
				"label": constants.LOAN_PAYMENTS_RECEIPTS_VALUE,
				"value": constants.LOAN_PAYMENTS_RECEIPTS,
			},
			{
				"label": constants.EQUITY_CAPITAL_VALUE,
				"value": constants.EQUITY_CAPITAL,
			},
			{
				"label": constants.COOPERATIVE_PRINCIPAL_SAVING_LABEL,
				"value": constants.COOPERATIVE_PRINCIPAL_SAVING,
			},
			{
				"label": constants.COOPERATIVE_MANDATORY_SAVING_LABEL,
				"value": constants.COOPERATIVE_MANDATORY_SAVING,
			},
			{
				"label": constants.COOPERATIVE_VOLUNTARY_SAVING_LABEL,
				"value": constants.COOPERATIVE_VOLUNTARY_SAVING,
			},
		},
	},
}

// NewAccountService returns a new instance of AccountService.
//
// The service is created by providing a GORM database instance and an ERP context.
// The ERP context is used for authentication and authorization purposes, while the
// database instance is used for CRUD (Create, Read, Update, Delete) operations.
func NewAccountService(db *gorm.DB, ctx *context.ERPContext) *AccountService {
	return &AccountService{db: db, ctx: ctx}
}

// Migrate runs database migrations for the account module.
//
// This function is used to create the database schema for the account model.
// It is called once during the application startup process.
//
// The function takes a GORM database instance as its argument and returns an
// error if the migration fails.
func Migrate(db *gorm.DB) error {
	fmt.Println("Migrating account model...")
	return db.AutoMigrate(&models.AccountModel{})
}

// CreateAccount creates a new account in the database.
//
// The function takes a pointer to AccountModel as its argument and returns an
// error if the creation fails.
//
// The function is idempotent, meaning that if the account already exists, no error
// is returned.
func (s *AccountService) CreateAccount(data *models.AccountModel) error {
	return s.db.Create(data).Error
}

// UpdateAccount updates an existing account in the database.
//
// The function takes an ID of the account to be updated and a pointer to
// AccountModel as its arguments. The AccountModel instance contains the new values
// to be updated in the database.
//
// The function returns an error if the update fails.
func (s *AccountService) UpdateAccount(id string, data *models.AccountModel) error {
	return s.db.Where("id = ?", id).Updates(data).Error
}

// DeleteAccount deletes an existing account from the database.
//
// The function takes an ID of the account to be deleted as its argument.
// It returns an error if the deletion operation fails.
//
// The function uses GORM to delete the account data from the accounts table.
// If the deletion is successful, the error is nil. Otherwise, the error contains
// information about what went wrong.

func (s *AccountService) DeleteAccount(id string) error {
	return s.db.Where("id = ?", id).Delete(&models.AccountModel{}).Error
}

// GetAccountByID retrieves an account record from the database by ID.
//
// It takes an ID as input and returns a pointer to a AccountModel and an error.
// The function uses GORM to retrieve the account data from the accounts table.
// If the operation fails, an error is returned.
func (s *AccountService) GetAccountByID(id string) (*models.AccountModel, error) {
	var account models.AccountModel
	err := s.db.Where("id = ?", id).First(&account).Error
	return &account, err
}

// GetAccountByCode retrieves an account record from the database by code.
//
// It takes a code as input and returns a pointer to an AccountModel and an error.
// The function uses GORM to retrieve the account data from the accounts table.
// If the operation fails, an error is returned. Otherwise, the error is nil.

func (s *AccountService) GetAccountByCode(code string) (*models.AccountModel, error) {
	var account models.AccountModel
	err := s.db.Where("code = ?", code).First(&account).Error
	return &account, err
}

// GetMasterAccounts retrieves a paginated list of master accounts from the database.
//
// It takes an HTTP request as its argument. The function uses pagination to manage
// the result set and includes any necessary request modifications using the
// utils.FixRequest utility.
//
// The function returns a paginated page of AccountModel and an error if the
// operation fails. Otherwise, the error is nil.
func (s *AccountService) GetMasterAccounts(request *http.Request) (paginate.Page, error) {
	pg := paginate.New()
	stmt := s.db.Model(&models.AccountModel{})
	utils.FixRequest(request)
	page := pg.With(stmt).Request(request).Response(&[]models.AccountModel{})
	page.Page = page.Page + 1
	return page, nil
}

// GetAccounts retrieves a paginated list of accounts based on various filters.
//
// It takes an HTTP request and a search query string as input. The method uses
// GORM to query the database for accounts, applying the search query to the
// account name and code fields. If the request contains a company ID header,
// the method filters the result by the company ID and ensures necessary accounts
// are created if they don't exist. Additional filters can be applied based on
// account type, cashflow subgroups, cashflow groups, categories, and specific
// account flags such as profit/loss, COGS, inventory, and tax accounts.
// The function utilizes pagination to manage the result set and includes any
// necessary request modifications using the utils.FixRequest utility.
//
// The function returns a paginated page of AccountModel and an error if the
// operation fails. Otherwise, the error is nil.

func (s *AccountService) GetAccounts(request http.Request, search string) (paginate.Page, error) {
	// GET COGS ACCOUNT
	if request.Header.Get("ID-Company") != "" {
		companyID := request.Header.Get("ID-Company")
		var profitLossSumAccount models.AccountModel
		err := s.db.Where("is_profit_loss_account = ? and company_id = ?", true, request.Header.Get("ID-Company")).First(&profitLossSumAccount).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			err := s.db.Create(&models.AccountModel{
				Code:                "32001",
				CompanyID:           &companyID,
				Name:                "Ikhtisar Laba Rugi",
				Type:                models.EQUITY,
				Category:            constants.ACCOUNT_PROFIT_LOSS,
				IsProfitLossAccount: true,
			}).Error
			if err != nil {
				return paginate.Page{}, err
			}
		}
		var profitLossClosingAccount models.AccountModel
		err = s.db.Where("is_profit_loss_closing_account = ? and company_id = ?", true, request.Header.Get("ID-Company")).First(&profitLossClosingAccount).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			err = s.db.Create(&models.AccountModel{
				Code:                       "32002",
				CompanyID:                  &companyID,
				Name:                       "Laba Ditahan / SHU Tahun Berjalan",
				Type:                       models.EQUITY,
				Category:                   constants.ACCOUNT_PROFIT_LOSS,
				IsProfitLossClosingAccount: true,
			}).Error
			if err != nil {
				return paginate.Page{}, err
			}
		}
		var cogsAccount models.AccountModel
		err = s.db.Where("is_cogs_account = ? and company_id = ?", true, request.Header.Get("ID-Company")).First(&cogsAccount).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			err := s.db.Create(&models.AccountModel{
				Code:          "61003",
				CompanyID:     &companyID,
				Name:          "Biaya Pokok Penjualan",
				Type:          models.COST,
				Category:      constants.CATEGORY_COST_OF_REVENUE,
				IsCogsAccount: true,
			}).Error
			if err != nil {
				return paginate.Page{}, err
			}
		}
		var cogsClosingAccount models.AccountModel
		err = s.db.Where("is_cogs_closing_account = ? and company_id = ? AND name = ?", true, request.Header.Get("ID-Company"), "HARGA POKOK PENJUALAN").First(&cogsClosingAccount).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			err := s.db.Create(&models.AccountModel{
				CompanyID:            &companyID,
				Name:                 "HARGA POKOK PENJUALAN",
				Type:                 models.EXPENSE,
				Category:             constants.COGS,
				IsCogsClosingAccount: true,
			}).Error
			if err != nil {
				return paginate.Page{}, err
			}
		}
		// GET INVENTORY ACCOUNT
		var inventoryAccount models.AccountModel
		err = s.db.Where("is_inventory_account = ? and company_id = ?", true, request.Header.Get("ID-Company")).First(&inventoryAccount).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			err := s.db.Create(&models.AccountModel{
				Code:               "13001",
				CompanyID:          &companyID,
				Name:               "Persediaan",
				Type:               models.ASSET,
				Category:           constants.CATEGORY_CURRENT_ASSET,
				CashflowGroup:      constants.CASHFLOW_GROUP_OPERATING,
				CashflowSubGroup:   constants.PAYMENT_TO_VENDORS,
				IsInventoryAccount: true,
			}).Error
			if err != nil {
				return paginate.Page{}, err
			}
		}

		var contraRevenueAccount models.AccountModel
		err = s.db.Where("type = ? and company_id = ?", models.CONTRA_REVENUE, request.Header.Get("ID-Company")).First(&contraRevenueAccount).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			err := s.db.Create(&models.AccountModel{
				Code:             "42001",
				CompanyID:        &companyID,
				Name:             "Diskon Penjualan",
				Type:             models.CONTRA_REVENUE,
				CashflowSubGroup: constants.ACCEPTANCE_FROM_CUSTOMERS,
				CashflowGroup:    constants.CASHFLOW_GROUP_OPERATING,
				Category:         constants.CATEGORY_REVENUE,
				IsDiscount:       true,
			}).Error
			if err != nil {
				return paginate.Page{}, err
			}
			err = s.db.Create(&models.AccountModel{
				Code:             "43001",
				CompanyID:        &companyID,
				Name:             "Retur Penjualan",
				Type:             models.CONTRA_REVENUE,
				CashflowSubGroup: constants.ACCEPTANCE_FROM_CUSTOMERS,
				CashflowGroup:    constants.CASHFLOW_GROUP_OPERATING,
				Category:         constants.CATEGORY_REVENUE,
				IsReturn:         true,
			}).Error
			if err != nil {
				return paginate.Page{}, err
			}
		}

		var stockOpnameAccount models.AccountModel
		err = s.db.Where("is_stock_opname_account = ? and company_id = ?", true, request.Header.Get("ID-Company")).First(&stockOpnameAccount).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			err := s.db.Create(&models.AccountModel{
				Code:                 "44001",
				CompanyID:            &companyID,
				Name:                 "Koreksi Persediaan Masuk",
				Type:                 models.REVENUE,
				Category:             constants.CATEGORY_REVENUE,
				IsStockOpnameAccount: true,
			}).Error
			if err != nil {
				return paginate.Page{}, err
			}
			err = s.db.Create(&models.AccountModel{
				Code:                 "54001",
				CompanyID:            &companyID,
				Name:                 "Penyesuaian Persediaan",
				Type:                 models.EXPENSE,
				Category:             constants.CATEGORY_EXPENSE,
				IsStockOpnameAccount: true,
			}).Error
			if err != nil {
				return paginate.Page{}, err
			}
		}
	}

	pg := paginate.New()
	stmt := s.db
	if request.Header.Get("ID-Company") != "" {
		stmt = stmt.Where("company_id = ?", request.Header.Get("ID-Company"))
	}
	if search != "" {
		stmt = stmt.Where("accounts.name ILIKE ? OR accounts.code ILIKE ? ",
			"%"+search+"%",
			"%"+search+"%",
		)
	}
	if request.URL.Query().Get("type") != "" {
		types := strings.Split(request.URL.Query().Get("type"), ",")
		stmt = stmt.Where("accounts.type IN (?) ", types)
	}
	if request.URL.Query().Get("cashflow_sub_group") != "" {
		subgroups := strings.Split(request.URL.Query().Get("cashflow_sub_group"), ",")
		stmt = stmt.Where("accounts.cashflow_sub_group IN (?) ", subgroups)
	}
	if request.URL.Query().Get("cashflow_group") != "" {
		groups := strings.Split(request.URL.Query().Get("cashflow_group"), ",")
		stmt = stmt.Where("accounts.cashflow_group IN (?) ", groups)
	}
	if request.URL.Query().Get("category") != "" {
		stmt = stmt.Where("accounts.category = ? ", request.URL.Query().Get("category"))
	}
	if request.URL.Query().Get("is_profit_loss_account") != "" {
		stmt = stmt.Where("accounts.is_profit_loss_account = ? ", true)
	}
	if request.URL.Query().Get("is_profit_loss_closing_account") != "" {
		stmt = stmt.Where("accounts.is_profit_loss_closing_account = ? ", true)
	}
	if request.URL.Query().Get("is_cogs_closing_account") != "" {
		stmt = stmt.Where("accounts.is_cogs_closing_account = ? ", true)
	}
	if request.URL.Query().Get("is_cogs_account") != "" {
		stmt = stmt.Where("accounts.is_cogs_account = ? ", true)
	}
	if request.URL.Query().Get("is_inventory_account") != "" {
		stmt = stmt.Where("accounts.is_inventory_account = ? ", true)
	}
	if request.URL.Query().Get("is_tax") != "" {
		isTax := request.URL.Query().Get("is_tax") == "true" || request.URL.Query().Get("is_tax") == "1"
		stmt = stmt.Where("accounts.is_tax = ? ", isTax)
	}

	stmt = stmt.Model(&models.AccountModel{})
	utils.FixRequest(&request)
	page := pg.With(stmt).Request(request).Response(&[]models.AccountModel{})
	page.Page = page.Page + 1
	return page, nil
}

// GetTypes returns a map of all account types, their categories, and cashflow groups.
func (s *AccountService) GetTypes() map[string]any {

	return map[string]any{
		string(models.CONTRA_ASSET): map[string]any{
			"name": string(models.CONTRA_ASSET),
			"categories": []string{
				constants.CATEGORY_FIXED_ASSET,
			},
			"groups": []map[string]any{
				{
					"label": constants.CASHFLOW_GROUP_FIXED_ASSET_VALUE,
					"value": constants.CASHFLOW_GROUP_FIXED_ASSET,
					"subgroups": []map[string]any{

						{
							"label": constants.CASHFLOW_GROUP_DEPRECIATION_AMORTIZATION_VALUE,
							"value": constants.CASHFLOW_GROUP_DEPRECIATION_AMORTIZATION,
						},
					},
				},
			},
		},
		string(models.ASSET): map[string]any{
			"name": string(models.ASSET),
			"categories": []string{
				constants.CATEGORY_FIXED_ASSET,
				constants.CATEGORY_CURRENT_ASSET,
			},
			"groups": []map[string]any{
				{
					"label": constants.CASHFLOW_GROUP_FIXED_ASSET_VALUE,
					"value": constants.CASHFLOW_GROUP_FIXED_ASSET,
					"subgroups": []map[string]any{
						{
							"label": constants.CASHFLOW_GROUP_FIXED_ASSET_VALUE,
							"value": constants.CASHFLOW_GROUP_FIXED_ASSET,
						},
						{
							"label": constants.CASHFLOW_GROUP_DEPRECIATION_AMORTIZATION_VALUE,
							"value": constants.CASHFLOW_GROUP_DEPRECIATION_AMORTIZATION,
						},
					},
				},
				{
					"label": constants.CASHFLOW_GROUP_CURRENT_ASSET_VALUE,
					"value": constants.CASHFLOW_GROUP_CURRENT_ASSET,
					"subgroups": []map[string]any{
						{
							"label": constants.CASHFLOW_GROUP_CURRENT_ASSET_VALUE,
							"value": constants.CASHFLOW_GROUP_CURRENT_ASSET,
						},
					},
				},
			},
		},
		string(models.RECEIVABLE): map[string]any{
			"name": string(models.RECEIVABLE),
			"categories": []string{
				constants.CATEGORY_RECEIVABLE,
			},
			"groups": defaultAccountGroups,
		},
		string(models.LIABILITY): map[string]any{
			"name": string(models.LIABILITY),
			"categories": []string{
				constants.CATEGORY_DEBT,
			},
			"groups": defaultAccountGroups,
		},
		string(models.EQUITY): map[string]any{
			"name": string(models.EQUITY),
			"categories": []string{
				constants.CATEGORY_EQUITY,
			},
			"groups": []map[string]any{
				{
					"label": constants.FINANCING_VALUE,
					"value": constants.FINANCING,
					"subgroups": []map[string]any{
						{
							"label": constants.LOAN_PAYMENTS_RECEIPTS_VALUE,
							"value": constants.LOAN_PAYMENTS_RECEIPTS,
						},
						{
							"label": constants.EQUITY_CAPITAL_VALUE,
							"value": constants.EQUITY_CAPITAL,
						},
						{
							"label": constants.COOPERATIVE_PRINCIPAL_SAVING_LABEL,
							"value": constants.COOPERATIVE_PRINCIPAL_SAVING,
						},
						{
							"label": constants.COOPERATIVE_MANDATORY_SAVING_LABEL,
							"value": constants.COOPERATIVE_MANDATORY_SAVING,
						},
						{
							"label": constants.COOPERATIVE_VOLUNTARY_SAVING_LABEL,
							"value": constants.COOPERATIVE_VOLUNTARY_SAVING,
						},
					},
				},
			},
		},
		string(models.REVENUE): map[string]any{
			"name": string(models.REVENUE),
			"categories": []string{
				constants.CATEGORY_SALES,
				constants.CATEGORY_OTHER_INCOME,
			},
			"groups": defaultAccountGroups,
		},
		string(models.EXPENSE): map[string]any{
			"name": string(models.EXPENSE),
			"categories": []string{
				constants.CATEGORY_EXPENSE,
				constants.CATEGORY_OPERATING,
			},
			"groups": defaultAccountGroups,
		},
		string(models.COST): map[string]any{
			"name": string(models.COST),
			"categories": []string{
				constants.CATEGORY_EXPENSE,
				constants.CATEGORY_OPERATING,
			},
			"groups": defaultAccountGroups,
		},
	}
}
