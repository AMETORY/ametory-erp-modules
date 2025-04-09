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

func NewAccountService(db *gorm.DB, ctx *context.ERPContext) *AccountService {
	return &AccountService{db: db, ctx: ctx}
}

func Migrate(db *gorm.DB) error {
	fmt.Println("Migrating account model...")
	return db.AutoMigrate(&models.AccountModel{})
}
func (s *AccountService) CreateAccount(data *models.AccountModel) error {
	return s.db.Create(data).Error
}

func (s *AccountService) UpdateAccount(id string, data *models.AccountModel) error {
	return s.db.Where("id = ?", id).Updates(data).Error
}

func (s *AccountService) DeleteAccount(id string) error {
	return s.db.Where("id = ?", id).Delete(&models.AccountModel{}).Error
}

func (s *AccountService) GetAccountByID(id string) (*models.AccountModel, error) {
	var account models.AccountModel
	err := s.db.Where("id = ?", id).First(&account).Error
	return &account, err
}

func (s *AccountService) GetAccountByCode(code string) (*models.AccountModel, error) {
	var account models.AccountModel
	err := s.db.Where("code = ?", code).First(&account).Error
	return &account, err
}

func (s *AccountService) GetAccounts(request http.Request, search string) (paginate.Page, error) {
	// GET COGS ACCOUNT
	if request.Header.Get("ID-Company") != "" {
		companyID := request.Header.Get("ID-Company")
		var cogsAccount models.AccountModel
		err := s.db.Where("is_cogs_account = ? and company_id = ?", true, request.Header.Get("ID-Company")).First(&cogsAccount).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			err := s.db.Create(&models.AccountModel{
				Code:          "61003",
				CompanyID:     &companyID,
				Name:          "Beban Pokok Penjualan",
				Type:          models.COST,
				Category:      constants.CATEGORY_COST_OF_REVENUE,
				IsCogsAccount: true,
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

func (s *AccountService) GetTypes() map[string]any {

	return map[string]any{
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
