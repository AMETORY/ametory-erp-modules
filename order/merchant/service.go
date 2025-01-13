package merchant

import (
	"net/http"

	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/finance"
	"github.com/AMETORY/ametory-erp-modules/utils"
	"github.com/morkid/paginate"
	"gorm.io/gorm"
)

type MerchantService struct {
	ctx            *context.ERPContext
	db             *gorm.DB
	financeService *finance.FinanceService
}

func NewMerchantService(db *gorm.DB, ctx *context.ERPContext, financeService *finance.FinanceService) *MerchantService {
	return &MerchantService{db: db, ctx: ctx, financeService: financeService}
}
func (s *MerchantService) GetNearbyMerchants(lat, lng float64, radius float64) ([]MerchantModel, error) {
	var merchants []MerchantModel

	rows, err := s.db.Raw(`
		SELECT *, (
			6371 * acos(
				cos(radians(?)) * cos(radians(latitude)) * cos(radians(longitude) - radians(?)) +
				sin(radians(?)) * sin(radians(latitude))
			)
		) AS distance
		FROM merchant
		HAVING distance <= ?
		ORDER BY distance
	`, lat, lng, lat, radius).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var merchant MerchantModel
		if err := s.db.ScanRows(rows, &merchant); err != nil {
			return nil, err
		}
		merchants = append(merchants, merchant)
	}
	return merchants, err
}

func (s *MerchantService) CreateMerchant(data *MerchantModel) error {
	return s.db.Create(data).Error
}

func (s *MerchantService) UpdateMerchant(id string, data *MerchantModel) error {
	return s.db.Where("id = ?", id).Updates(data).Error
}

func (s *MerchantService) DeleteMerchant(id string) error {
	return s.db.Where("id = ?", id).Delete(&MerchantModel{}).Error
}

func (s *MerchantService) GetMerchantByID(id string) (*MerchantModel, error) {
	var invoice MerchantModel
	err := s.db.Where("id = ?", id).First(&invoice).Error
	return &invoice, err
}

func (s *MerchantService) GetMerchants(request http.Request, search string) (paginate.Page, error) {
	pg := paginate.New()
	stmt := s.db
	if search != "" {
		stmt = stmt.Where("brands.description ILIKE ? OR brands.name ILIKE ?",
			"%"+search+"%",
			"%"+search+"%",
		)
	}
	if request.Header.Get("ID-Company") != "" {
		stmt = stmt.Where("company_id = ?", request.Header.Get("ID-Company"))
	}
	request.URL.Query().Get("page")
	stmt = stmt.Model(&MerchantModel{})
	utils.FixRequest(&request)
	page := pg.With(stmt).Request(request).Response(&[]MerchantModel{})
	page.Page = page.Page + 1
	return page, nil
}