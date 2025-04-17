package asset

import (
	"errors"
	"fmt"
	"math"
	"net/http"
	"time"

	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/shared"
	"github.com/AMETORY/ametory-erp-modules/shared/models"
	"github.com/AMETORY/ametory-erp-modules/utils"
	"github.com/AMETORY/ametory-erp-modules/utils/fin"
	"github.com/google/uuid"
	"github.com/morkid/paginate"
	"gorm.io/gorm"
)

type AssetService struct {
	db  *gorm.DB
	ctx *context.ERPContext
}

func NewAssetService(db *gorm.DB, ctx *context.ERPContext) *AssetService {
	return &AssetService{ctx: ctx, db: db}
}

func Migrate(db *gorm.DB) error {
	fmt.Println("Migrating account model...")
	return db.AutoMigrate(&models.AssetModel{}, &models.DepreciationCostModel{})
}

func (s *AssetService) CreateAsset(data *models.AssetModel) error {
	return s.db.Create(data).Error
}

func (s *AssetService) UpdateAsset(id string, data *models.AssetModel) error {
	return s.db.Where("id = ?", id).Updates(data).Error
}

func (s *AssetService) DeleteAsset(id string) error {
	return s.db.Where("id = ?", id).Delete(&models.AssetModel{}).Error
}

func (s *AssetService) GetAssetByID(id string) (*models.AssetModel, error) {
	var invoice models.AssetModel
	err := s.db.Where("id = ?", id).First(&invoice).Error
	return &invoice, err
}

func (s *AssetService) GetAssets(request http.Request, search string) (paginate.Page, error) {
	pg := paginate.New()
	stmt := s.db
	if search != "" {
		stmt = stmt.Where("description ILIKE ? OR asset_number ILIKE ? OR name ILIKE ?",
			"%"+search+"%",
			"%"+search+"%",
			"%"+search+"%",
		)
	}
	if request.Header.Get("ID-Company") != "" {
		stmt = stmt.Where("company_id = ?", request.Header.Get("ID-Company"))
	}
	request.URL.Query().Get("page")
	stmt = stmt.Model(&models.AssetModel{})
	utils.FixRequest(&request)
	page := pg.With(stmt).Request(request).Response(&[]models.AssetModel{})
	page.Page = page.Page + 1
	return page, nil
}

func (s *AssetService) CountDepreciation(asset *models.AssetModel) ([]float64, error) {
	// db := db
	// if m.Tx != nil {
	// 	db = m.Tx
	// }

	if !asset.IsDepreciationAsset {
		return []float64{}, errors.New("please set to depreciation asset")
	}
	costs := []float64{}
	switch asset.DepreciationMethod {
	case "SLN":
		dep, _ := fin.DepreciationStraightLine(asset.AcquisitionCost, asset.SalvageValue, int(asset.LifeTime))
		for i := 0; i < int(asset.LifeTime); i++ {
			costs = append(costs, dep)
		}
	case "DB":
		for i := 1; i <= int(asset.LifeTime); i++ {
			dep, _ := fin.DepreciationFixedDeclining(asset.AcquisitionCost, asset.SalvageValue, int(asset.LifeTime), i, 12)
			costs = append(costs, dep)
		}
	case "SYD":
		for i := 1; i <= int(asset.LifeTime); i++ {
			dep := fin.DepreciationSYD(asset.AcquisitionCost, asset.SalvageValue, int(asset.LifeTime), i)
			costs = append(costs, dep)
		}
	default:
		return []float64{}, errors.New(asset.DepreciationMethod + "not implemented")
	}

	return costs, nil
}

