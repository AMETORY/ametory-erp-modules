package net_surplus

import (
	"encoding/json"

	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/cooperative/cooperative_setting"
	"github.com/AMETORY/ametory-erp-modules/finance"
	"github.com/AMETORY/ametory-erp-modules/shared/models"
	"gorm.io/gorm"
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

	n.financeService.ReportService.GenerateProfitLoss(&profitLoss)
	// c, err := json.Marshal(profitLoss)
	// if err != nil {
	// 	return err
	// }

	// err := profitLoss.Generate(c)
	// if err != nil {
	// 	return err
	// }

	netSurplus.NetSurplusTotal = profitLoss.Amount
	totalTransactions := float64(0)
	totalSaving := float64(0)
	totalLoan := float64(0)
	var savings []models.SavingModel
	n.db.Where("company_id = ? and (date between ? and ?) and net_surplus_id is null", *netSurplus.CompanyID, netSurplus.StartDate, netSurplus.EndDate).Find(&savings)
	for _, s := range savings {
		totalSaving += s.Amount
	}
	var loans []models.LoanApplicationModel
	n.db.Where("company_id = ? and (submission_date between ? and ?) and net_surplus_id is null and (status = ? OR ?)", *netSurplus.CompanyID, netSurplus.StartDate, netSurplus.EndDate, "Settlement", "Disbursed").Find(&loans)
	for _, s := range loans {
		totalLoan += s.LoanAmount
	}

	var invoices []models.SalesModel
	n.db.Where("company_id = ? and member_id  is not null and net_surplus_id is null", *netSurplus.CompanyID).Find(&invoices)
	for _, s := range invoices {
		totalTransactions += s.Total
	}

	netSurplus.LoanTotal = totalLoan
	netSurplus.TransactionTotal = totalTransactions
	netSurplus.SavingsTotal = totalSaving

	b, err := json.Marshal(profitLoss)
	netSurplus.ProfitLossData = string(b)
	return err
}
