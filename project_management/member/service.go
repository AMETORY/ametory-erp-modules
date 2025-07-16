package member

import (
	"net/http"

	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/shared/models"
	"github.com/AMETORY/ametory-erp-modules/utils"
	"github.com/morkid/paginate"
	"gorm.io/gorm"
)

// MemberService provides methods for creating, updating, deleting, and retrieving members.
//
// The service requires a Gorm database instance and an ERP context.
type MemberService struct {
	db  *gorm.DB
	ctx *context.ERPContext
}

// NewMemberService creates a new instance of MemberService.
func NewMemberService(ctx *context.ERPContext) *MemberService {
	return &MemberService{
		db:  ctx.DB,
		ctx: ctx,
	}
}

// CreateMember creates a new member in the database.
//
// It takes a pointer to a MemberModel as a parameter and returns an error. The function
// uses the gorm.DB connection to create a new record in the members table. If the
// operation fails, an error is returned.
func (s *MemberService) CreateMember(data *models.MemberModel) error {
	return s.db.Create(data).Error
}

// UpdateMember updates a member in the database.
//
// It takes a string id and a pointer to a MemberModel as parameters and returns an error.
// The function uses the gorm.DB connection to update a record in the members table.
func (s *MemberService) UpdateMember(id string, data *models.MemberModel) error {
	return s.db.Where("id = ?", id).Updates(data).Error
}

// DeleteMember deletes a member from the database.
//
// It takes a string id as parameter and returns an error. The function uses the gorm.DB
// connection to delete a record in the members table.
func (s *MemberService) DeleteMember(id string) error {
	return s.db.Where("id = ?", id).Delete(&models.MemberModel{}).Error
}

// GetMemberByID retrieves a member by its ID.
//
// It takes a string id as parameter and returns a pointer to a MemberModel and an error.
// The function uses GORM to retrieve the member data from the members table. If the
// operation fails, an error is returned.
func (s *MemberService) GetMemberByID(id string) (*models.MemberModel, error) {
	var invoice models.MemberModel
	err := s.db.Where("id = ?", id).First(&invoice).Error
	return &invoice, err
}

// GetMembers retrieves a paginated list of members from the database.
//
// It takes an HTTP request and a search string as parameters. The search string filters the
// members by full name, email, and phone number. If a company ID is present in the request
// header, the result is filtered by the company ID. The function uses pagination to manage
// the result set and includes any necessary request modifications using the utils.FixRequest
// utility.
//
// The function returns a paginated page of MemberModel and an error if the operation fails.
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

// InviteMember creates a new invitation for a member in the database.
//
// It takes a pointer to a MemberInvitationModel as a parameter and returns the
// generated token and an error. The function generates a random token if one is not
// provided in the passed model. It uses the gorm.DB connection to create a new record in
// the member_invitations table. If the operation fails, an error is returned.
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

// GetInvitedMembers retrieves a paginated list of invited members from the database.
//
// It takes an HTTP request and a search string as parameters. The search string filters the
// invited members by full name. If a company ID is present in the request header, the
// result is filtered by the company ID. The function uses pagination to manage the result
// set and includes any necessary request modifications using the utils.FixRequest utility.
//
// The function returns a paginated page of MemberInvitationModel and an error if the
// operation fails.
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

// AcceptMemberInvitation accepts a member invitation and adds the user to the company
// with the specified role.
//
// It takes a string token and a user ID as parameters and returns an error. The function
// first retrieves the invitation from the database and then creates a new member with the
// specified role. If the operation fails, an error is returned.
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

	return nil
}

// DeleteInvitation deletes an invitation from the database.
//
// It takes a string id as parameter and returns an error. The function uses the gorm.DB
// connection to delete a record in the member_invitations table. If the operation fails,
// an error is returned.
func (s *MemberService) DeleteInvitation(id string) error {
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