func (s *AssetService) PreviewCosts(asset *models.AssetModel) ([]models.DepreciationCostModel, error) {
	costs, err := s.CountDepreciation(asset)
	if err != nil {
		return nil, err
	}

	depreciationCosts := []models.DepreciationCostModel{}
	for i, v := range costs {
		if asset.IsMonthly {
			for j := 1; j <= 12; j++ {
				depreciationCosts = append(depreciationCosts, models.DepreciationCostModel{
					Month:  j,
					Amount: v / 12,
					Period: i + 1,
				})
			}

		} else {
			depreciationCosts = append(depreciationCosts, models.DepreciationCostModel{
				Amount: v,
				Period: i + 1,
			})
		}

	}
	return depreciationCosts, nil
}

func (s *AssetService) ActivateAsset(asset *models.AssetModel, date time.Time, userID string) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		if asset.Status != "DRAFT" {
			return errors.New("asset is not in draft status")
		}

		now := time.Now()

		depreciationCosts, err := s.PreviewCosts(asset)
		if err != nil {
			return err
		}

		for _, v := range depreciationCosts {
			if asset.IsMonthly {
				diff := math.Ceil(now.Sub(asset.Date).Hours() / 24 / 30)
				if diff > float64(((v.Period-1)*12)+v.Month) {
					v.Status = "ACTIVE"
				}
			} else {
				diff := math.Ceil(now.Sub(asset.Date).Hours() / 24 / 365)
				if diff > float64((v.Period - 1)) {
					if v.Status == "PENDING" {
						v.Status = "ACTIVE"
					}
				}
			}
			v.UserID = &userID
			v.AssetID = &asset.ID
			v.CompanyID = asset.CompanyID
			tx.Create(&v)
		}

		return nil
	})
}

func (s *AssetService) DepreciationApply(asset *models.AssetModel, itemID string, date time.Time, userID string) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		depreciation := models.DepreciationCostModel{}
		if err := tx.Find(&depreciation, "asset_id = ? and uu_id = ? AND status = ?", asset.ID, itemID, "ACTIVE").Error; err != nil {
			return err
		}
		asset.BookValue -= depreciation.Amount

		// CREATE COST TRANSACTION
		code := utils.RandString(10, false)
		costTransID := uuid.New().String()
		depreciationTransID := uuid.New().String()

		costTrans := models.TransactionModel{
			BaseModel:                   shared.BaseModel{ID: costTransID},
			Code:                        code,
			CompanyID:                   asset.CompanyID,
			UserID:                      &userID,
			Debit:                       depreciation.Amount,
			Amount:                      depreciation.Amount,
			AccountID:                   asset.AccountDepreciationID,
			Description:                 "Biaya Penyusutan " + asset.AssetNumber,
			Date:                        date,
			TransactionRefID:            &depreciationTransID,
			TransactionRefType:          "transaction",
			TransactionSecondaryRefID:   &asset.ID,
			TransactionSecondaryRefType: "asset",
		}
		if err := tx.Create(&costTrans).Error; err != nil {
			return err
		}

		// CREATE DEPRECIATION TRANSACTION

		depreciationTrans := models.TransactionModel{
			BaseModel:                   shared.BaseModel{ID: depreciationTransID},
			Code:                        code,
			CompanyID:                   asset.CompanyID,
			UserID:                      &userID,
			Credit:                      depreciation.Amount,
			Amount:                      depreciation.Amount,
			AccountID:                   asset.AccountAccumulatedDepreciationID,
			Description:                 "Akumulasi Penyusutan " + asset.AssetNumber,
			Date:                        date,
			TransactionRefID:            &costTransID,
			TransactionRefType:          "transaction",
			TransactionSecondaryRefID:   &asset.ID,
			TransactionSecondaryRefType: "asset",
		}
		if err := tx.Create(&depreciationTrans).Error; err != nil {
			return err
		}

		if err := tx.Save(asset).Error; err != nil {
			return err
		}
		if err := tx.Model(&depreciation).Where("uu_id = ?", depreciation.ID).Updates(map[string]any{
			"status":      "DONE",
			"executed_at": date,
		}).Error; err != nil {
			return err
		}

		return nil
	})

}

