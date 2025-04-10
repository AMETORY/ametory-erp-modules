package net_surplus

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/cooperative/cooperative_setting"
	"github.com/AMETORY/ametory-erp-modules/finance"
	"github.com/AMETORY/ametory-erp-modules/shared"
	"github.com/AMETORY/ametory-erp-modules/shared/models"
	"github.com/AMETORY/ametory-erp-modules/utils"
	"github.com/morkid/paginate"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type NetSurplusService struct {
	db                        *gorm.DB
	ctx                       *context.ERPContext
	cooperativeSettingService *cooperative_setting.CooperativeSettingService
	financeService            *finance.FinanceService
}

func NewNetSurplusService(
	db *gorm.DB,
	ctx *context.ERPContext,
	cooperativeSettingService *cooperative_setting.CooperativeSettingService,
	financeService *finance.FinanceService,
) *NetSurplusService {
	return &NetSurplusService{
		db:                        db,
		ctx:                       ctx,
		cooperativeSettingService: cooperativeSettingService,
		financeService:            financeService,
	}
}

func (s *NetSurplusService) SetDB(db *gorm.DB) {
	s.db = db
}

func (n *NetSurplusService) GetTransactions(netSurplusID string) []models.TransactionModel {
	var transactions []models.TransactionModel

	n.db.Model(&models.TransactionModel{}).Preload("Account").
		Where("net_surplus_id = ?", netSurplusID).
		Find(&transactions)

	return transactions
}

func (n *NetSurplusService) GetNetSurplusTotal(netSurplus *models.NetSurplusModel) error {
	profitLoss := models.ProfitLoss{

		GeneralReport: models.GeneralReport{
			CompanyID: *netSurplus.CompanyID,
			Title:     "Sisa Hasil Usaha",
			StartDate: netSurplus.StartDate,
			EndDate:   netSurplus.EndDate,
		},
	}

	profitLossData, err := n.financeService.ReportService.GenerateProfitLossReport(profitLoss.GeneralReport)
	if err != nil {
		return err
	}
	// c, err := json.Marshal(profitLoss)
	// if err != nil {
	// 	return err
	// }

	// err := profitLoss.Generate(c)
	// if err != nil {
	// 	return err
	// }

	netSurplus.NetSurplusTotal = profitLossData.NetProfit
	totalTransactions := float64(0)
	totalSaving := float64(0)
	totalLoan := float64(0)
	var savings []models.SavingModel
	n.db.Where("company_id = ? and (date between ? and ?) and net_surplus_id is null", *netSurplus.CompanyID, netSurplus.StartDate, netSurplus.EndDate).Find(&savings)
	for _, s := range savings {
		totalSaving += s.Amount
	}
	var loans []models.LoanApplicationModel
	n.db.Where("company_id = ? and (submission_date between ? and ?) and net_surplus_id is null and (status = ? OR status = ?)", *netSurplus.CompanyID, netSurplus.StartDate, netSurplus.EndDate, "SETTLEMENT", "DISBURSED").Find(&loans)
	for _, s := range loans {
		totalLoan += s.LoanAmount
	}

	var invoices []models.SalesModel
	n.db.Where("company_id = ? and member_id  is not null and net_surplus_id is null", *netSurplus.CompanyID).Find(&invoices)
	for _, s := range invoices {
		totalTransactions += s.Total
	}
	fmt.Printf("total transactions: %f\ntotalLoan: %f\n", totalTransactions, totalLoan)
	if totalTransactions+totalLoan == 0 {
		return errors.New("total transactions is zero")
	}
	if totalSaving == 0 {
		return errors.New("total saving is zero")
	}

	netSurplus.LoanTotal = totalLoan
	netSurplus.TransactionTotal = totalTransactions
	netSurplus.SavingsTotal = totalSaving
	// fmt.Println(profitLoss)
	profitLossData.StartDate = netSurplus.StartDate
	profitLossData.EndDate = netSurplus.EndDate
	b, err := json.Marshal(profitLossData)
	if err != nil {
		return err
	}
	*netSurplus.ProfitLossData = string(b)

	return n.db.Save(&netSurplus).Error
}

