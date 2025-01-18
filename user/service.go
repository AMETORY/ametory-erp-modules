package user

import (
	"net/http"

	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/shared/models"
	"github.com/AMETORY/ametory-erp-modules/utils"
	"github.com/morkid/paginate"
	"gorm.io/gorm"
)

type UserService struct {
	erpContext *context.ERPContext
	db         *gorm.DB
}

func NewUserService(erpContext *context.ERPContext) *UserService {
	return &UserService{erpContext: erpContext, db: erpContext.DB}
}

func (service *UserService) GetUserByID(userID string) (*models.UserModel, error) {
	user := &models.UserModel{}
	if err := service.db.Where("id = ?", userID).First(user).Error; err != nil {
		return nil, err
	}
	return user, nil
}

func (s *UserService) GetUsers(request http.Request, search string) (paginate.Page, error) {
	pg := paginate.New()
	stmt := s.db
	if search != "" {
		stmt = stmt.Where("users.full_name ILIKE ? OR users.email ILIKE ? OR users.phone_number ILIKE ?",
			"%"+search+"%",
			"%"+search+"%",
			"%"+search+"%",
		)
	}

	stmt = stmt.Model(&models.UserModel{})
	utils.FixRequest(&request)
	page := pg.With(stmt).Request(request).Response(&[]models.UserModel{})
	page.Page = page.Page + 1
	items := page.Items.(*[]models.UserModel)
	newItems := make([]models.UserModel, 0)

	for _, v := range *items {
		file := models.FileModel{}
		s.db.Where("ref_id = ? and ref_type = ?", v.ID, "user").First(&file)
		if file.ID != "" {
			v.ProfilePicture = &file
		}
		newItems = append(newItems, v)
	}
	page.Items = &newItems
	return page, nil
}
