package net_surplus

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/cooperative/cooperative_setting"
	"github.com/AMETORY/ametory-erp-modules/cooperative/saving"
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
	savingService             *saving.SavingService
}

// NewNetSurplusService creates a new instance of NetSurplusService.
//
// The NetSurplusService is used to manage net surpluses, which are
// distributions of profits made by a cooperative.
//
// It takes as parameters:
//   - db: a pointer to a GORM database instance
//   - ctx: a pointer to an ERPContext, which contains the user's HTTP request
//     context and other relevant information
//   - cooperativeSettingService: a pointer to a CooperativeSettingService,
//     which is used to look up the settings for the cooperative
//   - financeService: a pointer to a FinanceService, which is used to manage
//     financial transactions
//   - savingService: a pointer to a SavingService, which is used to manage
//     savings accounts
func NewNetSurplusService(
	db *gorm.DB,
	ctx *context.ERPContext,
	cooperativeSettingService *cooperative_setting.CooperativeSettingService,
	financeService *finance.FinanceService,
	savingService *saving.SavingService,
) *NetSurplusService {
	return &NetSurplusService{
		db:                        db,
		ctx:                       ctx,
		cooperativeSettingService: cooperativeSettingService,
		financeService:            financeService,
		savingService:             savingService,
	}
}

// SetDB sets the underlying database connection for the NetSurplusService.
//
// This function should be used with caution, as it can potentially lead to
// unexpected behavior if the underlying database connection is changed
// unexpectedly.
func (s *NetSurplusService) SetDB(db *gorm.DB) {
	s.db = db
}

// GetTransactions retrieves a list of transactions associated with a specific net surplus ID.
// The transactions are preloaded with their related account information and sorted by date in ascending order.
//
// Parameters:
//   - netSurplusID: A string representing the unique identifier of the net surplus.
//
// Returns:
//   - A slice of TransactionModel objects containing the transactions linked to the specified net surplus ID.

func (n *NetSurplusService) GetTransactions(netSurplusID string) []models.TransactionModel {
	var transactions []models.TransactionModel

	n.db.Model(&models.TransactionModel{}).Preload("Account").
		Where("net_surplus_id = ?", netSurplusID).
		Order("date ASC").
		Find(&transactions)

	return transactions
}

// GetNetSurplusTotal calculates and updates the net surplus total for a given NetSurplusModel.
//
// This function retrieves the closing book associated with the provided net surplus using the ClosingBookID,
// and calculates the total savings, loans, and transactions within the specified date range. It updates the
// net surplus model with these calculated totals and the net income from the closing book's closing summary.
//
// Parameters:
//   - tx: A pointer to the GORM database transaction.
//   - netSurplus: A pointer to the NetSurplusModel for which the net surplus total is being calculated.
//
// Returns:
//   - An error if any database operation fails or if required data is missing.

func (n *NetSurplusService) GetNetSurplusTotal(tx *gorm.DB, netSurplus *models.NetSurplusModel) error {
	if netSurplus.ClosingBookID == nil {
		return errors.New("closing book id is required")
	}
	var closingBook models.ClosingBook
	err := tx.Model(&models.ClosingBook{}).Where("id = ?", netSurplus.ClosingBookID).First(&closingBook).Error
	if err != nil {
		return err
	}

	profitLossData := closingBook.ProfitLoss

	// c, err := json.Marshal(profitLoss)
	// if err != nil {
	// 	return err
	// }

	// err := profitLoss.Generate(c)
	// if err != nil {
	// 	return err
	// }

	netSurplus.NetSurplusTotal = closingBook.ClosingSummary.NetIncome
	totalTransactions := float64(0)
	totalSaving := float64(0)
	totalLoan := float64(0)
	var savings []models.SavingModel
	tx.Where("company_id = ? and (date between ? and ?) and net_surplus_id is null", *netSurplus.CompanyID, netSurplus.StartDate, netSurplus.EndDate).Find(&savings)
	for _, s := range savings {
		totalSaving += s.Amount
	}
	var loans []models.LoanApplicationModel
	tx.Where("company_id = ? and (submission_date between ? and ?) and net_surplus_id is null and (status = ? OR status = ?)", *netSurplus.CompanyID, netSurplus.StartDate, netSurplus.EndDate, "SETTLEMENT", "DISBURSED").Debug().Find(&loans)
	for _, s := range loans {
		totalLoan += s.LoanAmount
	}

	var invoices []models.SalesModel
	tx.Where("company_id = ? and member_id  is not null and net_surplus_id is null", *netSurplus.CompanyID).Find(&invoices)
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

	return tx.Save(&netSurplus).Error
}