func (n *NetSurplusService) CreateDistribution(netSurplus *models.NetSurplusModel) error {

	setting, err := n.cooperativeSettingService.GetSetting(netSurplus.CompanyID)
	if err != nil {
		return err
	}

	allocations := []models.NetSurplusAllocation{}
	allocations = append(allocations, models.NetSurplusAllocation{
		Name:       "Jasa Modal",
		Percentage: setting.NetSurplusMandatorySavings,
		Amount:     utils.AmountRound(netSurplus.NetSurplusTotal*setting.NetSurplusMandatorySavings/100, 2),
		AccountID:  setting.NetSurplusMandatorySavingsAccountID,
		Key:        "net_surplus_mandatory_savings",
	})
	allocations = append(allocations, models.NetSurplusAllocation{
		Name:       "Dana Cadangan",
		Percentage: setting.NetSurplusReserve,
		Amount:     utils.AmountRound(netSurplus.NetSurplusTotal*setting.NetSurplusReserve/100, 2),
		AccountID:  setting.NetSurplusReserveAccountID,
		Key:        "net_surplus_reserve",
	})
	allocations = append(allocations, models.NetSurplusAllocation{
		Name:       "Jasa Usaha",
		Percentage: setting.NetSurplusBusinessProfit,
		Amount:     utils.AmountRound(netSurplus.NetSurplusTotal*setting.NetSurplusBusinessProfit/100, 2),
		AccountID:  setting.NetSurplusBusinessProfitAccountID,
		Key:        "net_surplus_business_profit",
	})
	allocations = append(allocations, models.NetSurplusAllocation{
		Name:       "Dana Sosial",
		Percentage: setting.NetSurplusSocialFund,
		Amount:     utils.AmountRound(netSurplus.NetSurplusTotal*setting.NetSurplusSocialFund/100, 2),
		AccountID:  setting.NetSurplusSocialFundAccountID,
		Key:        "net_surplus_social_fund",
	})
	allocations = append(allocations, models.NetSurplusAllocation{
		Name:       "Dana Pendidikan",
		Percentage: setting.NetSurplusEducationFund,
		Amount:     utils.AmountRound(netSurplus.NetSurplusTotal*setting.NetSurplusEducationFund/100, 2),
		AccountID:  setting.NetSurplusEducationFundAccountID,
		Key:        "net_surplus_education_fund",
	})
	allocations = append(allocations, models.NetSurplusAllocation{
		Name:       "Dana Pengurus",
		Percentage: setting.NetSurplusManagement,
		Amount:     utils.AmountRound(netSurplus.NetSurplusTotal*setting.NetSurplusManagement/100, 2),
		AccountID:  setting.NetSurplusManagementAccountID,
		Key:        "net_surplus_management",
	})
	allocations = append(allocations, models.NetSurplusAllocation{
		Name:       "Dana Lainnya",
		Percentage: setting.NetSurplusOtherFunds,
		Amount:     utils.AmountRound(netSurplus.NetSurplusTotal*setting.NetSurplusOtherFunds/100, 2),
		AccountID:  setting.NetSurplusOtherFundsAccountID,
		Key:        "net_surplus_other_funds",
	})
	netSurplus.Distribution = allocations
	b, err := json.Marshal(allocations)
	if err != nil {
		return err
	}
	// utils.LogJson(allocations)
	netSurplus.DistributionData = string(b)

	return nil
}

func (s *NetSurplusService) GetNetSurplusList(request http.Request, search string, memberID *string) (paginate.Page, error) {
	pg := paginate.New()
	stmt := s.db
	if search != "" {
		stmt = stmt.Where("description ILIKE ?",
			"%"+search+"%",
		)
	}
	if request.Header.Get("ID-Company") != "" {
		stmt = stmt.Where("company_id = ? or company_id is null", request.Header.Get("ID-Company"))
	}
	if memberID != nil {
		stmt = stmt.Where("member_id = ?", memberID)
	}
	request.URL.Query().Get("page")
	stmt = stmt.Model(&models.NetSurplusModel{})
	utils.FixRequest(&request)
	page := pg.With(stmt).Request(request).Response(&[]models.NetSurplusModel{})
	page.Page = page.Page + 1
	return page, nil
}

