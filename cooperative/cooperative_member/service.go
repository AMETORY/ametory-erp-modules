package cooperative_member

import (
	"net/http"

	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/shared/models"
	"github.com/AMETORY/ametory-erp-modules/utils"
	"github.com/morkid/paginate"
	"gorm.io/gorm"
)

type CooperativeMemberService struct {
	db  *gorm.DB
	ctx *context.ERPContext
}

func NewCooperativeMemberService(db *gorm.DB, ctx *context.ERPContext) *CooperativeMemberService {
	return &CooperativeMemberService{
		db:  db,
		ctx: ctx,
	}
}

func (s *CooperativeMemberService) CreateMember(data *models.CooperativeMemberModel) error {
	return s.db.Create(data).Error
}

func (s *CooperativeMemberService) UpdateMember(id string, data *models.CooperativeMemberModel) error {
	return s.db.Where("id = ?", id).Updates(data).Error
}

func (s *CooperativeMemberService) DeleteMember(id string) error {
	return s.db.Where("id = ?", id).Delete(&models.CooperativeMemberModel{}).Error
}

func (s *CooperativeMemberService) GetMemberByID(id string) (*models.CooperativeMemberModel, error) {
	var invoice models.CooperativeMemberModel
	err := s.db.Where("id = ?", id).First(&invoice).Error
	return &invoice, err
}

func (s *CooperativeMemberService) GetMembers(request http.Request, search string) (paginate.Page, error) {
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
	stmt = stmt.Model(&models.CooperativeMemberModel{})
	utils.FixRequest(&request)
	page := pg.With(stmt).Request(request).Response(&[]models.CooperativeMemberModel{})
	page.Page = page.Page + 1
	return page, nil
}

func (s *CooperativeMemberService) InviteMember(data *models.MemberInvitationModel) (string, error) {
	if data.Token == "" {
		data.Token = utils.RandString(32, false)
	}
	err := s.db.Create(&data).Error
	if err != nil {
		return "", err
	}
	return data.Token, nil
}

func (s *CooperativeMemberService) GetInvitedMembers(request http.Request, search string) (paginate.Page, error) {
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

func (s *CooperativeMemberService) AcceptMemberInvitation(token string, userID string) error {
	var invitation models.MemberInvitationModel
	if err := s.db.Where("token = ?", token).First(&invitation).Error; err != nil {
		return err
	}
	data := models.CooperativeMemberModel{
		ConnectedTo: &userID,
		CompanyID:   invitation.CompanyID,
	}
	err := s.db.Create(&data).Error
	if err != nil {
		return err
	}

	err = s.db.Delete(&invitation).Error
	if err != nil {
		return err
	}

	// add role to user

	return nil
}

func (s *CooperativeMemberService) DeleteInvitation(id string) error {
	var invitation models.MemberInvitationModel
	if err := s.db.Where("id = ?", id).First(&invitation).Error; err != nil {
		return err
	}

	err := s.db.Delete(&invitation).Error
	if err != nil {
		return err
	}

	return nil
}