func (s *AssetService) DepreciationCancel(asset *models.AssetModel, itemID string, date time.Time, userID string) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		depreciation := models.DepreciationCostModel{}
		if err := tx.Find(&depreciation, "asset_id = ? and uu_id = ? AND status = ?", asset.ID, itemID, "ACTIVE").Error; err != nil {
			return err
		}
		asset.BookValue += depreciation.Amount

		if err := tx.Model(&depreciation).Where("uu_id = ?", depreciation.ID).Updates(map[string]any{
			"status":      "PENDING",
			"executed_at": nil,
		}).Error; err != nil {
			return err
		}

		return nil
	})
}

func (s *AssetService) DepreciationDone(asset *models.AssetModel, itemID string, date time.Time, userID string) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		depreciation := models.DepreciationCostModel{}
		if err := tx.Find(&depreciation, "asset_id = ? and uu_id = ? AND status = ?", asset.ID, itemID, "ACTIVE").Error; err != nil {
			return err
		}
		asset.BookValue -= depreciation.Amount

		if err := tx.Model(&depreciation).Where("uu_id = ?", depreciation.ID).Updates(map[string]any{
			"status":      "DONE",
			"executed_at": date,
		}).Error; err != nil {
			return err
		}

		return nil
	})
}

func (s *AssetService) DepreciationPending(asset *models.AssetModel, itemID string, date time.Time, userID string) error {
	return s.db.Transaction(func(tx *gorm.DB) error {

		depreciation := models.DepreciationCostModel{}
		if err := tx.Find(&depreciation, "asset_id = ? and uu_id = ? AND status = ?", asset.ID, itemID, "ACTIVE").Error; err != nil {
			return err
		}
		if depreciation.Status != "ACTIVE" {
			return errors.New("depreciation is not active")
		}
		asset.BookValue += depreciation.Amount

		if err := tx.Model(&depreciation).Where("uu_id = ?", depreciation.ID).Updates(map[string]any{
			"status":      "PENDING",
			"executed_at": nil,
		}).Error; err != nil {
			return err
		}

		return nil
	})
}

func (s *AssetService) DepreciationActive(asset *models.AssetModel, itemID string, date time.Time, userID string) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		depreciation := models.DepreciationCostModel{}
		if err := tx.Find(&depreciation, "asset_id = ? and uu_id = ? AND status = ?", asset.ID, itemID, "PENDING").Error; err != nil {
			return err
		}
		asset.BookValue -= depreciation.Amount

		if err := tx.Model(&depreciation).Where("uu_id = ?", depreciation.ID).Updates(map[string]any{
			"status":      "ACTIVE",
			"executed_at": date,
		}).Error; err != nil {
			return err
		}

		return nil
	})
}

func (s *AssetService) GetDepreciation(asset *models.AssetModel) {
	now := time.Now()
	depreciations := []models.DepreciationCostModel{}
	s.db.Find(&depreciations, "asset_id = ?", asset.ID)
	asset.Depreciations = depreciations
	if asset.IsMonthly {
		diff := math.Ceil(now.Sub(asset.Date).Hours() / 24 / 30)
		for _, v := range depreciations {
			if diff > float64(((v.Period-1)*12)+v.Month) {
				if v.Status == "PENDING" {
					s.db.Model(v).Where("uu_id = ?", v.ID).Update("status", "ACTIVE")
				}
			}
		}
	} else {
		diff := math.Ceil(now.Sub(asset.Date).Hours() / 24 / 365)
		for _, v := range depreciations {
			if diff > float64((v.Period - 1)) {
				if v.Status == "PENDING" {
					s.db.Model(v).Where("uu_id = ?", v.ID).Update("status", "ACTIVE")
				}
			}
		}
	}

	s.db.Find(&depreciations, "asset_id = ?", asset.ID)
	asset.Depreciations = depreciations
}