func (s *NetSurplusService) GetNetSurplusByID(id string, memberID *string) (*models.NetSurplusModel, error) {
	var netSurplus models.NetSurplusModel
	db := s.db.Preload(clause.Associations)
	if memberID != nil {
		db = db.Where("member_id = ?", memberID)
	}
	if err := db.Where("id = ?", id).First(&netSurplus).Error; err != nil {
		return nil, err
	}
	trans := s.GetTransactions(id)
	netSurplus.Transactions = trans

	return &netSurplus, nil
}
func (c *NetSurplusService) CreateNetSurplus(netSurplus *models.NetSurplusModel) error {
	if err := c.db.Create(netSurplus).Error; err != nil {
		return err
	}

	err := c.GetNetSurplusTotal(netSurplus)
	if err != nil {
		return err
	}
	err = c.CreateDistribution(netSurplus)
	if err != nil {
		return err
	}
	err = c.GetMembers(netSurplus)
	if err != nil {
		return err
	}

	c.GenNumber(netSurplus, netSurplus.CompanyID)

	return c.db.Save(netSurplus).Error

}

func (c *NetSurplusService) UpdateNetSurplus(id string, netSurplus *models.NetSurplusModel) error {

	err := c.db.Where("id = ?", id).Save(netSurplus).Error
	if err != nil {
		return err
	}

	return nil
}

func (s *NetSurplusService) DeleteNetSurplus(id string) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		err := tx.Where("net_surplus_id = ?", id).Delete(&models.TransactionModel{}).Error
		if err != nil {
			return err
		}
		return tx.Where("id = ?", id).Delete(&models.NetSurplusModel{}).Error
	})
}

func (n *NetSurplusService) GetMembers(netSurplus *models.NetSurplusModel) error {
	// getCompany, _ := c.Get("companySession")
	// company := getCompany.(CompanyModel)
	// company.GetCooperativeSetting()
	// setting := company.CooperativeSetting
	var memberData []models.CooperativeMemberModel
	n.db.Find(&memberData, "company_id = ?", netSurplus.CompanyID)

	var members []models.NetSurplusMember

	for _, member := range memberData {
		totalTransactions := float64(0)
		totalSaving := float64(0)
		totalLoan := float64(0)
		var savings []models.SavingModel
		n.db.Where("company_id = ? and member_id = ? and (date between ? and ?) and net_surplus_id is null", netSurplus.CompanyID, member.ID, netSurplus.StartDate, netSurplus.EndDate).Find(&savings)
		for _, s := range savings {
			totalSaving += s.Amount
			s.NetSurplusID = &netSurplus.ID
			n.db.Save(&s)
		}
		var loans []models.LoanApplicationModel
		n.db.Where("company_id = ? and member_id = ? and (submission_date between ? and ?) and net_surplus_id is null and (status = ? OR status = ?)", netSurplus.CompanyID, member.ID, netSurplus.StartDate, netSurplus.EndDate, "SETTLEMENT", "DISBURSED").Find(&loans)
		for _, s := range loans {
			totalLoan += s.LoanAmount
			s.NetSurplusID = &netSurplus.ID
			n.db.Save(&s)
		}

		// fmt.Println("totalSaving", totalSaving)

		var invoices []models.SalesModel
		err := n.db.Where("company_id = ? and member_id = ? and net_surplus_id is null", netSurplus.CompanyID, member.ID).Find(&invoices).Error
		if err != nil {
			// fmt.Println("ERROR", err)
			return err
		}
		for _, s := range invoices {
			totalTransactions += s.Total
			s.NetSurplusID = &netSurplus.ID
			n.db.Save(&s)
		}
		// fmt.Println("totalTransactions", totalTransactions)

		var savingAllocation, transactionAllocation float64
		for _, d := range netSurplus.Distribution {
			if d.Key == "net_surplus_mandatory_savings" {
				savingAllocation = d.Amount

			}
			if d.Key == "net_surplus_business_profit" {
				transactionAllocation = d.Amount

			}
		}

		fmt.Printf("TOTAL TRANS + LOAN: %f\n,NET  TOTAL TRANS + LOAN: %f\nALLOCATION %f\n", (totalTransactions + totalLoan), (netSurplus.TransactionTotal + netSurplus.LoanTotal), transactionAllocation)

		members = append(members, models.NetSurplusMember{
			ID:                                   member.ID,
			FullName:                             member.Name,
			MemberID:                             member.MemberIDNumber,
			SavingsTotal:                         utils.AmountRound(totalSaving, 2),
			LoanTotal:                            utils.AmountRound(totalLoan, 2),
			TransactionTotal:                     utils.AmountRound(totalTransactions, 2),
			NetSurplusMandatorySavingsAllocation: utils.AmountRound(totalSaving/netSurplus.SavingsTotal*savingAllocation, 2),
			NetSurplusBusinessProfitAllocation:   utils.AmountRound((totalTransactions+totalLoan)/(netSurplus.TransactionTotal+netSurplus.LoanTotal)*transactionAllocation, 2),
			Status:                               "PENDING",
		})

	}
	// fmt.Println(members)
	b, err := json.Marshal(members)
	if err != nil {
		return err
	}
	netSurplus.MemberData = string(b)

	return nil
}

