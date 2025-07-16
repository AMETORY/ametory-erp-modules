package payroll

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/hris/employee"
	"github.com/AMETORY/ametory-erp-modules/shared/models"
	"github.com/AMETORY/ametory-erp-modules/utils"
	"github.com/morkid/paginate"
	"gorm.io/gorm"
)

type PayrollService struct {
	db              *gorm.DB
	ctx             *context.ERPContext
	employeeService *employee.EmployeeService
}

// NewPayrollService creates a new PayrollService instance.
//
// The service is created by providing a GORM database instance and an ERP context.
// The ERP context is used for authentication and authorization purposes, while the
// database instance is used for CRUD (Create, Read, Update, Delete) operations.
// The service also takes an EmployeeService as argument, which is used to retrieve
// employee data.
func NewPayrollService(ctx *context.ERPContext, employeeService *employee.EmployeeService) *PayrollService {
	return &PayrollService{db: ctx.DB, ctx: ctx, employeeService: employeeService}
}

func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&models.PayRollModel{},
		&models.PayrollItemModel{},
		&models.PayRollCostModel{},
		&models.PayRollInstallment{},
		&models.PayRollPeriodeModel{},
	)
}

// CreatePayRoll adds a new payroll record to the database.
//
// This function takes a PayRollModel as input and attempts to create a new
// record in the database. It returns an error if the creation fails, otherwise
// it returns nil.
func (s *PayrollService) CreatePayRoll(payRoll *models.PayRollModel) error {
	return s.db.Create(payRoll).Error
}

