package member

import (
	"net/http"

	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/shared/models"
	"github.com/AMETORY/ametory-erp-modules/utils"
	"github.com/morkid/paginate"
	"gorm.io/gorm"
)

type MemberService struct {
	db  *gorm.DB
	ctx *context.ERPContext
}

func NewMemberService(ctx *context.ERPContext) *MemberService {
	return &MemberService{
		db:  ctx.DB,
		ctx: ctx,
	}
}

func (s *MemberService) CreateMember(data *models.MemberModel) error {
	return s.db.Create(data).Error
}

func (s *MemberService) UpdateMember(id string, data *models.MemberModel) error {
	return s.db.Where("id = ?", id).Updates(data).Error
}

func (s *MemberService) DeleteMember(id string) error {
	return s.db.Where("id = ?", id).Delete(&models.MemberModel{}).Error
}

func (s *MemberService) GetMemberByID(id string) (*models.MemberModel, error) {
	var invoice models.MemberModel
	err := s.db.Where("id = ?", id).First(&invoice).Error
	return &invoice, err
}

func (s *MemberService) GetMembers(request http.Request, search string) (paginate.Page, error) {
	pg := paginate.New()
	stmt := s.db.Preload("User").Preload("Role")
	if search != "" {
		stmt = stmt.
			Joins("LEFT JOIN users ON users.id = members.user_id")
		stmt = stmt.Where("users.full_name ILIKE ? OR users.email ILIKE ? OR users.phone_number ILIKE ?",
			"%"+search+"%",
			"%"+search+"%",
			"%"+search+"%",
		)

	}
	if request.Header.Get("ID-Company") != "" {
		stmt = stmt.Where("company_id = ?", request.Header.Get("ID-Company"))
	}
	request.URL.Query().Get("page")
	stmt = stmt.Model(&models.MemberModel{})
	utils.FixRequest(&request)
	page := pg.With(stmt).Request(request).Response(&[]models.MemberModel{})
	page.Page = page.Page + 1
	return page, nil
}

func (s *MemberService) InviteMember(data *models.MemberInvitationModel) (string, error) {
	if data.Token == "" {
		data.Token = utils.RandString(32, false)
	}
	err := s.db.Create(&data).Error
	if err != nil {
		return "", err
	}
	return data.Token, nil
}

func (s *MemberService) GetInvitedMembers(request http.Request, search string) (paginate.Page, error) {
	pg := paginate.New()
	stmt := s.db.Preload("Inviter").Preload("Role")

	if request.Header.Get("ID-Company") != "" {
		stmt = stmt.Where("company_id = ?", request.Header.Get("ID-Company"))
	}
	if search != "" {
		stmt = stmt.Where("full_name ILIKE ?",
			"%"+search+"%",
		)
	}

	utils.FixRequest(&request)
	stmt = stmt.Model(&models.MemberInvitationModel{})
	page := pg.With(stmt).Request(request).Response(&[]models.MemberInvitationModel{})
	utils.FixRequest(&request)
	page.Page = page.Page + 1
	return page, nil
}

func (s *MemberService) AcceptMemberInvitation(token string, userID string) error {
	var invitation models.MemberInvitationModel
	if err := s.db.Where("token = ?", token).First(&invitation).Error; err != nil {
		return err
	}
	data := models.MemberModel{
		UserID:    userID,
		CompanyID: invitation.CompanyID,
		RoleID:    invitation.RoleID,
	}
	err := s.db.Create(&data).Error
	if err != nil {
		return err
	}

	if invitation.TeamID != nil {
		var team models.TeamModel
		if err := s.db.Where("id = ?", *invitation.TeamID).First(&team).Error; err == nil {
			team.Members = append(team.Members, data)
			s.db.Save(&team)
		}

	}
	var role models.RoleModel
	if err := s.db.Where("id = ?", invitation.RoleID).First(&role).Error; err == nil {
		var user models.UserModel
		if err := s.db.Where("id = ?", userID).First(&user).Error; err == nil {
			user.Roles = append(user.Roles, role)
			s.db.Save(&user)
		}
	}

	err = s.db.Delete(&invitation).Error
	if err != nil {
		return err
	}

	// add role to user

	return nil
}