func (c *NetSurplusService) GenNumber(netSurplus *models.NetSurplusModel, companyID *string) error {
	setting, err := c.cooperativeSettingService.GetSetting(companyID)
	if err != nil {
		return err
	}
	lastLoan := models.NetSurplusModel{}
	nextNumber := ""
	data := shared.InvoiceBillSettingModel{
		StaticCharacter:       setting.NetSurplusStaticCharacter,
		NumberFormat:          setting.NumberFormat,
		AutoNumericLength:     setting.AutoNumericLength,
		RandomNumericLength:   setting.RandomNumericLength,
		RandomCharacterLength: setting.RandomCharacterLength,
	}
	if err := c.db.Where("company_id = ?", companyID).Limit(1).Order("created_at desc").Find(&lastLoan).Error; err != nil {
		nextNumber = shared.GenerateInvoiceBillNumber(data, "00")
	} else {
		nextNumber = shared.ExtractNumber(data, lastLoan.NetSurplusNumber)
	}

	netSurplus.NetSurplusNumber = nextNumber
	return nil
}

func (n *NetSurplusService) Distribute(netSurplus *models.NetSurplusModel, sourceID string, allocations []models.NetSurplusAllocation, userID string) error {

	now := time.Now()
	b, err := json.Marshal(allocations)
	if err != nil {
		return err
	}
	return n.db.Transaction(func(tx *gorm.DB) error {
		netSurplus.DistributionData = string(b)
		netSurplus.Status = "DISTRIBUTED"
		if err := tx.Save(&n).Error; err != nil {
			return err
		}
		// CREATE TRANSACTION NET SURPLUS DISTRIBUTION
		for _, v := range allocations {
			if v.AccountCashID == nil {
				return errors.New("account cash id is required")
			}
			equityID := utils.Uuid()
			assetID := utils.Uuid()
			err := tx.Create(&models.TransactionModel{
				BaseModel:                   shared.BaseModel{ID: equityID},
				Date:                        now,
				UserID:                      &userID,
				CompanyID:                   netSurplus.CompanyID,
				Credit:                      utils.AmountRound(v.Amount, 2),
				Description:                 fmt.Sprintf("Distribusi Alokasi SHU : %s", v.Name),
				NetSurplusID:                &netSurplus.ID,
				AccountID:                   v.AccountID,
				TransactionRefID:            &assetID,
				TransactionRefType:          "transaction",
				TransactionSecondaryRefID:   &netSurplus.ID,
				TransactionSecondaryRefType: v.Key,
			}).Error
			if err != nil {
				return err
			}
			err = tx.Create(&models.TransactionModel{
				BaseModel:                   shared.BaseModel{ID: assetID},
				Date:                        now,
				UserID:                      &userID,
				CompanyID:                   netSurplus.CompanyID,
				Credit:                      utils.AmountRound(v.Amount, 2),
				Description:                 fmt.Sprintf("Distribusi Alokasi SHU : %s", v.Name),
				NetSurplusID:                &netSurplus.ID,
				AccountID:                   v.AccountCashID,
				TransactionRefID:            &equityID,
				TransactionRefType:          "transaction",
				TransactionSecondaryRefID:   &netSurplus.ID,
				TransactionSecondaryRefType: v.Key,
			}).Error
			if err != nil {
				return err
			}

		}
		return nil
	})

}