// GetPayRollByID retrieves a payroll record by ID from the database.
//
// The function takes an ID as input and attempts to fetch the corresponding
// record from the database. It returns the PayRollModel and an error if the
// retrieval fails. If the record is not found, a nil pointer is returned together
// with a gorm.ErrRecordNotFound error.
// The function also preloads the Employee model associated with the payroll.
func (s *PayrollService) GetPayRollByID(id string) (*models.PayRollModel, error) {
	var payRoll models.PayRollModel
	err := s.db.Preload("Employee").First(&payRoll, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &payRoll, nil
}

// UpdatePayRoll updates an existing payroll record in the database.
//
// The function takes a PayRollModel as input and attempts to update
// the corresponding record in the database. It returns an error if the update
// fails, otherwise returns nil.
func (s *PayrollService) UpdatePayRoll(payRoll *models.PayRollModel) error {
	return s.db.Save(payRoll).Error
}

// DeletePayRoll deletes a payroll record by ID from the database.
//
// The function takes an ID as input and attempts to delete the corresponding
// record from the database. It returns an error if the deletion fails, otherwise
// it returns nil.
func (s *PayrollService) DeletePayRoll(id string) error {
	return s.db.Delete(&models.PayRollModel{}, "id = ?", id).Error
}

// AddItemByPayroll adds a new item to an existing payroll.
//
// The function takes a payroll ID and a PayrollItemModel as input and attempts to add
// the item to the payroll's Items association. It returns an error if the addition
// fails, otherwise returns nil.
func (s *PayrollService) AddItemByPayroll(payRollID string, item *models.PayrollItemModel) error {
	return s.db.Model(&models.PayRollModel{}).Where("id = ?", payRollID).Association("Items").Append(item)
}

// UpdateItemByPayroll updates an existing item in a payroll's Items association.
//
// The function takes a payroll ID and a PayrollItemModel as input and attempts to update
// the corresponding item in the payroll's Items association. It returns an error if the update
// fails, otherwise returns nil.
func (s *PayrollService) UpdateItemByPayroll(payRollID string, item *models.PayrollItemModel) error {
	return s.db.Model(&models.PayRollModel{}).Where("id = ?", payRollID).Association("Items").Replace(item)
}

// DeleteItemByPayroll deletes an item from a payroll's Items association.
//
// The function takes a payroll ID and a PayrollItemModel as input and attempts to delete
// the corresponding item from the payroll's Items association. It returns an error if the deletion
// fails, otherwise returns nil.
func (s *PayrollService) DeleteItemByPayroll(payRollID string, item *models.PayrollItemModel) error {
	return s.db.Model(&models.PayRollModel{}).Where("id = ?", payRollID).Association("Items").Delete(item)
}

// FindAllPayroll retrieves a paginated list of payroll records from the database.
//
// The function takes an HTTP request as input and filters the payroll records based on
// the company ID provided in the request header, if available. The function utilizes
// pagination to manage the result set and returns a paginate.Page object containing the
// list of payroll records and pagination details. If the operation fails, it returns an error.
func (s *PayrollService) FindAllPayroll(request *http.Request) (paginate.Page, error) {
	pg := paginate.New()
	stmt := s.db.Model(&models.PayRollModel{})
	if request.Header.Get("ID-Company") != "" {
		stmt = stmt.Where("company_id = ?", request.Header.Get("ID-Company"))
	}
	utils.FixRequest(request)
	page := pg.With(stmt).Request(request).Response(&[]models.PayRollModel{})
	page.Page = page.Page + 1
	return page, nil
}

// GetItemsFromPayroll retrieves all items associated with a given payroll ID.
//
// The function takes a payroll ID as input and queries the database for all
// PayrollItemModel records associated with the specified payroll. It returns
// a slice of PayrollItemModel and an error if the operation fails.

func (s *PayrollService) GetItemsFromPayroll(payRollID string) ([]models.PayrollItemModel, error) {
	var items []models.PayrollItemModel
	err := s.db.Model(&models.PayRollModel{}).Where("id = ?", payRollID).Find(&items).Error
	if err != nil {
		return nil, err
	}
	return items, nil
}

// GetPayRollCostFromPayroll retrieves all payroll cost records associated with a given payroll ID.
//
// The function takes a payroll ID as input and queries the database for all
// PayRollCostModel records related to the specified payroll. It returns a slice
// of pointers to PayRollCostModel and an error if the operation fails. If no
// records are found, an empty slice is returned without an error.
func (s *PayrollService) GetPayRollCostFromPayroll(payRollID string) ([]*models.PayRollCostModel, error) {
	var costs []*models.PayRollCostModel
	err := s.db.Model(&models.PayRollModel{}).Where("id = ?", payRollID).Find(&costs).Error
	if err != nil {
		return nil, err
	}
	return costs, nil
}

// AddPayment creates a new payment transaction associated with a payroll.
//
// The function takes a payroll ID and a TransactionModel as input. It sets the
// TransactionRefID and TransactionRefType fields of the transaction to the
// provided payroll ID and "payroll" respectively, then attempts to create a new
// transaction record in the database. It returns an error if the creation fails.
func (s *PayrollService) AddPayment(payRollID string, payment *models.TransactionModel) error {
	payment.TransactionRefID = &payRollID
	payment.TransactionRefType = "payroll"
	return s.db.Create(payment).Error
}

// DeletePayment deletes a payment transaction associated with a payroll record.
//
// The function takes the ID of the transaction to be deleted as input and
// attempts to delete the corresponding record in the database. It returns an
// error if the deletion fails.
func (s *PayrollService) DeletePayment(paymentID string) error {
	return s.db.Delete(&models.TransactionModel{}, "id = ? and transaction_ref_type = ?", paymentID, "payroll").Error
}

// UpdatePayment updates a payment transaction associated with a payroll record.
//
// The function takes a TransactionModel as input and attempts to update the
// corresponding record in the database. It returns an error if the update
// fails.
func (s *PayrollService) UpdatePayment(payment *models.TransactionModel) error {
	return s.db.Save(payment).Error
}

// ResetBPJS resets the BPJS configuration for a given payroll record to the default
// values, and removes all BPJS-related items from the payroll.
//
// The function takes a PayRollModel as input and resets the BPJS configuration
// by setting the BpjsSetting field to the default values. It then retrieves all
// items associated with the payroll and checks if each item is a BPJS-related item
// by checking the Bpjs field. If the item is a BPJS-related item, it is deleted from
// the database using the Unscoped().Delete() method. The function returns an error
// if the operation fails.
func (s *PayrollService) ResetBPJS(payroll *models.PayRollModel) error {
	bpjs := models.InitBPJS()
	payroll.BpjsSetting = bpjs
	items, _ := s.GetItemsFromPayroll(payroll.ID)
	for _, item := range items {
		if item.Bpjs {
			s.db.Unscoped().Delete(&models.PayrollItemModel{}, "id = ?", item.ID)
		}
	}

	return nil
}

// BpjsCount calculates and applies BPJS contributions for a given payroll record.
//
// This function iterates over payroll items to calculate the total salary and allowance
// that are considered for BPJS contributions. It then checks the BPJS settings enabled
// for the payroll and calculates the contributions for BPJS Kesehatan, Ketenagakerjaan JHT,
// JP, JKM, and JKK, creating corresponding payroll items and cost records in the database.
// The function returns an error if any operation fails, including missing employee or BPJS
// settings, or database insertion errors.
func (s *PayrollService) BpjsCount(payroll *models.PayRollModel) error {
	totalSalaryAndAllowance := float64(0)
	if payroll.Employee == nil {
		return errors.New("employee not found")
	}
	if payroll.BpjsSetting == nil {
		return errors.New("bpjs setting not found")
	}
	items, _ := s.GetItemsFromPayroll(payroll.ID)
	for _, item := range items {
		if item.BpjsCounted && item.ItemType != DEDUCTION {
			totalSalaryAndAllowance += item.Amount
		}
	}
	fmt.Println("TOTAL_SALARY_AND_ALLOWANCE", totalSalaryAndAllowance)
	if payroll.BpjsSetting.BpjsKesEnabled {
		employeerCost, employeeCost, _ := payroll.BpjsSetting.CalculateBPJSKes(totalSalaryAndAllowance)
		payrollItem := models.PayrollItemModel{
			ItemType:     DEDUCTION,
			IsDefault:    false,
			IsDeductible: false,
			Amount:       employeeCost,
			PayRollID:    payroll.ID,
			Title:        "BPJS Kesehatan",
			Tariff:       payroll.BpjsSetting.BpjsKesRateEmployee,
			CompanyID:    payroll.CompanyID,
			Bpjs:         true,
			Data:         "{}",
		}

		if err := s.db.Create(&payrollItem).Error; err != nil {
			return err
		}
		if err := s.db.Create(&models.PayRollCostModel{
			Description:   "Biaya BPJS Kesehatan",
			PayRollID:     payroll.ID,
			PayRollItemID: payrollItem.ID,
			Amount:        employeerCost,
			Tariff:        payroll.BpjsSetting.BpjsKesRateEmployer,
			CompanyID:     payroll.CompanyID,
		}).Error; err != nil {
			return err
		}
		if err := s.db.Create(&models.PayRollCostModel{
			Description:   "Pungutan BPJS Kesehatan",
			PayRollID:     payroll.ID,
			PayRollItemID: payrollItem.ID,
			Amount:        employeeCost,
			Tariff:        payroll.BpjsSetting.BpjsKesRateEmployee,
			CompanyID:     payroll.CompanyID,
			DebtDeposit:   true,
		}).Error; err != nil {
			return err
		}
	}
	if payroll.BpjsSetting.BpjsTkJhtEnabled {
		employeerCost, employeeCost, _ := payroll.BpjsSetting.CalculateBPJSTkJht(totalSalaryAndAllowance)
		payrollItem := models.PayrollItemModel{
			ItemType:     DEDUCTION,
			IsDefault:    false,
			IsDeductible: false,
			Amount:       employeeCost,
			PayRollID:    payroll.ID,
			Title:        "BPJS Ketenagakerjaan JHT",
			Tariff:       payroll.BpjsSetting.BpjsTkJhtRateEmployee,
			CompanyID:    payroll.CompanyID,
			Bpjs:         true,
			Data:         "{}",
		}

		if err := s.db.Create(&payrollItem).Error; err != nil {
			return err
		}
		if err := s.db.Create(&models.PayRollCostModel{
			Description:   "Biaya BPJS Ketenagakerjaan JHT",
			PayRollID:     payroll.ID,
			PayRollItemID: payrollItem.ID,
			Amount:        employeerCost,
			Tariff:        payroll.BpjsSetting.BpjsTkJhtRateEmployer,
			BpjsTkJht:     true,
			CompanyID:     payroll.CompanyID,
		}).Error; err != nil {
			return err
		}
		if err := s.db.Create(&models.PayRollCostModel{
			Description:   "Pungutan BPJS Ketenagakerjaan JHT",
			PayRollID:     payroll.ID,
			PayRollItemID: payrollItem.ID,
			Amount:        employeeCost,
			Tariff:        payroll.BpjsSetting.BpjsTkJhtRateEmployee,
			BpjsTkJht:     true,
			DebtDeposit:   true,
			CompanyID:     payroll.CompanyID,
		}).Error; err != nil {
			return err
		}
	}
	if payroll.BpjsSetting.BpjsTkJpEnabled {
		employeerCost, employeeCost, _ := payroll.BpjsSetting.CalculateBPJSTkJp(totalSalaryAndAllowance)
		payrollItem := models.PayrollItemModel{
			ItemType:     DEDUCTION,
			IsDefault:    false,
			IsDeductible: false,
			Amount:       employeeCost,
			PayRollID:    payroll.ID,
			Title:        "BPJS Ketenagakerjaan JP",
			Tariff:       payroll.BpjsSetting.BpjsTkJpRateEmployee,
			CompanyID:    payroll.CompanyID,
			Bpjs:         true,
			Data:         "{}",
		}

		if err := s.db.Create(&payrollItem).Error; err != nil {
			return err
		}
		if err := s.db.Create(&models.PayRollCostModel{
			Description:   "Biaya BPJS Ketenagakerjaan JP",
			PayRollID:     payroll.ID,
			PayRollItemID: payrollItem.ID,
			Amount:        employeerCost,
			Tariff:        payroll.BpjsSetting.BpjsTkJpRateEmployer,
			BpjsTkJp:      true,
			CompanyID:     payroll.CompanyID,
		}).Error; err != nil {
			return err
		}
		if err := s.db.Create(&models.PayRollCostModel{
			Description:   "Pungutan BPJS Ketenagakerjaan JP",
			PayRollID:     payroll.ID,
			PayRollItemID: payrollItem.ID,
			Amount:        employeeCost,
			Tariff:        payroll.BpjsSetting.BpjsTkJpRateEmployee,
			BpjsTkJp:      true,
			DebtDeposit:   true,
			CompanyID:     payroll.CompanyID,
		}).Error; err != nil {
			return err
		}
	}
	if payroll.BpjsSetting.BpjsTkJkmEnabled {
		employeeCost := payroll.BpjsSetting.CalculateBPJSTkJkm(totalSalaryAndAllowance)
		payrollItem := models.PayrollItemModel{
			ItemType:     DEDUCTION,
			IsDefault:    false,
			IsDeductible: false,
			Amount:       employeeCost,
			PayRollID:    payroll.ID,
			Title:        "BPJS Ketenagakerjaan JKM",
			Tariff:       payroll.BpjsSetting.BpjsTkJkmEmployee,
			CompanyID:    payroll.CompanyID,
			Bpjs:         true,
			Data:         "{}",
		}

		if err := s.db.Create(&payrollItem).Error; err != nil {
			return err
		}
	}
	if payroll.BpjsSetting.BpjsTkJkkEnabled {
		employeeCost, tariff := payroll.BpjsSetting.CalculateBPJSTkJkk(totalSalaryAndAllowance, payroll.Employee.WorkSafetyRisks)
		payrollItem := models.PayrollItemModel{
			ItemType:     DEDUCTION,
			IsDefault:    false,
			IsDeductible: false,
			Amount:       employeeCost,
			PayRollID:    payroll.ID,
			Title:        "BPJS Ketenagakerjaan JKK",
			Tariff:       tariff,
			CompanyID:    payroll.CompanyID,
			Bpjs:         true,
			Data:         "{}",
		}

		if err := s.db.Create(&payrollItem).Error; err != nil {
			return err
		}
	}
	return nil

}

// GetDeductible calculates the total income, reimbursement, deductible and non-deductible
// amounts from the given list of PayrollItemModel.
//
// It iterates through the list and sums the amounts of each item type.
// If the item type is DEDUCTION, it checks if the item is deductible or not and sums the
// amount accordingly.
// If the item type is REIMBURSEMENT, it sums the amount to the total reimbursement.
// If the item type is not DEDUCTION or REIMBURSEMENT, it sums the amount to the total income.
//
// The function returns four values: total income, total reimbursement, total deductible and
// total non-deductible amounts.
func (m *PayrollService) GetDeductible(items []models.PayrollItemModel) (float64, float64, float64, float64) {
	var totalIncome, totalReimbursement, totalDeductible, totalNonDeductible float64
	for _, item := range items {
		if item.IsTax {
			continue
		}
		if item.ItemType == DEDUCTION {
			if item.IsDeductible {
				totalDeductible += item.Amount
			} else {
				totalNonDeductible += item.Amount
			}
		} else {
			if item.ItemType == REIMBURSEMENT {
				totalReimbursement += item.Amount
			} else {
				totalIncome += item.Amount
			}
		}
	}

	return totalIncome, totalReimbursement, totalDeductible, totalNonDeductible
}

// GetNonDeductibleItems returns a list of PayrollItemModel that are not deductible.
//
// It takes a list of PayrollItemModel as input and iterates through the list.
// If the item type is DEDUCTION and the item is not deductible, it is added to the
// new list. If the item type is not DEDUCTION, it is skipped.
// If the item is a tax, it is skipped.
// The function returns the new list of non-deductible items.
func (m *PayrollService) GetNonDeductibleItems(items []models.PayrollItemModel) []models.PayrollItemModel {
	var newItems []models.PayrollItemModel
	for _, item := range items {
		if item.IsTax {
			continue
		}
		if item.ItemType == DEDUCTION {
			if !item.IsDeductible {
				newItems = append(newItems, item)
			} else {
				continue
			}
		} else {
			continue
		}
	}
	return newItems
}

// GetReimbursementItems returns a list of PayrollItemModel that are reimbursements.
//
// It takes a list of PayrollItemModel as input and iterates through the list.
// If the item type is REIMBURSEMENT, it is added to the new list.
// The function returns the new list of reimbursement items.
func (m *PayrollService) GetReimbursementItems(items []models.PayrollItemModel) []models.PayrollItemModel {
	var newItems []models.PayrollItemModel
	for _, item := range items {
		if item.ItemType == REIMBURSEMENT {
			newItems = append(newItems, item)
		}
	}
	return newItems
}

// GetLoans returns a list of PayrollItemModel that are loans.
//
// It takes a list of PayrollItemModel as input and iterates through the list.
// If the item type is DEDUCTION and the item has an EmployeeLoanID, it is added
// to the new list.
// The function returns the new list of loans.
func (m *PayrollService) GetLoans(items []models.PayrollItemModel) []models.PayrollItemModel {
	var newItems []models.PayrollItemModel
	for _, item := range items {
		if item.ItemType == DEDUCTION && item.EmployeeLoanID != nil {
			newItems = append(newItems, item)
		}
	}
	return newItems
}

// CountDeductible counts the total income, reimbursement, deductible, and non-deductible amounts
// from the payroll items and sets the corresponding fields on the payroll model.
//
// It takes a PayRollModel as input, iterates over the items, and calls GetDeductible
// to calculate the total income, reimbursement, deductible, and non-deductible amounts.
// It then sets the fields on the payroll model accordingly.
func (m *PayrollService) CountDeductible(payroll *models.PayRollModel) {
	totalIncome, totalReimbursement, totalDeductible, totalNonDeductible := m.GetDeductible(payroll.Items)
	payroll.TotalReimbursement = totalReimbursement
	payroll.TotalIncome = totalIncome + payroll.TaxAllowance
	payroll.TotalDeduction = totalDeductible + totalNonDeductible
	payroll.NetIncomeBeforeTaxCost = payroll.TotalIncome - totalDeductible
}

// GetTaxCost calculates the tax cost from the given payroll and non-taxable amount.
//
// If the non-taxable amount is greater than 0, it calculates the tax cost by taking
// 5% of the net income before tax cost and capping it at 500,000.
// Otherwise, it returns 0.
// The function returns the calculated tax cost.
func (m *PayrollService) GetTaxCost(payroll *models.PayRollModel, nonTaxable float64) float64 {
	taxCost := float64(0)
	if nonTaxable > 0 {
		taxCost = min((payroll.NetIncomeBeforeTaxCost)*5/100, 500000)
	}
	return taxCost
}

// EffectiveRateAverageTariff calculates the effective rate average tariff from the given payroll and category.
//
// It takes a PayRollModel, a category string, and a gross salary float64 as input.
// It calls the appropriate EffectiveRateAverageCategoryX method from the EmployeeService to get the tax tariff.
// It then calculates the tax amount by multiplying the gross salary with the tax tariff.
// The function sets the TotalTax and TaxTariff fields on the payroll model accordingly.
func (m *PayrollService) EffectiveRateAverageTariff(payroll *models.PayRollModel, category string, grossSalary float64) {
	// ac := accounting.Accounting{Symbol: "", Precision: 4}
	taxTariff := float64(0)
	switch category {
	case "A":
		taxTariff = m.employeeService.EffectiveRateAverageCategoryA(grossSalary)
	case "B":
		taxTariff = m.employeeService.EffectiveRateAverageCategoryB(grossSalary)
	case "C":
		taxTariff = m.employeeService.EffectiveRateAverageCategoryC(grossSalary)
	default:
		taxTariff = 0
	}
	taxAmount := grossSalary * taxTariff
	// fmt.Printf("GROSS SALARY %s * TAXTARIFF %s = TAXAMOUNT %s \n", utils.FormatRupiah(grossSalary), utils.FormatRupiah(taxTariff), utils.FormatRupiah(taxAmount))
	payroll.TotalTax = taxAmount
	payroll.TaxTariff = taxTariff
}

// RefreshTax resets various financial fields and tax-related amounts for the given payroll model.
//
// This function sets the amounts related to tax cost and tax to zero in the PayrollItemModel
// associated with the given payroll ID. It updates the tax allowance to zero if the payroll
// is not marked as gross up. Additionally, it resets multiple total and net income fields
// in the PayRollModel to zero, effectively clearing out any previous financial calculations.
// The function returns an error if any database operation fails during these updates.
func (m *PayrollService) RefreshTax(payroll *models.PayRollModel) error {
	m.db.Model(&models.PayrollItemModel{}).Where("pay_roll_id = ? AND is_tax_cost = ?", payroll.ID, true).Update("amount", 0)
	m.db.Model(&models.PayrollItemModel{}).Where("pay_roll_id = ? AND is_tax = ? and tax_auto_count = ?", payroll.ID, true, true).Update("amount", 0)
	if !payroll.IsGrossUp {
		err := m.db.Model(&payroll).Where("id = ?", payroll.ID).Update("tax_allowance", 0).Error
		if err != nil {
			return err
		}
		payroll.TaxAllowance = 0
	}

	payroll.TotalIncome = 0
	payroll.TotalReimbursement = 0
	payroll.TotalDeduction = 0
	payroll.TotalTax = 0
	payroll.NetIncomeBeforeTaxCost = 0
	payroll.NetIncome = 0
	payroll.TakeHomePay = 0
	payroll.TotalPayable = 0
	payroll.TaxCost = 0
	fmt.Println("RESET AMOUNT")
	return m.db.Save(payroll).Error
}

// GetTotalPayRollCost calculates the total payroll cost for a given payroll model.
//
// The function queries the database for all PayRollCostModel records associated
// with the specified payroll ID, excluding those marked as debt deposits. It
// iterates over the retrieved records and sums the amount fields to compute the
// total payroll cost. The total is returned as a float64 value.
func (m *PayrollService) GetTotalPayRollCost(payroll *models.PayRollModel) float64 {

	items := []models.PayRollCostModel{}
	m.db.Order("created_at asc").Where("debt_deposit", 0).Find(&items, "pay_roll_id = ?", payroll.ID)
	// m.Items = items
	totalCost := float64(0)
	for _, v := range items {
		totalCost += v.Amount
	}

	return totalCost
}

// RegularTaxTariff calculates the progressive tax tariff for a given payroll model.
//
// The function applies a tiered tax rate to the taxable income, reducing the taxable
// amount at each tier and adding the calculated tax to the total tax amount. There are
// five tax levels, each with a different rate: 5% for income up to 60,000,000, 15% for
// the next 250,000,000, 25% for the next 500,000,000, 30% for the next 5,000,000,000,
// and 35% for any remaining income. The function logs the details of each tax level
// calculation and updates the TotalTax field of the payroll model with the monthly
// tax amount.
func (m *PayrollService) RegularTaxTariff(payroll *models.PayRollModel, taxAmount float64, taxable float64) {
	// LEVEL 1
	if taxable > 0 {
		amountForTax := taxable - 60000000
		if amountForTax < 0 {
			amountForTax = taxable
			taxable = 0
		} else if amountForTax > 60000000 {
			taxable = amountForTax
			amountForTax = 60000000
		} else {
			taxable = amountForTax
			amountForTax = 60000000
		}
		taxValue := amountForTax * 5 / 100
		taxAmount += taxValue
		utils.LogJson(map[string]interface{}{
			"msg":                "Level 1 => 5% add tax",
			"amountForTax":       utils.FormatRupiah(amountForTax),
			"taxValue":           utils.FormatRupiah(taxValue),
			"taxValue per month": utils.FormatRupiah(taxValue / 12),
			"taxable":            utils.FormatRupiah(taxable),
			"currentTaxAmount":   utils.FormatRupiah(taxAmount),
		})
	}
	// LEVEL 2
	if taxable > 0 {
		amountForTax := taxable - 250000000
		if amountForTax < 0 {
			amountForTax = taxable
			taxable = 0
		} else if amountForTax > 250000000 {
			taxable = amountForTax
			amountForTax = 250000000
		} else {
			taxable = amountForTax
			amountForTax = 250000000
		}
		taxValue := amountForTax * 15 / 100
		taxAmount += taxValue
		utils.LogJson(map[string]interface{}{
			"msg":                "Level 2 => 15% add tax",
			"amountForTax":       utils.FormatRupiah(amountForTax),
			"taxValue":           utils.FormatRupiah(taxValue),
			"taxValue per month": utils.FormatRupiah(taxValue / 12),
			"taxable":            utils.FormatRupiah(taxable),
			"currentTaxAmount":   utils.FormatRupiah(taxAmount),
		})
	}
	// LEVEL 3
	if taxable > 0 {
		amountForTax := taxable - 500000000
		if amountForTax < 0 {
			amountForTax = taxable
			taxable = 0
		} else if amountForTax > 500000000 {
			taxable = amountForTax
			amountForTax = 500000000
		} else {
			taxable = amountForTax
			amountForTax = 500000000
		}
		taxValue := amountForTax * 25 / 100
		taxAmount += taxValue
		utils.LogJson(map[string]interface{}{
			"msg":                "Level 3 => 25% add tax",
			"amountForTax":       utils.FormatRupiah(amountForTax),
			"taxValue":           utils.FormatRupiah(taxValue),
			"taxValue per month": utils.FormatRupiah(taxValue / 12),
			"taxable":            utils.FormatRupiah(taxable),
			"currentTaxAmount":   utils.FormatRupiah(taxAmount),
		})
	}
	// LEVEL 4
	if taxable > 0 {
		amountForTax := taxable - 5000000000
		if amountForTax < 0 {
			amountForTax = taxable
			taxable = 0
		} else if amountForTax > 5000000000 {
			taxable = amountForTax
			amountForTax = 5000000000
		} else {
			taxable = amountForTax
			amountForTax = 5000000000
		}
		taxValue := amountForTax * 30 / 100
		taxAmount += taxValue
		utils.LogJson(map[string]interface{}{
			"msg":                "Level 4 => 30% add tax",
			"amountForTax":       utils.FormatRupiah(amountForTax),
			"taxValue":           utils.FormatRupiah(taxValue),
			"taxValue per month": utils.FormatRupiah(taxValue / 12),
			"taxable":            utils.FormatRupiah(taxable),
			"currentTaxAmount":   utils.FormatRupiah(taxAmount),
		})
	}
	// LEVEL 5
	if taxable > 0 {
		taxValue := taxable * 35 / 100
		taxAmount += taxValue
		utils.LogJson(map[string]interface{}{
			"msg":                "Level 5 => 35% add tax",
			"taxValue":           utils.FormatRupiah(taxValue),
			"taxValue per month": utils.FormatRupiah(taxValue / 12),
			"taxable":            utils.FormatRupiah(taxable),
			"currentTaxAmount":   utils.FormatRupiah(taxAmount),
		})
	}
	utils.LogJson(map[string]interface{}{
		"taxAmount":           utils.FormatRupiah(taxAmount),
		"taxAmount per month": utils.FormatRupiah(taxAmount / 12),
	})
	payroll.TotalTax = taxAmount / 12
}

// CountTax counts the tax of a payroll. It takes a PayRollModel as input and sets the TotalTax and TaxTariff fields of the model.
// It also updates the TakeHomePay and TaxSummary fields of the model.
// If the payroll has a non-taxable income, it calls the CountDeductible function to count the deductible.
// If the payroll has a tax cost, it calls the CountTaxCost function to count the tax cost.
// If the payroll is an effective rate average, it calls the EffectiveRateAverageTariff function to calculate the tax tariff.
// If the payroll is not an effective rate average, it calls the RegularTaxTariff function to calculate the tax tariff.
// It also sets the TaxCost field of the model to the total tax cost.
// Finally, it updates the payroll record in the database.
func (m *PayrollService) CountTax(payroll *models.PayRollModel) error {
	if payroll.Employee == nil {
		return errors.New("employee not found")
	}
	fmt.Println("COUNT_TAX")

	// ac := accounting.Accounting{Symbol: "", Precision: 0}
	m.RefreshTax(payroll)
	items, _ := m.GetItemsFromPayroll(payroll.ID)
	payroll.Items = items

	nonTaxable := m.employeeService.GetNonTaxableIncomeLevelAmount(payroll.Employee)
	nonTaxableCategory := m.employeeService.GetNonTaxableIncomeLevelCategory(payroll.Employee)

	countTaxRecord := int64(0)
	taxCost := float64(0)

	taxCostItem := models.PayrollItemModel{}
	// taxItem := PayRollItem{}
	m.db.Find(&taxCostItem, "pay_roll_id = ? AND is_tax_cost = ?", payroll.ID, true)
	m.db.Model(&models.PayrollItemModel{}).Where("pay_roll_id = ? AND is_tax = ? and tax_auto_count = ?", payroll.ID, true, true).Count(&countTaxRecord)
	m.CountDeductible(payroll)

	taxCost = m.GetTaxCost(payroll, nonTaxable)
	if payroll.IsEffectiveRateAverage {
		taxCost = 0
	}

	payroll.NetIncome = payroll.NetIncomeBeforeTaxCost - taxCost
	m.db.Model(&taxCostItem).Where("id = ?", taxCostItem.ID).Update("amount", taxCost)

	m.db.Model(&payroll).Update("tax_cost", taxCost)
	fmt.Println("TAX COST", utils.FormatRupiah(taxCost))
	fmt.Println("NET INCOME AFTER REDUCE TAX COST", utils.FormatRupiah(payroll.NetIncomeBeforeTaxCost))
	fmt.Println("TAX AUTO COUNT", countTaxRecord)
	// GET TAX COST
	// 1. GET NET INCOME
	items, _ = m.GetItemsFromPayroll(payroll.ID)
	payroll.Items = items

	yearlyNetIncome := (payroll.NetIncome + m.GetTotalPayRollCost(payroll)) * 12

	taxable := yearlyNetIncome - nonTaxable

	utils.LogJson(map[string]interface{}{
		"yearlyGrossIncome":    utils.FormatRupiah(payroll.NetIncomeBeforeTaxCost * 12),
		"taxCost":              utils.FormatRupiah(taxCost * 12),
		"yearlyNetIncome":      utils.FormatRupiah(yearlyNetIncome),
		"taxable":              utils.FormatRupiah(taxable),
		"nonTaxable":           utils.FormatRupiah(nonTaxable),
		"taxAllowancePerMonth": utils.FormatRupiah(payroll.TaxAllowance),
	})
	var taxAmount float64 = payroll.TotalTax
	if countTaxRecord == 0 {
		taxAmount = 0
		payroll.TotalTax = 0
		m.db.Model(&payroll).Update("total_tax", 0)

	}

	if nonTaxable != 0 && countTaxRecord > 0 {
		if payroll.IsEffectiveRateAverage {
			fmt.Println("NON_TAXABLE_CATEGORY", nonTaxableCategory)
			m.EffectiveRateAverageTariff(payroll, nonTaxableCategory, payroll.NetIncomeBeforeTaxCost+m.GetTotalPayRollCost(payroll))
		} else {
			payroll.TaxTariff = 0
			m.RegularTaxTariff(payroll, taxAmount, taxable)
			m.db.Model(&payroll).Update("tax_tariff", 0)

		}
	}

	payroll.TakeHomePay = payroll.TotalIncome - payroll.TotalDeduction - payroll.TotalTax - payroll.TaxCost
	m.db.Model(&models.PayrollItemModel{}).Where("pay_roll_id = ? AND is_tax = true and tax_auto_count = true", payroll.ID).Update("amount", payroll.TotalTax)
	if err := m.db.Model(&payroll).Updates(&payroll).Error; err != nil {
		return err
	}

	if payroll.TotalReimbursement == 0 {
		if err := m.db.Model(&payroll).Update("total_reimbursement", 0).Error; err != nil {
			return err
		}
	}

	if payroll.IsGrossUp {
		fmt.Println("m.TaxAllowance", payroll.TaxAllowance)
		fmt.Println("m.TotalTax", payroll.TotalTax)
		if payroll.TotalTax != payroll.TaxAllowance {
			payroll.TaxAllowance = payroll.TotalTax
			if err := m.db.Model(&payroll).Updates(&payroll).Error; err != nil {
				return err
			}
			fmt.Println("m.TaxAllowance UPDATED", payroll.TaxAllowance)
			return m.CountTax(payroll)
		}
	}

	payroll.TaxSummary = models.CountTaxSummary{
		JobExpenseMonthly:               taxCost,
		JobExpenseYearly:                taxCost * 12,
		PtkpYearly:                      nonTaxable,
		GrossIncomeMonthly:              payroll.NetIncomeBeforeTaxCost,
		GrossIncomeYearly:               payroll.NetIncomeBeforeTaxCost * 12,
		PkpMonthly:                      (payroll.NetIncome*12 - nonTaxable) / 12,
		PkpYearly:                       payroll.NetIncome*12 - nonTaxable,
		TaxYearlyBasedOnProgressiveRate: payroll.TotalTax * 12,
		TaxYearly:                       payroll.TotalTax * 12,
		TaxMonthly:                      payroll.TotalTax,
		NetIncomeMonthly:                payroll.NetIncome,
		NetIncomeYearly:                 payroll.NetIncome,
		CutoffPensiunMonthly:            0,
		CutoffPensiunYearly:             0,
		CutoffMonthly:                   0,
		CutoffYearly:                    0,
		Ter:                             0,
	}

	return nil

}