// CreateDistribution create distribution of net surplus based on company setting
//
// Net surplus distribution is divided into 6 categories:
// 1. Jasa Modal (Mandatory Savings)
// 2. Dana Cadangan (Reserve)
// 3. Jasa Usaha (Business Profit)
// 4. Dana Sosial (Social Fund)
// 5. Dana Pendidikan (Education Fund)
// 6. Dana Pengurus (Management)
// 7. Dana Lainnya (Other Funds)
//
// The percentage of each category is based on the company setting
// The amount of each category is calculated based on the net surplus total
// and the percentage of each category
func (n *NetSurplusService) CreateDistribution(tx *gorm.DB, netSurplus *models.NetSurplusModel) error {

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

// GetNetSurplusList retrieves a paginated list of net surplus records.
//
// It accepts an HTTP request, a search query string, and an optional member ID. The search
// query is applied to the description field of the net surplus records. If a company ID is
// present in the request header, results are filtered by the company ID or if the company ID
// is null. Additionally, if a member ID is provided, results are further filtered by the member ID.
// Pagination is applied to manage the result set, and any necessary request modifications are
// handled using the utils.FixRequest utility.
//
// The function returns a paginated page of NetSurplusModel and an error if the operation fails.

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

// GetNetSurplusByID retrieves a net surplus record by ID.
//
// It accepts an ID and an optional member ID. If a company ID is present in the request
// header, results are filtered by the company ID or if the company ID is null. Additionally,
// if a member ID is provided, results are further filtered by the member ID. The function
// returns a NetSurplusModel and an error if the operation fails.
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

// CreateNetSurplus creates a new net surplus record in the database.
//
// This function initiates a database transaction to ensure atomicity.
// It first creates the net surplus record, and then calculates and updates
// the net surplus total. It also creates the distribution of net surplus
// based on company settings, retrieves the members associated with the net surplus,
// and generates a unique net surplus number. If any operation fails, the transaction
// is rolled back and an error is returned.
//
// Parameters:
//   - netSurplus: A pointer to the NetSurplusModel to be created.
//
// Returns:
//   - An error if any operation within the transaction fails.

func (c *NetSurplusService) CreateNetSurplus(netSurplus *models.NetSurplusModel) error {
	err := c.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(netSurplus).Error; err != nil {
			return err
		}

		err := c.GetNetSurplusTotal(tx, netSurplus)
		if err != nil {
			return err
		}
		err = c.CreateDistribution(tx, netSurplus)
		if err != nil {
			return err
		}
		err = c.GetMembers(tx, netSurplus)
		if err != nil {
			return err
		}

		c.GenNumber(tx, netSurplus, netSurplus.CompanyID)

		return tx.Save(netSurplus).Error
	})
	if err != nil {
		return err
	}
	return nil

}

// UpdateNetSurplus updates an existing net surplus record in the database.
//
// This function updates the net surplus record with the provided NetSurplusModel.
// If the operation fails, an error is returned.
//
// Parameters:
//   - id: The ID of the net surplus record to be updated.
//   - netSurplus: A pointer to the NetSurplusModel that contains the updated values.
//
// Returns:
//   - An error if the operation fails.
func (c *NetSurplusService) UpdateNetSurplus(id string, netSurplus *models.NetSurplusModel) error {

	err := c.db.Where("id = ?", id).Save(netSurplus).Error
	if err != nil {
		return err
	}

	return nil
}

// DeleteNetSurplus deletes a net surplus record from the database.
//
// This function deletes the net surplus record with the given ID, and also
// removes any associated transactions, loan applications, and savings records.
// If any operation fails, an error is returned.
//
// Parameters:
//   - id: The ID of the net surplus record to be deleted.
//
// Returns:
//   - An error if the operation fails.
func (s *NetSurplusService) DeleteNetSurplus(id string) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		err := tx.Where("net_surplus_id = ?", id).Delete(&models.TransactionModel{}).Error
		if err != nil {
			return err
		}
		err = tx.Where("net_surplus_id = ?", id).Model(&models.LoanApplicationModel{}).Update("net_surplus_id", nil).Error
		if err != nil {
			return err
		}
		err = tx.Where("net_surplus_id = ?", id).Model(&models.SavingModel{}).Update("net_surplus_id", nil).Error
		if err != nil {
			return err
		}
		return tx.Where("id = ?", id).Delete(&models.NetSurplusModel{}).Error
	})
}

