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

func (s *PayrollService) CreatePayRoll(payRoll *models.PayRollModel) error {
	return s.db.Create(payRoll).Error
}

func (s *PayrollService) GetPayRollByID(id string) (*models.PayRollModel, error) {
	var payRoll models.PayRollModel
	err := s.db.Preload("Employee").First(&payRoll, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &payRoll, nil
}

func (s *PayrollService) UpdatePayRoll(payRoll *models.PayRollModel) error {
	return s.db.Save(payRoll).Error
}

func (s *PayrollService) DeletePayRoll(id string) error {
	return s.db.Delete(&models.PayRollModel{}, "id = ?", id).Error
}

func (s *PayrollService) AddItemByPayroll(payRollID string, item *models.PayrollItemModel) error {
	return s.db.Model(&models.PayRollModel{}).Where("id = ?", payRollID).Association("Items").Append(item)
}

func (s *PayrollService) UpdateItemByPayroll(payRollID string, item *models.PayrollItemModel) error {
	return s.db.Model(&models.PayRollModel{}).Where("id = ?", payRollID).Association("Items").Replace(item)
}

func (s *PayrollService) DeleteItemByPayroll(payRollID string, item *models.PayrollItemModel) error {
	return s.db.Model(&models.PayRollModel{}).Where("id = ?", payRollID).Association("Items").Delete(item)
}

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

func (s *PayrollService) GetItemsFromPayroll(payRollID string) ([]models.PayrollItemModel, error) {
	var items []models.PayrollItemModel
	err := s.db.Model(&models.PayRollModel{}).Where("id = ?", payRollID).Association("Items").Find(&items)
	if err != nil {
		return nil, err
	}
	return items, nil
}

func (s *PayrollService) GetPayRollCostFromPayroll(payRollID string) ([]*models.PayRollCostModel, error) {
	var costs []*models.PayRollCostModel
	err := s.db.Model(&models.PayRollModel{}).Where("id = ?", payRollID).Association("Costs").Find(&costs)
	if err != nil {
		return nil, err
	}
	return costs, nil
}

func (s *PayrollService) AddPayment(payRollID string, payment *models.TransactionModel) error {
	payment.TransactionRefID = &payRollID
	payment.TransactionRefType = "payroll"
	return s.db.Create(payment).Error
}

func (s *PayrollService) DeletePayment(paymentID string) error {
	return s.db.Delete(&models.TransactionModel{}, "id = ? and transaction_ref_type = ?", paymentID, "payroll").Error
}

func (s *PayrollService) UpdatePayment(payment *models.TransactionModel) error {
	return s.db.Save(payment).Error
}

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

func (m *PayrollService) GetReimbursementItems(items []models.PayrollItemModel) []models.PayrollItemModel {
	var newItems []models.PayrollItemModel
	for _, item := range items {
		if item.ItemType == REIMBURSEMENT {
			newItems = append(newItems, item)
		}
	}
	return newItems
}
func (m *PayrollService) GetLoans(items []models.PayrollItemModel) []models.PayrollItemModel {
	var newItems []models.PayrollItemModel
	for _, item := range items {
		if item.ItemType == DEDUCTION && item.EmployeeLoanID != nil {
			newItems = append(newItems, item)
		}
	}
	return newItems
}

func (m *PayrollService) CountDeductible(payroll *models.PayRollModel) {
	totalIncome, totalReimbursement, totalDeductible, totalNonDeductible := m.GetDeductible(payroll.Items)
	payroll.TotalReimbursement = totalReimbursement
	payroll.TotalIncome = totalIncome + payroll.TaxAllowance
	payroll.TotalDeduction = totalDeductible + totalNonDeductible
	payroll.NetIncomeBeforeTaxCost = payroll.TotalIncome - totalDeductible
}

func (m *PayrollService) GetTaxCost(payroll *models.PayRollModel, nonTaxable float64) float64 {
	taxCost := float64(0)
	if nonTaxable > 0 {
		taxCost = min((payroll.NetIncomeBeforeTaxCost)*5/100, 500000)
	}
	return taxCost
}

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
