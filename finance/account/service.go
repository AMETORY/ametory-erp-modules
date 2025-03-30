package account

import (
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
					},
				},
			},
		},
		string(models.INCOME): map[string]any{
			"name": string(models.INCOME),
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