// GetMembers retrieves and processes cooperative members for a specified net surplus.
//
// This function fetches the cooperative members associated with a given company ID from the
// database, calculates their total savings, loans, and transactions within the net surplus
// period, and updates their records with the net surplus ID. It also calculates the allocation
// of mandatory savings and business profit for each member based on their contributions and
// the net surplus distribution settings.
//
// Parameters:
//   - tx: A pointer to the GORM database transaction.
//   - netSurplus: A pointer to the NetSurplusModel that contains the net surplus details.
//
// Returns:
//   - An error if any database operation fails or if required data is missing.

func (n *NetSurplusService) GetMembers(tx *gorm.DB, netSurplus *models.NetSurplusModel) error {
	// getCompany, _ := c.Get("companySession")
	// company := getCompany.(CompanyModel)
	// company.GetCooperativeSetting()
	// setting := company.CooperativeSetting
	var memberData []models.CooperativeMemberModel
	err := tx.Find(&memberData, "company_id = ?", netSurplus.CompanyID).Error
	if err != nil {
		return err
	}

	var members []models.NetSurplusMember

	for _, member := range memberData {
		totalTransactions := float64(0)
		totalSaving := float64(0)
		totalLoan := float64(0)
		var savings []models.SavingModel
		tx.Where("company_id = ? and cooperative_member_id = ? and (date between ? and ?) and net_surplus_id is null", netSurplus.CompanyID, member.ID, netSurplus.StartDate, netSurplus.EndDate).Find(&savings)
		for _, s := range savings {
			totalSaving += s.Amount
			s.NetSurplusID = &netSurplus.ID
			err := tx.Save(&s).Error
			if err != nil {
				return err
			}
		}
		var loans []models.LoanApplicationModel
		tx.Where("company_id = ? and member_id = ? and (submission_date between ? and ?) and net_surplus_id is null and (status = ? OR status = ?)", netSurplus.CompanyID, member.ID, netSurplus.StartDate, netSurplus.EndDate, "SETTLEMENT", "DISBURSED").Find(&loans)
		for _, s := range loans {
			totalLoan += s.LoanAmount
			s.NetSurplusID = &netSurplus.ID
			err := tx.Save(&s).Error
			if err != nil {
				return err
			}
		}

		// fmt.Println("totalSaving", totalSaving)

		var invoices []models.SalesModel
		err := tx.Where("company_id = ? and member_id = ? and net_surplus_id is null", netSurplus.CompanyID, member.ID).Find(&invoices).Error
		if err != nil {
			// fmt.Println("ERROR", err)
			return err
		}
		for _, s := range invoices {
			totalTransactions += s.Total
			s.NetSurplusID = &netSurplus.ID
			err := tx.Save(&s).Error
			if err != nil {
				return err
			}
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

// GenNumber generates the next number for a new net surplus. It queries the database to get the latest
// net surplus number for the given company, and then uses the invoice bill setting to generate the next
// number. If the query fails, it falls back to generating the number from the invoice bill setting
// with a prefix of "00".
func (c *NetSurplusService) GenNumber(tx *gorm.DB, netSurplus *models.NetSurplusModel, companyID *string) error {
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
	if err := tx.Where("company_id = ?", companyID).Limit(1).Order("created_at desc").Find(&lastLoan).Error; err != nil {
		nextNumber = shared.GenerateInvoiceBillNumber(data, "00")
	} else {
		nextNumber = shared.ExtractNumber(data, lastLoan.NetSurplusNumber)
	}

	netSurplus.NetSurplusNumber = nextNumber
	return nil
}

// Disbursement disburses the net surplus to the members' accounts.
// The method first generates the transaction numbers for the disbursement.
// Then, it creates the transaction for each member, and updates the net surplus member data.
// It also creates a saving transaction for the voluntary asset, if it is given.
// The method returns an error if there is an error in generating the transaction numbers,
// or in creating the transaction, or in updating the net surplus member data.
func (n *NetSurplusService) Disbursement(date time.Time, members []models.NetSurplusMember, netSurplus *models.NetSurplusModel, destinationID, userID, notes string, voluntaryAssetID *string) error {
	return n.db.Transaction(func(tx *gorm.DB) error {

		var accountMandatoryID, accountBusinessProfitID string
		for _, v := range netSurplus.Distribution {
			if v.Key == "net_surplus_mandatory_savings" {
				if v.AccountID == nil {
					return errors.New("account equity id is required")
				}
				accountMandatoryID = *v.AccountID
			}

			if v.Key == "net_surplus_business_profit" {
				if v.AccountID == nil {
					return errors.New("account equity id is required")
				}
				accountBusinessProfitID = *v.AccountID
			}

		}
		for i, v := range members {
			if v.Status == "DISBURSED" {
				continue
			}
			mandatoryID := utils.Uuid()
			businessProfitID := utils.Uuid()
			assetID := utils.Uuid()

			// MANDATORY SAVINGS
			if v.NetSurplusMandatorySavingsAllocation > 0 {
				err := tx.Create(&models.TransactionModel{
					Code:                        utils.RandString(10, false),
					BaseModel:                   shared.BaseModel{ID: mandatoryID},
					Date:                        date,
					UserID:                      &userID,
					CompanyID:                   netSurplus.CompanyID,
					Debit:                       utils.AmountRound(v.NetSurplusMandatorySavingsAllocation, 2),
					Amount:                      utils.AmountRound(v.NetSurplusMandatorySavingsAllocation, 2),
					Description:                 fmt.Sprintf("Pencairan SHU Jasa Modal [%s]: %s", netSurplus.NetSurplusNumber, v.FullName),
					Notes:                       notes,
					NetSurplusID:                &netSurplus.ID,
					AccountID:                   &accountMandatoryID,
					TransactionRefID:            &assetID,
					TransactionRefType:          "transaction",
					TransactionSecondaryRefID:   &netSurplus.ID,
					TransactionSecondaryRefType: "net_surplus_mandatory_savings",
				}).Error
				if err != nil {
					return err
				}
			}
			// BUSINESS PROFIT
			if v.NetSurplusBusinessProfitAllocation > 0 {
				err := tx.Create(&models.TransactionModel{
					Code:                        utils.RandString(10, false),
					BaseModel:                   shared.BaseModel{ID: businessProfitID},
					Date:                        date,
					UserID:                      &userID,
					CompanyID:                   netSurplus.CompanyID,
					Debit:                       utils.AmountRound(v.NetSurplusBusinessProfitAllocation, 2),
					Amount:                      utils.AmountRound(v.NetSurplusBusinessProfitAllocation, 2),
					Description:                 fmt.Sprintf("Pencairan SHU Jasa Usaha [%s]: %s", netSurplus.NetSurplusNumber, v.FullName),
					Notes:                       notes,
					NetSurplusID:                &netSurplus.ID,
					AccountID:                   &accountBusinessProfitID,
					TransactionRefID:            &assetID,
					TransactionRefType:          "transaction",
					TransactionSecondaryRefID:   &netSurplus.ID,
					TransactionSecondaryRefType: "net_surplus_business_profit",
				}).Error
				if err != nil {
					return err
				}
			}

			// DISBURSEMENT
			if v.NetSurplusBusinessProfitAllocation+v.NetSurplusMandatorySavingsAllocation > 0 {
				err := tx.Create(&models.TransactionModel{
					BaseModel:          shared.BaseModel{ID: assetID},
					Code:               utils.RandString(10, false),
					Date:               date,
					UserID:             &userID,
					CompanyID:          netSurplus.CompanyID,
					Credit:             utils.AmountRound(v.NetSurplusBusinessProfitAllocation+v.NetSurplusMandatorySavingsAllocation, 2),
					Amount:             utils.AmountRound(v.NetSurplusBusinessProfitAllocation+v.NetSurplusMandatorySavingsAllocation, 2),
					Description:        fmt.Sprintf("Pencairan SHU [%s]: %s", netSurplus.NetSurplusNumber, v.FullName),
					NetSurplusID:       &netSurplus.ID,
					AccountID:          &destinationID,
					TransactionRefID:   &netSurplus.ID,
					TransactionRefType: "net-surplus",
				}).Error
				if err != nil {
					return err
				}

				if voluntaryAssetID != nil {
					saving := models.SavingModel{
						CompanyID:            netSurplus.CompanyID,
						UserID:               &userID,
						CooperativeMemberID:  &v.ID,
						AccountDestinationID: voluntaryAssetID,
						SavingType:           "VOLUNTARY",
						Amount:               utils.AmountRound(v.NetSurplusBusinessProfitAllocation+v.NetSurplusMandatorySavingsAllocation, 2),
						Notes:                fmt.Sprintf("Konversi SHU [%s]: %s", netSurplus.NetSurplusNumber, v.FullName),
						Date:                 &date,
					}

					if err := tx.Create(&saving).Error; err != nil {
						return err
					}
					var company models.CompanyModel
					err := tx.Where("id = ?", netSurplus.CompanyID).First(&company).Error
					if err != nil {
						return err
					}
					saving.Company = &company

					err = n.savingService.CreateTransaction(saving, true)
					if err != nil {
						return err
					}
				}
			}

			v.Status = "DISBURSED"
			members[i] = v
		}

		netSurplus.Members = members
		b, err := json.Marshal(members)
		if err != nil {
			return err
		}

		if err := tx.Model(&netSurplus).Where("id = ?", netSurplus.ID).Updates(map[string]any{
			"member_data": string(b),
		}).Error; err != nil {
			return err
		}

		return nil
	})
}

// Distribute distribute net surplus to equity and other accounts
//
// The function will delete all transaction with net surplus id
// and create new transaction for each allocation
//
// The function will also create a transaction for net surplus distribution
// with debit account is sourceID and credit account is equityID
//
// # The function will return error if net surplus is already distributed
//
// The function will return error if there is an error when creating transaction
func (n *NetSurplusService) Distribute(netSurplus *models.NetSurplusModel, sourceID string, allocations []models.NetSurplusAllocation, userID string) error {

	if netSurplus.Status == "DISTRIBUTED" {
		return errors.New("net surplus already distributed")
	}

	now := time.Now()
	b, err := json.Marshal(allocations)
	if err != nil {
		return err
	}
	return n.db.Transaction(func(tx *gorm.DB) error {
		err := tx.Where("net_surplus_id = ?", netSurplus.ID).Unscoped().Delete(&models.TransactionModel{}).Error
		if err != nil {
			return err
		}
		if err := tx.Model(&netSurplus).Where("id = ?", netSurplus.ID).Updates(map[string]any{
			"status":            "DISTRIBUTED",
			"distribution_data": string(b),
		}).Error; err != nil {
			return err
		}
		assetID := utils.Uuid()
		totalNetSurplus := 0.0

		// CREATE TRANSACTION NET SURPLUS DISTRIBUTION
		for _, v := range allocations {
			// if v.AccountCashID == nil {
			// 	return errors.New("account cash id is required")
			// }

			equityID := utils.Uuid()

			err := tx.Create(&models.TransactionModel{
				Code:                        utils.RandString(10, false),
				BaseModel:                   shared.BaseModel{ID: equityID},
				Date:                        now,
				UserID:                      &userID,
				CompanyID:                   netSurplus.CompanyID,
				Credit:                      utils.AmountRound(v.Amount, 2),
				Amount:                      utils.AmountRound(v.Amount, 2),
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
			totalNetSurplus += utils.AmountRound(v.Amount, 2)

		}

		err = tx.Create(&models.TransactionModel{
			Code:               utils.RandString(10, false),
			BaseModel:          shared.BaseModel{ID: assetID},
			Date:               now,
			UserID:             &userID,
			CompanyID:          netSurplus.CompanyID,
			Debit:              utils.AmountRound(netSurplus.NetSurplusTotal, 2),
			Amount:             utils.AmountRound(netSurplus.NetSurplusTotal, 2),
			Description:        fmt.Sprintf("SHU Tahun Berjalan : %s", netSurplus.NetSurplusNumber),
			NetSurplusID:       &netSurplus.ID,
			AccountID:          &sourceID,
			TransactionRefID:   &netSurplus.ID,
			TransactionRefType: "net-surplus",
			IsNetSurplus:       true,
		}).Error
		if err != nil {
			return err
		}
		return nil
	})

}
