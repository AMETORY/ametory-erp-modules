package account

import (
	"net/http"

	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/morkid/paginate"
	"gorm.io/gorm"
)

type AccountService struct {
	db  *gorm.DB
	ctx *context.ERPContext
}

func NewAccountService(db *gorm.DB, ctx *context.ERPContext) *AccountService {
	return &AccountService{db: db, ctx: ctx}
}

func (s *AccountService) CreateAccount(data *AccountModel) error {
	return s.db.Create(data).Error
}

func (s *AccountService) UpdateAccount(id string, data *AccountModel) error {
	return s.db.Where("id = ?", id).Updates(data).Error
}

func (s *AccountService) DeleteAccount(id string) error {
	return s.db.Where("id = ?", id).Delete(&AccountModel{}).Error
}

func (s *AccountService) GetAccountByID(id string) (*AccountModel, error) {
	var account AccountModel
	err := s.db.Where("id = ?", id).First(&account).Error
	return &account, err
}

func (s *AccountService) GetAccountByCode(code string) (*AccountModel, error) {
	var account AccountModel
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
		stmt = stmt.Where("accounts.name LIKE ? OR accounts.code LIKE ? ",
			"%"+search+"%",
			"%"+search+"%",
		)
	}
	stmt = stmt.Model(&AccountModel{})
	page := pg.With(stmt).Request(request).Response(&[]AccountModel{})
	return page, nil
}
